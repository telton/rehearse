package cmds

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"

	"github.com/telton/rehearse/workflow"
)

var (
	// Styles for list command
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	workflowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	filenameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	countStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))

	listCmd = &cli.Command{
		Name:        "list",
		Aliases:     []string{"ls"},
		Usage:       "list all available workflows",
		Description: `List finds all workflows in the directory and prints them out.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dir",
				Aliases: []string{"d"},
				Usage:   "Directory to list workflows in",
				Value:   filepath.Join(".github", "workflows"),
				Validator: func(s string) error {
					if _, err := os.Stat(s); errors.Is(err, fs.ErrNotExist) {
						return fmt.Errorf("directory does not exist: %s", s)
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "format",
				Aliases: []string{"f"},
				Usage:   "The formatting style (text, json)",
				Value:   "text",
				Validator: func(s string) error {
					if s == "text" || s == "json" {
						return nil
					}
					return fmt.Errorf("unknown format value: %s", s)
				},
			},
			&cli.BoolFlag{
				Name:  "pretty-print",
				Usage: "Enable pretty-print formatting",
				Value: false,
			},
		},
		Action: func(ctx context.Context, c *cli.Command) error {
			dir := c.String("dir")
			outFmt := c.String("format")
			prettyPrint := c.Bool("pretty-print")

			type workflowFile struct {
				Filename     string `json:"filename"`
				Filepath     string `json:"filepath"`
				WorkflowName string `json:"workflow_name"`
			}

			entries, err := os.ReadDir(dir)
			if err != nil {
				return fmt.Errorf("read dir %s: %w", dir, err)
			}

			var files []*workflowFile
			for _, e := range entries {
				isYaml := strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml")
				if isYaml {
					fullPath := filepath.Join(dir, e.Name())

					wrkFlw, err := workflow.Parse(fullPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error parsing workflow: %v", err)
						continue
					}

					f := &workflowFile{
						Filename:     e.Name(),
						Filepath:     fullPath,
						WorkflowName: wrkFlw.Name,
					}

					files = append(files, f)
				}
			}

			switch outFmt {
			case "json":
				encoder := json.NewEncoder(os.Stdout)
				if prettyPrint {
					encoder.SetIndent("", "  ")
				}
				if err := encoder.Encode(files); err != nil {
					return fmt.Errorf("write json: %w", err)
				}
			case "text":
				fallthrough
			default:
				if prettyPrint {
					fmt.Println(headerStyle.Render("Available Workflows"))
					fmt.Println()
					for _, f := range files {
						fmt.Printf("• %s %s\n",
							filenameStyle.Render(f.Filename),
							workflowStyle.Render("→ "+f.WorkflowName))
					}
					fmt.Println()
					fmt.Printf("%s workflow(s) found\n", countStyle.Render(fmt.Sprintf("%d", len(files))))
				} else {
					for _, f := range files {
						fmt.Printf("%s: %s\n", f.Filename, f.WorkflowName)
					}
				}
			}

			return nil
		},
	}
)
