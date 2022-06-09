package main

import (
	"io"
	"log"
	"fmt"
	"os"
	"golang.org/x/term"
	// "time"
	// "strings"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)


type model struct {
	cursor int
	stack *StackInfo
	stackList list.Model
}

// List item interface
func (i StackCall) FilterValue() string {
	return ""
}
type callListItemDelegate struct{}

func (d callListItemDelegate) Height() int {
	return 1
}
func (d callListItemDelegate) Spacing() int {
	return 0
}
func (d callListItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd {
	return nil
}
func (d callListItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(StackCall)
	if !ok {
		return
	}

	str := fmt.Sprintf("%4d  %s %s %5d:%d", index + 1, listFileNameStyle.Render(item.fileName), listEntryNameStyle.Render(item.entryName), item.line, item.column)

	var fn func(string) string

	if index == m.Index() {
		fn = selectedItemStyle.Render
	} else {
		fn = itemStyle.Render
	}

	fmt.Fprintf(w, fn(str))
}

var (
	titleStyle        = lipgloss.NewStyle().
												MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().
												PaddingLeft(2)
	selectedItemStyle = lipgloss.NewStyle().
												PaddingLeft(1).
												Border(lipgloss.Border{Left: "▋"}, false, false, false, true).
										    BorderForeground(lipgloss.Color("#22c2f2")).
										    BorderBackground(lipgloss.Color("#07171c")).
												Foreground(lipgloss.Color("#22c2f2")).
												Background(lipgloss.Color("#07171c"))
	listFileNameStyle = lipgloss.NewStyle().
												MaxWidth(30).
												Width(58)
	listEntryNameStyle = lipgloss.NewStyle().
												MaxWidth(30).
												Width(34)

	paginationStyle   = list.DefaultStyles().PaginationStyle.
												PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.
												PaddingLeft(4).
												PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().
												Margin(1, 0, 2, 4)
)


func startGUI(stack *StackInfo) {
	physicalWidth, physicalHeigth, _ := term.GetSize(int(os.Stdout.Fd()))

	listItems := make([]list.Item, len(stack.calls))
	for i, call := range stack.calls {
		listItems[i] = call //callListItem(fmt.Sprintf("%s -> %s [%d:%d]", call.fileName, call.entryName, call.line, call.column))
	}

	m := model{0, stack, list.New(listItems, callListItemDelegate{}, physicalWidth, physicalHeigth - 2)}

	m.stackList.Title = "Stack"
	m.stackList.SetShowStatusBar(false)
	m.stackList.SetFilteringEnabled(false)
	m.stackList.Styles.Title = titleStyle
	m.stackList.Styles.PaginationStyle = paginationStyle
	m.stackList.Styles.HelpStyle = helpStyle

	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		// case "up":
		// 	m.cursor--
		// 	if m.cursor < 0 {
		// 		m.cursor = 0
		// 	}
		// case "down":
		// 	m.cursor++
		// 	if m.cursor >= len(m.stack.calls) {
		// 		m.cursor = len(m.stack.calls) - 1
		// 	}
		}
	}

	var cmd tea.Cmd
	m.stackList, cmd = m.stackList.Update(message)
	return m, cmd
}

func (m model) View() string {
	physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))

	itemStyle = itemStyle.MaxWidth(physicalWidth - 2).Width(physicalWidth)
	selectedItemStyle = selectedItemStyle.MaxWidth(physicalWidth - 2).Width(physicalWidth)

	physicalWidth -= (26)
	enamew := int(physicalWidth / 3)
	fnamew := physicalWidth - enamew
	listFileNameStyle = listFileNameStyle.MaxWidth(fnamew)
	listEntryNameStyle = listEntryNameStyle.MaxWidth(enamew)

	return "\n" + m.stackList.View()
}
