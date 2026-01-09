package ui

import "fmt"

// TableColumn represents a table column configuration
type TableColumn struct {
	Header string
	Width  int
	Align  string // "left", "center", "right"
}

// TableRenderer renders data in table format
type TableRenderer struct {
	columns []TableColumn
	rows    [][]string
}

// NewTable creates a new table renderer
func NewTable() *TableRenderer {
	return &TableRenderer{}
}

// AddColumn adds a column to the table
func (t *TableRenderer) AddColumn(header string, width int, align string) *TableRenderer {
	if align == "" {
		align = "left"
	}
	t.columns = append(t.columns, TableColumn{
		Header: header,
		Width:  width,
		Align:  align,
	})
	return t
}

// AddRow adds a data row to the table
func (t *TableRenderer) AddRow(cells ...string) *TableRenderer {
	t.rows = append(t.rows, cells)
	return t
}

// Render outputs the formatted table
func (t *TableRenderer) Render() string {
	if len(t.columns) == 0 {
		return ""
	}

	var lines []string

	var headerCells []string
	for i, col := range t.columns {
		cell := t.formatCell(col.Header, col.Width, col.Align)
		if i == 0 {
			headerCells = append(headerCells, Header.Render(cell))
		} else {
			headerCells = append(headerCells, Bold.Render(cell))
		}
	}
	lines = append(lines, joinCells(headerCells))

	var sepCells []string
	for _, col := range t.columns {
		sep := ""
		for i := 0; i < col.Width; i++ {
			sep += "─"
		}
		sepCells = append(sepCells, Muted.Render(sep))
	}
	lines = append(lines, joinCells(sepCells))

	for _, row := range t.rows {
		var rowCells []string
		for i, col := range t.columns {
			var cellValue string
			if i < len(row) {
				cellValue = row[i]
			}
			cell := t.formatCell(cellValue, col.Width, col.Align)
			rowCells = append(rowCells, cell)
		}
		lines = append(lines, joinCells(rowCells))
	}

	return joinLines(lines)
}

// formatCell formats a cell with the specified width and alignment
func (t *TableRenderer) formatCell(text string, width int, align string) string {
	if len(text) > width {
		if width > 3 {
			return text[:width-3] + "..."
		}
		return text[:width]
	}

	padding := width - len(text)
	switch align {
	case "center":
		leftPad := padding / 2
		rightPad := padding - leftPad
		return fmt.Sprintf("%*s%s%*s", leftPad, "", text, rightPad, "")
	case "right":
		return fmt.Sprintf("%*s", width, text)
	case "left":
		fallthrough
	default:
		return fmt.Sprintf("%-*s", width, text)
	}
}

// joinCells joins table cells with proper spacing
func joinCells(cells []string) string {
	result := ""
	for i, cell := range cells {
		if i > 0 {
			result += "  "
		}
		result += cell
	}
	return result
}

// joinLines joins multiple lines with newlines
func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

// ProgressBar renders a progress indicator
type ProgressBar struct {
	current int
	total   int
	width   int
	filled  string
	empty   string
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int) *ProgressBar {
	return &ProgressBar{
		total:  total,
		width:  30,
		filled: "█",
		empty:  "░",
	}
}

// WithWidth sets the progress bar width
func (p *ProgressBar) WithWidth(width int) *ProgressBar {
	p.width = width
	return p
}

// WithChars sets the fill and empty characters
func (p *ProgressBar) WithChars(filled, empty string) *ProgressBar {
	p.filled = filled
	p.empty = empty
	return p
}

// SetProgress updates the current progress
func (p *ProgressBar) SetProgress(current int) *ProgressBar {
	p.current = current
	return p
}

// Render outputs the progress bar
func (p *ProgressBar) Render() string {
	if p.total == 0 {
		return ""
	}

	percentage := float64(p.current) / float64(p.total)
	filled := int(float64(p.width) * percentage)
	empty := p.width - filled

	bar := ""
	for range filled {
		bar += p.filled
	}
	for range empty {
		bar += p.empty
	}

	progress := WithColor(Muted, theme.Data).Render(bar)
	percent := fmt.Sprintf(" %d%% (%d/%d)", int(percentage*100), p.current, p.total)

	return progress + Muted.Render(percent)
}
