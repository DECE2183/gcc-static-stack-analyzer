package main

import (
  // "os"
	"io"
	"log"
	"fmt"
  "strings"
	// "golang.org/x/term"
  "github.com/charmbracelet/bubbles/viewport"
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
	baseGraph *CodeGraphNode

	stackList list.Model
  codeViewport viewport.Model

	filterInput bool
	filterEnabled bool
	page modelPage

  width, height int
}

// List item interface
type callListItem struct{*CodeGraphNode}

func (i callListItem) Title() string {
  return i.FileName
}
func (i callListItem) Description() string {
  return i.EntryName
}
func (i callListItem) FilterValue() string {
	return i.NodeName
}

// List delegate interface
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
	item, ok := listItem.(callListItem)
	if !ok {
		return
	}

	line1 := fmt.Sprintf("%4d  %s     %5d:%d", index + 1, listFileNameStyle.Render(item.FileName), item.Line, item.Column)
	line2 := listEntryNameStyle.Render(fmt.Sprintf("-> %s ", item.EntryName)) + listMemUsageStyle.Render(fmt.Sprintf("%d B | %d B", item.MaxStackUsage, item.SelfStackUsage))

	var fn func(string) string

  // Conditions
	var (
		isSelected  = index == m.Index()
		emptyFilter = m.FilterState() == list.Filtering && m.FilterValue() == ""
	)

  if emptyFilter {
    fn = itemStyle.Render
  } else if isSelected && m.FilterState() != list.Filtering {
    fn = selectedItemStyle.Render
  } else {
    fn = itemStyle.Render
  }

	fmt.Fprintf(w, fn(lipgloss.JoinVertical(lipgloss.Top, line1, line2)))
}

var (
	titleStyle        = lipgloss.NewStyle().
												MarginLeft(0)
	itemStyle         = lipgloss.NewStyle().
												PaddingLeft(2).
												Border(lipgloss.NormalBorder(), false, false, true, false).
										    BorderForeground(lipgloss.Color("#4C4C4C"))
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
												Margin(1, 0, 1, 2)

  viewPortStyle = lipgloss.NewStyle().
                        MarginLeft(2)
  viewportHeaderStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().Border(b, false, true, false, true).Padding(0, 1, 0, 1)
	}()
	viewportFooterStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.Copy().Border(b, false, true, false, true).Padding(0, 1, 0, 1).Margin(0)
	}()
)


func startGUI(baseGraph *CodeGraphNode) {
	listItems := make([]list.Item, len(baseGraph.ChildNodes))
	for i, nodePtr := range baseGraph.ChildNodes {
		listItems[i] = callListItem{nodePtr}
	}

	m := model{
		baseGraph: baseGraph,
		stackList: list.New(listItems, callListItemDelegate{}, 512, 512),
	}

	m.stackList.Title = "Functions"
	m.stackList.SetShowStatusBar(true)
	m.stackList.SetFilteringEnabled(true)
	m.stackList.Styles.Title = titleStyle
	m.stackList.Styles.PaginationStyle = paginationStyle
	m.stackList.Styles.HelpStyle = helpStyle
	m.stackList.SetSize(256, 256)

  m.codeViewport = viewport.New(256, 256)
  m.codeViewport.HighPerformanceRendering = false
  m.codeViewport.YPosition = 5

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
  var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

  switch msg := message.(type) {
	case tea.WindowSizeMsg:
		m.resize(msg.Width, msg.Height)
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
          m.codeViewport.SetContent(m.stackList.SelectedItem().(callListItem).CodeBlock)
					return m, nil
				}
			}

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

  switch m.page {
  case pageMain:
    m.stackList, cmd = m.stackList.Update(message)
    return m, cmd
  case pageInfo:
    m.codeViewport, cmd = m.codeViewport.Update(message)
	  return m, cmd
  }

  cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// m.resize()

	switch m.page {
	case pageMain:
		return "\n" + m.stackList.View()
	case pageInfo:
		return m.renderInfoPage()
	case pageQuit:
		str := titleTextStyle.Render("Realy quit?")
		return str
	}

	return ""
}

func (m *model) renderInfoPage() string {
  var infoContent, line string

  selectedItem := m.stackList.SelectedItem().(callListItem)

  line = lipgloss.JoinVertical(lipgloss.Top, selectedItem.FileName, "-> " + selectedItem.EntryName)
  infoContent = titleTextStyle.Render(line) + "\n"

  viewportHeader := viewportHeaderStyle.Render("Code preview")
  line = strings.Repeat("─", m.codeViewport.Width - lipgloss.Width(viewportHeader))
  viewportHeader = lipgloss.JoinHorizontal(lipgloss.Center, viewportHeader, line)

	viewportFooter := viewportFooterStyle.Render(fmt.Sprintf("%3.f%%", m.codeViewport.ScrollPercent() * 100))
	line = strings.Repeat("─", m.codeViewport.Width - lipgloss.Width(viewportFooter))
	viewportFooter = lipgloss.JoinHorizontal(lipgloss.Center, line, viewportFooter)

  infoContent += viewportHeader + "\n\n"
  infoContent += viewPortStyle.Render(m.codeViewport.View()) + "\n\n"
  infoContent += viewportFooter

  return infoContent
}

func (m *model) resize(w, h int) {
  m.width, m.height = w, h
	// w, h, _ := term.GetSize(int(os.Stdout.Fd()))

	itemStyle = itemStyle.MaxWidth(w - 2).Width(w)
	selectedItemStyle = selectedItemStyle.MaxWidth(w - 2).Width(w)

	namew := w - 26
	listFileNameStyle = listFileNameStyle.MaxWidth(namew).Width(w * 2)
	listEntryNameStyle = listEntryNameStyle.MaxWidth(namew + 6).Width(w * 2)

	listMemUsageStyle.Width(14)

  m.codeViewport.Width = w - 2
  m.codeViewport.Height = h - 10

	m.stackList.SetSize(w, h - 2)
}
