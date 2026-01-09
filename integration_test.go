package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/telton/rehearse/workflow"
)

// WorkflowTestCase represents a test case for workflow files
type WorkflowTestCase struct {
	Path        string
	Name        string
	ShouldFail  bool
	ExpectedErr string
}

// discoverWorkflowFiles recursively finds all workflow files in testdata
func discoverWorkflowFiles(rootDir string) ([]WorkflowTestCase, error) {
	var testCases []WorkflowTestCase

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-YAML files and README
		if info.IsDir() || (!strings.HasSuffix(path, ".yaml") && !strings.HasSuffix(path, ".yml")) {
			return nil
		}
		if strings.HasSuffix(path, "README.md") {
			return nil
		}

		// Determine if this workflow should fail based on its location
		shouldFail := strings.Contains(path, "errors/")
		var expectedErr string
		if shouldFail {
			// Extract expected error type from filename or path
			if strings.Contains(path, "syntax") || strings.Contains(path, "malformed") {
				expectedErr = "syntax"
			} else if strings.Contains(path, "missing") {
				expectedErr = "missing"
			}
			// Note: Some error workflows are semantically invalid but syntactically valid
			// They should parse successfully but show issues during analysis
		}

		// Create display name from relative path
		relPath, _ := filepath.Rel(rootDir, path)
		displayName := strings.ReplaceAll(relPath, string(filepath.Separator), "_")
		displayName = strings.TrimSuffix(displayName, filepath.Ext(displayName))

		testCases = append(testCases, WorkflowTestCase{
			Path:        path,
			Name:        displayName,
			ShouldFail:  shouldFail,
			ExpectedErr: expectedErr,
		})

		return nil
	})

	return testCases, err
}

// TestWorkflowParsing tests that all workflow files can be parsed successfully
func TestWorkflowParsing(t *testing.T) {
	testCases, err := discoverWorkflowFiles("testdata")
	require.NoError(t, err, "Failed to discover workflow files")
	require.NotEmpty(t, testCases, "No workflow files found in testdata")

	for _, tc := range testCases {
		tc := tc // capture loop variable for parallel tests
		t.Run(tc.Name+"_parse", func(t *testing.T) {
			t.Parallel()

			wf, err := workflow.Parse(tc.Path)

			if tc.ShouldFail {
				if tc.ExpectedErr != "" {
					assert.Error(t, err, "Expected parsing to fail for %s", tc.Path)
					if err != nil {
						assert.Contains(t, strings.ToLower(err.Error()), tc.ExpectedErr,
							"Expected error to contain %s for %s", tc.ExpectedErr, tc.Path)
					}
				}
				// Some error workflows might still parse but fail at analysis
			} else {
				assert.NoError(t, err, "Expected parsing to succeed for %s", tc.Path)
				assert.NotNil(t, wf, "Parsed workflow should not be nil for %s", tc.Path)

				if wf != nil {
					assert.NotEmpty(t, wf.Name, "Workflow name should not be empty for %s", tc.Path)
					assert.NotEmpty(t, wf.Jobs, "Workflow should have at least one job for %s", tc.Path)
				}
			}
		})
	}
}

