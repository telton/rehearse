// Package ui provides reusable styled components for rehearse CLI commands.
//
// The ui package offers a consistent design system using charmbracelet/lipgloss
// for terminal styling. It includes:
//
//   - Centralized color palette and theming
//   - Reusable components (headers, status, lists, boxes)
//   - Specialized workflow renderers
//   - Table and progress bar utilities
//
// Usage:
//
//	// Basic components
//	header := ui.NewHeader("My Title").WithEmoji("🎭").WithMargin()
//	fmt.Println(header.Render())
//
//	status := ui.NewStatus("success", "Operation completed").WithIcon("✓")
//	fmt.Println(status.Render())
//
//	// Workflow rendering
//	renderer := ui.NewWorkflowRenderer()
//	fmt.Println(renderer.RenderJobHeader("build", "Build Application"))
//
//	// Tables
//	table := ui.NewTable().
//		AddColumn("Name", 20, "left").
//		AddColumn("Status", 10, "center").
//		AddRow("job1", "success").
//		AddRow("job2", "failed")
//	fmt.Println(table.Render())
//
// The package follows semantic color usage:
//   - Green: Success, completed operations
//   - Red: Errors, failures
//   - Yellow: Warnings, skipped items
//   - Blue: Information, docker operations
//   - Purple: Headers, emphasis
//   - Cyan: Data values
//   - Gray: Muted text, secondary information
//   - Pink: Special syntax (expressions, code)
//   - Orange: Environment variables, configuration
package ui
