package ui

import (
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/glazed/pkg/types"
	"io"
)

// TODO(2023-02-17, manuel) Turn this into a selector widget like gum
//
// This would be something similar to `gum choose`, but currently
// we would have to figure out how to display it to stderr yet output the
// chosen value (instead of standard OutputFormatter) to stdout.
//
// In a way, the output formatter is actually the UI here.

type SelectListModel struct {
	list     list.Model
	choice   interface{}
	quitting bool
}

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct {
}

func (i itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item_, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d %s", index+1, item_)
	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render("> " + s)
		}
	}

	fmt.Fprint(w, fn(str))
}

func (i itemDelegate) Height() int {
	return 1
}

func (i itemDelegate) Spacing() int {
	return 0
}

func (i itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}

func NewSelectListModelFromTable(table *types.Table, column string) *SelectListModel {
	var items []list.Item
	for _, row := range table.Rows {
		row_ := row.GetValues()
		item_ := fmt.Sprintf("%s", row_[column])

		items = append(items, item(item_))
	}

	const defaultWidth = 50
	const listHeight = 14

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.Title = titleStyle
	l.Styles.HelpStyle = helpStyle
	l.Title = "Select a " + column
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	return &SelectListModel{
		list: l,
	}
}

func (m *SelectListModel) Init() tea.Cmd {
	return nil
}

func (m *SelectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = i
				m.quitting = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *SelectListModel) View() string {
	if m.choice != "" {
		return fmt.Sprintf("%s", m.choice)
	}
	if m.quitting {
		return "Quitting wow"

	}
	return ""
}

func (m *SelectListModel) Run() (interface{}, error) {
	m_, err := tea.NewProgram(m).Run()
	if err != nil {
		return nil, err
	}

	return m_.(*SelectListModel).choice, nil
}