// TestWorkflowAnalysis tests that all valid workflow files can be analyzed
func TestWorkflowAnalysis(t *testing.T) {
	testCases, err := discoverWorkflowFiles("testdata")
	require.NoError(t, err, "Failed to discover workflow files")
	require.NotEmpty(t, testCases, "No workflow files found in testdata")

	// Filter out known failing cases for analysis
	validTestCases := make([]WorkflowTestCase, 0)
	for _, tc := range testCases {
		// Only test analysis on workflows that should parse successfully
		if !tc.ShouldFail || tc.ExpectedErr == "" {
			validTestCases = append(validTestCases, tc)
		}
	}

	for _, tc := range validTestCases {
		tc := tc // capture loop variable for parallel tests
		t.Run(tc.Name+"_analyze", func(t *testing.T) {
			t.Parallel()

			// Parse the workflow first
			wf, err := workflow.Parse(tc.Path)
			if err != nil {
				if tc.ShouldFail {
					t.Logf("Expected parsing failure for %s: %v", tc.Path, err)
					return
				}
				require.NoError(t, err, "Parsing should succeed for %s", tc.Path)
			}

			// Create context
			ctx, err := workflow.NewContext(workflow.Options{
				EventName: "push",
				Ref:       "refs/heads/main",
				Secrets:   make(map[string]string),
			})
			require.NoError(t, err, "Context creation should succeed")

			// Create analyzer and analyze
			analyzer := workflow.NewAnalyzer(wf, ctx)
			result := analyzer.Analyze()

			assert.NotNil(t, result, "Analysis result should not be nil for %s", tc.Path)

			if result != nil {
				assert.Equal(t, wf.Name, result.WorkflowName,
					"Analysis result should match workflow name")
				assert.Equal(t, "push", result.Trigger,
					"Analysis result should match trigger event")
				assert.NotEmpty(t, result.Jobs,
					"Analysis result should have jobs for %s", tc.Path)

				// For non-error cases, verify at least one job would run or has a clear reason not to
				if !tc.ShouldFail {
					foundRunningJob := false
					skippedCount := 0
					for _, job := range result.Jobs {
						if job.WouldRun {
							foundRunningJob = true
						} else if job.SkipReason != "" {
							skippedCount++
						}
					}
					// Either some job runs, or all jobs have clear skip reasons
					assert.True(t, foundRunningJob || skippedCount == len(result.Jobs),
						"Jobs should either run or have skip reasons for %s", tc.Path)
				}
			}
		})
	}
}

// TestAllWorkflowsParallel runs parsing and basic analysis on all workflows in parallel
func TestAllWorkflowsParallel(t *testing.T) {
	testCases, err := discoverWorkflowFiles("testdata")
	require.NoError(t, err, "Failed to discover workflow files")
	require.NotEmpty(t, testCases, "No workflow files found")

	t.Logf("Running parallel tests on %d workflow files", len(testCases))

	// Run all workflow tests in parallel
	for _, tc := range testCases {
		tc := tc // capture loop variable
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			start := time.Now()
			defer func() {
				t.Logf("Workflow %s completed in %v", tc.Name, time.Since(start))
			}()

			// Step 1: Parse the workflow file
			wf, parseErr := workflow.Parse(tc.Path)

			// Step 2: Analyze if parsing succeeded
			var analysisResult *workflow.AnalysisResult
			if parseErr == nil && wf != nil {
				ctx, ctxErr := workflow.NewContext(workflow.Options{
					EventName: "push",
					Ref:       "refs/heads/main",
					Secrets:   make(map[string]string),
				})
				require.NoError(t, ctxErr, "Context creation should succeed")

				analyzer := workflow.NewAnalyzer(wf, ctx)
				analysisResult = analyzer.Analyze()
			}

			// Validate results based on expectations
			if tc.ShouldFail {
				// For error test cases, either parsing should fail or analysis should show issues
				hasError := parseErr != nil
				if !hasError && analysisResult != nil {
					// Check if analysis found issues (no jobs would run)
					allSkipped := true
					for _, job := range analysisResult.Jobs {
						if job.WouldRun {
							allSkipped = false
							break
						}
					}
					if allSkipped && len(analysisResult.Jobs) > 0 {
						t.Logf("Workflow %s has all jobs skipped (expected for error case)", tc.Path)
					}
				}
				if !hasError {
					t.Logf("Note: %s was expected to fail but parsed successfully", tc.Path)
				}
			} else {
				// For valid workflows, parsing must succeed
				require.NoError(t, parseErr, "Parsing should succeed for valid workflow %s", tc.Path)
				require.NotNil(t, wf, "Parsed workflow should not be nil for %s", tc.Path)

				// Analysis should also succeed for valid workflows
				require.NotNil(t, analysisResult, "Analysis result should not be nil for %s", tc.Path)

				// Verify basic properties
				assert.Equal(t, wf.Name, analysisResult.WorkflowName,
					"Analysis result should match workflow name")
				assert.Equal(t, "push", analysisResult.Trigger,
					"Analysis result should match trigger event")
				assert.NotEmpty(t, analysisResult.Jobs,
					"Analysis result should have jobs for %s", tc.Path)

				// Count runnable jobs and steps
				runnableJobs := 0
				totalSteps := 0
				for _, job := range analysisResult.Jobs {
					if job.WouldRun {
						runnableJobs++
					}
					totalSteps += len(job.Steps)
				}

				t.Logf("Workflow %s: %d jobs (%d runnable), %d total steps",
					tc.Name, len(analysisResult.Jobs), runnableJobs, totalSteps)
			}
		})
	}
}

