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

type modelPage uint
const (
	pageMain modelPage = 0
	pageInfo modelPage = 1
	pageQuit modelPage = 2
)

type model struct {
	stack *StackInfo
	stackList list.Model
	filterInput bool
	filterEnabled bool
	page modelPage
}

// List item interface
func (i StackCall) FilterValue() string {
	return i.entryName
}
type callListItemDelegate struct{}

func (d callListItemDelegate) Height() int {
	return 3
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

	line1 := fmt.Sprintf("%4d  %s %5d:%d", index + 1, listFileNameStyle.Render(item.fileName), item.line, item.column)
	line2 := listEntryNameStyle.Render(fmt.Sprintf("-> %s ", item.entryName)) + listMemUsageStyle.Render(fmt.Sprintf("9%d B", item.memUsage))

	var fn func(string) string

	if index == m.Index() {
		fn = selectedItemStyle.Render
	} else {
		fn = itemStyle.Render
	}

	fmt.Fprintf(w, fn(lipgloss.JoinVertical(lipgloss.Top, line1, line2)))
}

var (
	titleStyle        = lipgloss.NewStyle().
												MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().
												PaddingLeft(2).
												Border(lipgloss.NormalBorder(), false, false, true, false).
										    BorderForeground(lipgloss.Color("#3C3C3C"))
	selectedItemStyle = lipgloss.NewStyle().
												PaddingLeft(2).
												Border(lipgloss.ThickBorder(), false, false, true, false).
										    BorderForeground(lipgloss.Color("#22c2f2")).
												Foreground(lipgloss.Color("#22c2f2"))
	listFileNameStyle = lipgloss.NewStyle()
	listEntryNameStyle = lipgloss.NewStyle().
												MarginLeft(6)
	listMemUsageStyle = lipgloss.NewStyle().
												Align(lipgloss.Right)

	paginationStyle   = list.DefaultStyles().PaginationStyle.
												PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.
												PaddingLeft(4).
												PaddingBottom(1)
	titleTextStyle    = lipgloss.NewStyle().
												Margin(1, 0, 2, 4)
)


func startGUI(stack *StackInfo) {
	// physicalWidth, physicalHeigth, _ := term.GetSize(int(os.Stdout.Fd()))

	listItems := make([]list.Item, len(stack.calls))
	for i, call := range stack.calls {
		listItems[i] = call //callListItem(fmt.Sprintf("%s -> %s [%d:%d]", call.fileName, call.entryName, call.line, call.column))
	}

	m := model{
		stack: stack,
		stackList: list.New(listItems, callListItemDelegate{}, 512, 512),
	}

	m.stackList.Title = "Stack"
	// m.stackList.SetShowStatusBar(true)
	// m.stackList.SetFilteringEnabled(true)
	m.stackList.Styles.Title = titleStyle
	m.stackList.Styles.PaginationStyle = paginationStyle
	m.stackList.Styles.HelpStyle = helpStyle
	m.stackList.SetSize(256, 256)

	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {

	case tea.WindowSizeMsg:
		// fmt.Printf("W: %d H: %d", msg.Width, msg.Height)
		m.resize()
		return m, nil

	case tea.KeyMsg:
		keypress := msg.String()
		switch m.page {

		case pageMain:

			if m.filterInput {
				switch keypress {
				case "esc":
					m.filterInput = false
				case"enter":
					m.filterInput = false
					m.filterEnabled = true
				}
			} else {
				switch keypress {
				case "/":
					m.filterInput = true
				case "esc":
					if m.filterEnabled {
						m.filterEnabled = false
					} else {
						m.page = pageQuit
						return m, tea.Quit
					}
				case "q", "ctrl+c":
					m.page = pageQuit
					return m, tea.Quit
				case "enter":
					m.page = pageInfo
					return m, nil
				}
			}

			var cmd tea.Cmd
			m.stackList, cmd = m.stackList.Update(message)
			return m, cmd

		case pageInfo:
			switch keypress {
			case "ctrl+c":
				m.page = pageQuit
				return m, nil
			case "q", "esc", "enter":
				m.page = pageMain
				return m, nil
			}
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	// physicalWidth, physicalHeigth, _ := term.GetSize(int(os.Stdout.Fd()))
	// m.resize(physicalWidth, physicalHeigth)
	// m.stackList.SetWidth(physicalWidth)
	// m.stackList.SetHeight(physicalHeigth - 2)

	m.resize()

	switch m.page {
	case pageMain:
		return "\n" + m.stackList.View()
	case pageInfo:
		str := titleTextStyle.Render("Info.")
		str += titleTextStyle.Render(m.stackList.SelectedItem().(StackCall).entryName)
		return str
	case pageQuit:
		str := titleTextStyle.Render("Realy quit?")
		return str
	}

	return ""
}

func (m *model) resize() {
	w, h, _ := term.GetSize(int(os.Stdout.Fd()))

	itemStyle = itemStyle.MaxWidth(w - 2).Width(w)
	selectedItemStyle = selectedItemStyle.MaxWidth(w - 2).Width(w)

	namew := w - 26
	listFileNameStyle = listFileNameStyle.MaxWidth(namew).Width(namew + 2)
	listEntryNameStyle = listEntryNameStyle.MaxWidth(namew + 6).Width(namew + 8)

	listMemUsageStyle.Width(8)

	m.stackList.SetSize(w, h - 2)
}
