package workflow

// AnalysisResult holds the complete run analysis.
type AnalysisResult struct {
	WorkflowName string
	Trigger      string
	Context      *Context
	Jobs         []JobResult
}

// JobResult holds analysis for a single job.
type JobResult struct {
	Name       string
	RunsOn     string
	Needs      []string
	Condition  *ConditionResult
	WouldRun   bool
	SkipReason string
	Steps      []StepResult
}

// StepResult holds analysis for a single step.
type StepResult struct {
	Name      string
	Type      string // "run" or "action"
	Action    string
	Command   string
	Condition *ConditionResult
	WouldRun  bool
}

// ConditionResult holds an evaluated condition.
type ConditionResult struct {
	Expression string
	Value      bool
	Trace      string
}

// Analyzer performs analysis.
type Analyzer struct {
	workflow *Workflow
	ctx      *Context
	eval     *Evaluator
}

func NewAnalyzer(w *Workflow, ctx *Context) *Analyzer {
	return &Analyzer{
		workflow: w,
		ctx:      ctx,
		eval:     NewEvaluator(ctx),
	}
}

// Analyze performs analysis.
func (a *Analyzer) Analyze() *AnalysisResult {
	result := &AnalysisResult{
		WorkflowName: a.workflow.Name,
		Trigger:      a.ctx.GitHub.EventName,
		Context:      a.ctx,
	}

	// Get job execution order.
	order := a.topologicalSort()

	for _, jobName := range order {
		job := a.workflow.Jobs[jobName]
		jobResult := a.analyzeJob(jobName, job)
		result.Jobs = append(result.Jobs, jobResult)

		// Update context for dependent jobs.
		status := "success"
		if !jobResult.WouldRun {
			status = "skipped"
		}
		a.ctx.Jobs[jobName] = JobContext{Status: status}
	}

	return result
}

func (a *Analyzer) analyzeJob(name string, job Job) JobResult {
	result := JobResult{
		Name:   name,
		RunsOn: job.RunsOn.String(),
		Needs:  job.Needs.Jobs,
	}

	// Check if dependencies are satisfied.
	needsSatisfied := true
	for _, dep := range job.Needs.Jobs {
		if jobCtx, ok := a.ctx.Jobs[dep]; ok {
			if jobCtx.Status != "success" {
				needsSatisfied = false
				result.SkipReason = "dependency '" + dep + "' was skipped"
				break
			}
		}
	}

	// Evaludate job condition.
	if job.If != "" {
		condResult := a.evaluateCondition(job.If)
		result.Condition = condResult
		if !condResult.Value {
			result.WouldRun = false
			result.SkipReason = "condition evaluated to false"
		} else if needsSatisfied {
			result.WouldRun = true
		}
	} else {
		result.WouldRun = needsSatisfied
	}

	if !needsSatisfied && result.SkipReason == "" {
		result.SkipReason = "dependency not satisfied"
	}

	// Analyze steps.
	for _, step := range job.Steps {
		stepResult := a.analyzeStep(step)
		result.Steps = append(result.Steps, stepResult)
	}

	return result
}

func (a *Analyzer) analyzeStep(step Step) StepResult {
	result := StepResult{
		Name:    step.Name,
		Command: step.Run,
		Action:  step.Uses,
	}

	// Determine step name if not set.
	if result.Name == "" {
		if step.Run != "" {
			result.Name = truncate(step.Run, 40)
		} else if step.Uses != "" {
			result.Name = step.Uses
		}
	}

	// Determine type.
	if step.Uses != "" {
		result.Type = "action"
	} else {
		result.Type = "run"
	}

	// Evaluate condition.
	if step.If != "" {
		condResult := a.evaluateCondition(step.If)
		result.Condition = condResult
		result.WouldRun = condResult.Value
	} else {
		result.WouldRun = true
	}

	return result
}

func (a *Analyzer) evaluateCondition(expr string) *ConditionResult {
	result, err := a.eval.Evaluate(expr)
	if err != nil {
		return &ConditionResult{
			Expression: expr,
			Value:      false,
			Trace:      "error: " + err.Error(),
		}
	}

	boolVal := false
	if b, ok := result.Value.(bool); ok {
		boolVal = b
	}

	return &ConditionResult{
		Expression: expr,
		Value:      boolVal,
		Trace:      result.Trace,
	}
}

// topologicalSort returns jobs in dependency order.
func (a *Analyzer) topologicalSort() []string {
	visited := make(map[string]bool)
	order := []string{}

	var visit func(name string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true

		job := a.workflow.Jobs[name]
		for _, dep := range job.Needs.Jobs {
			visit(dep)
		}
		order = append(order, name)
	}

	for name := range a.workflow.Jobs {
		visit(name)
	}

	return order
}

func truncate(s string, max int) string {
	// Get first line.
	for i, c := range s {
		if c == '\n' {
			s = s[:i]
			break
		}
	}

	if len(s) > max {
		return s[:max-3] + "..."
	}

	return s
}