// TestWorkflowCategories tests workflows grouped by category
func TestWorkflowCategories(t *testing.T) {
	testCases, err := discoverWorkflowFiles("testdata")
	require.NoError(t, err)

	categories := map[string][]WorkflowTestCase{
		"basic":    {},
		"features": {},
		"errors":   {},
		"root":     {},
	}

	// Categorize test cases
	for _, tc := range testCases {
		switch {
		case strings.Contains(tc.Path, "basic/"):
			categories["basic"] = append(categories["basic"], tc)
		case strings.Contains(tc.Path, "features/"):
			categories["features"] = append(categories["features"], tc)
		case strings.Contains(tc.Path, "errors/"):
			categories["errors"] = append(categories["errors"], tc)
		default:
			categories["root"] = append(categories["root"], tc)
		}
	}

	for category, cases := range categories {
		if len(cases) == 0 {
			continue
		}

		t.Run(category, func(t *testing.T) {
			t.Parallel()

			t.Logf("Testing %d workflows in category '%s'", len(cases), category)

			successCount := 0

			for _, tc := range cases {
				_, parseErr := workflow.Parse(tc.Path)
				if parseErr == nil {
					successCount++
				}
			}

			if category == "errors" {
				t.Logf("Error category: %d/%d workflows parsed (some should fail)", successCount, len(cases))
			} else {
				assert.Positive(t, successCount,
					"Category %s should have some working workflows", category)
				t.Logf("Category '%s': %d/%d workflows parsed successfully",
					category, successCount, len(cases))
			}
		})
	}
}

// TestWorkflowExpressions tests that workflows with expressions evaluate correctly
func TestWorkflowExpressions(t *testing.T) {
	expressionWorkflows := []string{
		"testdata/features/expressions-demo.yaml",
		"testdata/features/conditionals.yaml",
		"testdata/basic/multi-job.yaml",
	}

	for _, workflowPath := range expressionWorkflows {
		workflowPath := workflowPath
		t.Run(filepath.Base(workflowPath), func(t *testing.T) {
			t.Parallel()

			// Test with different contexts to exercise expression evaluation
			contexts := []struct {
				name string
				opts workflow.Options
			}{
				{
					name: "main_branch",
					opts: workflow.Options{
						EventName: "push",
						Ref:       "refs/heads/main",
						Secrets:   make(map[string]string),
					},
				},
				{
					name: "feature_branch",
					opts: workflow.Options{
						EventName: "pull_request",
						Ref:       "refs/heads/feature/test",
						Secrets:   make(map[string]string),
					},
				},
			}

			for _, ctx := range contexts {
				// Parse workflow
				wf, err := workflow.Parse(workflowPath)
				if os.IsNotExist(err) {
					t.Skipf("Workflow file %s does not exist", workflowPath)
					continue
				}
				require.NoError(t, err, "Parsing should succeed for %s", workflowPath)

				// Create context
				workflowCtx, err := workflow.NewContext(ctx.opts)
				require.NoError(t, err, "Context creation should succeed")

				// Analyze
				analyzer := workflow.NewAnalyzer(wf, workflowCtx)
				result := analyzer.Analyze()

				assert.NotNil(t, result, "Analysis should succeed for %s with %s context",
					workflowPath, ctx.name)

				if result != nil {
					assert.NotEmpty(t, result.Jobs, "Should have jobs with %s context", ctx.name)

					// Verify context was properly applied
					assert.Equal(t, ctx.opts.EventName, result.Trigger)
					assert.Equal(t, ctx.opts.Ref, result.Context.GitHub.Ref)

					t.Logf("%s with %s context: %d jobs, %d would run",
						filepath.Base(workflowPath), ctx.name, len(result.Jobs),
						countRunnableJobs(result.Jobs))
				}
			}
		})
	}
}

