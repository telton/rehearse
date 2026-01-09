package ui

import (
	"strings"
)

// Component represents a reusable UI component
type Component interface {
	Render() string
}

// HeaderComponent renders styled headers
type HeaderComponent struct {
	Text   string
	Emoji  string
	Margin bool
}

// NewHeader creates a new header component
func NewHeader(text string) *HeaderComponent {
	return &HeaderComponent{Text: text, Margin: false}
}

// WithEmoji adds an emoji prefix to the header
func (h *HeaderComponent) WithEmoji(emoji string) *HeaderComponent {
	h.Emoji = emoji
	return h
}

// WithMargin adds bottom margin to the header
func (h *HeaderComponent) WithMargin() *HeaderComponent {
	h.Margin = true
	return h
}

// Render outputs the styled header
func (h *HeaderComponent) Render() string {
	text := h.Text
	if h.Emoji != "" {
		text = h.Emoji + " " + text
	}

	style := Header
	if h.Margin {
		style = style.MarginBottom(1)
	}

	return style.Render(text)
}

// LabelValueComponent renders label: value pairs
type LabelValueComponent struct {
	Label  string
	Value  string
	Indent int
}

// NewLabelValue creates a new label-value component
func NewLabelValue(label, value string) *LabelValueComponent {
	return &LabelValueComponent{Label: label, Value: value}
}

// WithIndent adds left indentation
func (lv *LabelValueComponent) WithIndent(spaces int) *LabelValueComponent {
	lv.Indent = spaces
	return lv
}

// Render outputs the styled label-value pair
func (lv *LabelValueComponent) Render() string {
	labelText := Label.Render(lv.Label)
	valueText := Value.Render(lv.Value)
	result := labelText + " " + valueText

	if lv.Indent > 0 {
		indent := strings.Repeat(" ", lv.Indent)
		result = indent + result
	}

	return result
}

// StatusComponent renders status indicators with appropriate colors
type StatusComponent struct {
	Status string
	Text   string
	Icon   string
}

// NewStatus creates a new status component
func NewStatus(status, text string) *StatusComponent {
	return &StatusComponent{Status: status, Text: text}
}

// WithIcon adds an icon to the status
func (s *StatusComponent) WithIcon(icon string) *StatusComponent {
	s.Icon = icon
	return s
}

// Render outputs the styled status
func (s *StatusComponent) Render() string {
	text := s.Text
	if s.Icon != "" {
		text = s.Icon + " " + text
	}

	return StatusColor(s.Status).Render(text)
}

// ListComponent renders bulleted or numbered lists
type ListComponent struct {
	Items  []string
	Bullet string
	Indent int
}

// NewList creates a new list component
func NewList(items []string) *ListComponent {
	return &ListComponent{Items: items, Bullet: "•", Indent: 2}
}

// WithBullet sets a custom bullet character
func (l *ListComponent) WithBullet(bullet string) *ListComponent {
	l.Bullet = bullet
	return l
}

// WithIndent sets the indentation level
func (l *ListComponent) WithIndent(spaces int) *ListComponent {
	l.Indent = spaces
	return l
}

// Render outputs the styled list
func (l *ListComponent) Render() string {
	var lines []string
	indent := strings.Repeat(" ", l.Indent)

	for _, item := range l.Items {
		lines = append(lines, indent+Muted.Render(l.Bullet)+" "+item)
	}

	return strings.Join(lines, "\n")
}

// SeparatorComponent renders visual separators
type SeparatorComponent struct {
	Length int
	Char   string
	Color  string
}

// NewSeparator creates a new separator
func NewSeparator() *SeparatorComponent {
	return &SeparatorComponent{Length: 40, Char: "─", Color: "muted"}
}

// WithLength sets separator length
func (s *SeparatorComponent) WithLength(length int) *SeparatorComponent {
	s.Length = length
	return s
}

// WithChar sets separator character
func (s *SeparatorComponent) WithChar(char string) *SeparatorComponent {
	s.Char = char
	return s
}

// Render outputs the separator
func (s *SeparatorComponent) Render() string {
	line := strings.Repeat(s.Char, s.Length)
	return Muted.Render(line)
}

// BoxComponent renders content in a bordered box
type BoxComponent struct {
	Content string
	Title   string
	Width   int
}

// NewBox creates a new box component
func NewBox(content string) *BoxComponent {
	return &BoxComponent{Content: content}
}

// WithTitle adds a title to the box
func (b *BoxComponent) WithTitle(title string) *BoxComponent {
	b.Title = title
	return b
}

// WithWidth sets the box width
func (b *BoxComponent) WithWidth(width int) *BoxComponent {
	b.Width = width
	return b
}

// Render outputs the bordered box
func (b *BoxComponent) Render() string {
	style := Box
	if b.Width > 0 {
		style = style.Width(b.Width)
	}

	content := b.Content
	if b.Title != "" {
		titleLine := Bold.Bold(true).Foreground(theme.Emphasis).Render(b.Title)
		content = titleLine + "\n" + content
	}

	return style.Render(content)
}