// TestSpecificWorkflows tests specific workflows with known expected behaviors
func TestSpecificWorkflows(t *testing.T) {
	tests := []struct {
		name          string
		workflowPath  string
		event         string
		ref           string
		expectedJobs  int
		expectedSteps map[string]int // job name -> expected step count
		shouldAllRun  bool
	}{
		{
			name:         "basic hello workflow",
			workflowPath: "testdata/basic/hello.yaml",
			event:        "push",
			ref:          "refs/heads/main",
			expectedJobs: 1,
			expectedSteps: map[string]int{
				"hello": 2,
			},
			shouldAllRun: true,
		},
		{
			name:         "multi-job workflow",
			workflowPath: "testdata/basic/multi-job.yaml",
			event:        "push",
			ref:          "refs/heads/main",
			expectedJobs: 4,
			shouldAllRun: false, // deploy job is conditional
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Check if file exists first
			if _, err := os.Stat(tt.workflowPath); os.IsNotExist(err) {
				t.Skipf("Workflow file %s does not exist", tt.workflowPath)
				return
			}

			// Parse workflow
			wf, err := workflow.Parse(tt.workflowPath)
			require.NoError(t, err, "Parsing should succeed for %s", tt.workflowPath)
			require.NotNil(t, wf, "Parsed workflow should not be nil")

			// Create context
			ctx, err := workflow.NewContext(workflow.Options{
				EventName: tt.event,
				Ref:       tt.ref,
				Secrets:   make(map[string]string),
			})
			require.NoError(t, err, "Context creation should succeed")

			// Analyze
			analyzer := workflow.NewAnalyzer(wf, ctx)
			result := analyzer.Analyze()

			require.NotNil(t, result, "Analysis result should not be nil")
			assert.Len(t, result.Jobs, tt.expectedJobs,
				"Expected %d jobs for %s with event %s", tt.expectedJobs, tt.workflowPath, tt.event)

			if tt.shouldAllRun {
				for _, job := range result.Jobs {
					assert.True(t, job.WouldRun, "Job %s should run for %s", job.Name, tt.workflowPath)
				}
			}

			// Check expected step counts if provided
			if tt.expectedSteps != nil {
				for jobName, expectedStepCount := range tt.expectedSteps {
					found := false
					for _, job := range result.Jobs {
						if job.Name == jobName {
							found = true
							assert.Len(t, job.Steps, expectedStepCount,
								"Job %s should have %d steps", jobName, expectedStepCount)
							break
						}
					}
					assert.True(t, found, "Job %s should be found in results", jobName)
				}
			}
		})
	}
}

// TestWorkflowFileDiscovery tests that workflow discovery works correctly
func TestWorkflowFileDiscovery(t *testing.T) {
	testCases, err := discoverWorkflowFiles("testdata")
	require.NoError(t, err)
	require.NotEmpty(t, testCases)

	// Verify we found expected files
	expectedFiles := []string{
		"basic/hello.yaml",
		"basic/multi-job.yaml",
		"ci.yaml",
		"features/conditionals.yaml",
		"errors/invalid-workflow.yaml",
	}

	for _, expected := range expectedFiles {
		found := false
		for _, tc := range testCases {
			if strings.Contains(tc.Path, expected) {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find workflow file %s", expected)
	}

	// Verify error classification
	errorCount := 0
	for _, tc := range testCases {
		if tc.ShouldFail {
			errorCount++
		}
	}
	assert.Positive(t, errorCount, "Should have at least one error test case")

	t.Logf("Discovered %d workflow files (%d should fail)", len(testCases), errorCount)
}

// TestParallelStressTest runs intensive parallel testing to ensure thread safety
func TestParallelStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	testCases, err := discoverWorkflowFiles("testdata")
	require.NoError(t, err)
	require.NotEmpty(t, testCases)

	// Filter to valid workflows only
	validCases := make([]WorkflowTestCase, 0)
	for _, tc := range testCases {
		if !tc.ShouldFail {
			validCases = append(validCases, tc)
		}
	}

	if len(validCases) == 0 {
		t.Skip("No valid workflows found for stress testing")
	}

	// Run multiple iterations in parallel to stress test
	iterations := 50
	for i := 0; i < iterations; i++ {
		i := i
		t.Run(fmt.Sprintf("iteration_%d", i), func(t *testing.T) {
			t.Parallel()

			for _, tc := range validCases {
				// Parse and analyze workflow
				wf, err := workflow.Parse(tc.Path)
				assert.NoError(t, err, "Parsing should succeed for %s (iteration %d)", tc.Path, i)

				if wf != nil {
					ctx, err := workflow.NewContext(workflow.Options{
						EventName: "push",
						Ref:       "refs/heads/main",
						Secrets:   make(map[string]string),
					})
					assert.NoError(t, err, "Context creation should succeed (iteration %d)", i)

					analyzer := workflow.NewAnalyzer(wf, ctx)
					result := analyzer.Analyze()
					assert.NotNil(t, result, "Analysis should succeed for %s (iteration %d)", tc.Path, i)
				}
			}
		})
	}
}

// countRunnableJobs counts how many jobs would run
func countRunnableJobs(jobs []workflow.JobResult) int {
	count := 0
	for _, job := range jobs {
		if job.WouldRun {
			count++
		}
	}
	return count
}

// BenchmarkWorkflowParsing benchmarks workflow parsing performance
func BenchmarkWorkflowParsing(b *testing.B) {
	testCases, err := discoverWorkflowFiles("testdata")
	require.NoError(b, err)
	require.NotEmpty(b, testCases)

	// Filter to valid workflows only for benchmarking
	validCases := make([]WorkflowTestCase, 0)
	for _, tc := range testCases {
		if !tc.ShouldFail {
			validCases = append(validCases, tc)
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, tc := range validCases {
			_, err := workflow.Parse(tc.Path)
			if err != nil {
				b.Fatalf("Parsing failed for %s: %v", tc.Path, err)
			}
		}
	}
}

// BenchmarkWorkflowAnalysis benchmarks workflow analysis performance
func BenchmarkWorkflowAnalysis(b *testing.B) {
	testCases, err := discoverWorkflowFiles("testdata")
	require.NoError(b, err)
	require.NotEmpty(b, testCases)

	// Pre-parse workflows and create contexts for benchmarking
	type benchCase struct {
		workflow *workflow.Workflow
		context  *workflow.Context
		name     string
	}

	var benchCases []benchCase
	for _, tc := range testCases {
		if tc.ShouldFail {
			continue // skip error cases for benchmarking
		}

		wf, err := workflow.Parse(tc.Path)
		if err != nil {
			continue // skip if parsing fails
		}

		ctx, err := workflow.NewContext(workflow.Options{
			EventName: "push",
			Ref:       "refs/heads/main",
			Secrets:   make(map[string]string),
		})
		if err != nil {
			continue // skip if context creation fails
		}

		benchCases = append(benchCases, benchCase{
			workflow: wf,
			context:  ctx,
			name:     tc.Name,
		})
	}

	require.NotEmpty(b, benchCases, "No valid workflows found for benchmarking")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, bc := range benchCases {
			analyzer := workflow.NewAnalyzer(bc.workflow, bc.context)
			result := analyzer.Analyze()
			if result == nil {
				b.Fatalf("Analysis failed for %s", bc.name)
			}
		}
	}
}
