package views

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

type Model struct {
	content  string
	ready    bool
	viewport viewport.Model
}

func NewModel(content string) Model {
	return Model{
		content: content,
		ready:   false,
		viewport: viewport.Model{
			Width:  0,
			Height: 0,
		},
	}
}

var (
	modelTitleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	modelInfoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()
)

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" || k == "q" || k == "esc" {
			return m, tea.Quit
		}
		if k := msg.String(); k == tea.KeyDown.String() {

		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.HeaderView())
		footerHeight := lipgloss.Height(m.FooterView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			// Since this program is using the full size of the viewport we
			// need to wait until we've received the window dimensions before
			// we can initialize the viewport. The initial dimensions come in
			// quickly, though asynchronously, which is why we wait for them
			// here.
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.SetContent(m.content)
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}
	}

	// Handle keyboard and mouse events in the viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) MoveQuestionUo() {

}

func (m Model) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}

	//physicalWidth, _, _ := term.GetSize(int(os.Stdout.Fd()))
	doc := strings.Builder{}

	row := lipgloss.JoinHorizontal(
		lipgloss.Top,
		activeTab.Render("Questions"),
		tab.Render("Statistics"),
		tab.Render("Help"),
	)
	gap := tabGap.Render(strings.Repeat(" ", max(0, width-lipgloss.Width(row)-2)))
	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)
	doc.WriteString(row + "\n")

	const (
		historyA = "The Romans learned from the Greeks that quinces\n\n\n\n\n slowly cooked with honey would “set” when cool. The Apicius gives a recipe for preserving whole quinces, stems and leaves attached, in a bath of honey diluted with defrutum: Roman marmalade. Preserves of quince and lemon appear (along with rose, apple, plum and pear) in the Book of ceremonies of the Byzantine Emperor Constantine VII Porphyrogennetos."
	)

	questionSpace := lipgloss.JoinHorizontal(
		lipgloss.Top,
		historyStyle.Width(width).MaxWidth(width).Align(lipgloss.Center).AlignVertical(lipgloss.Left).AlignHorizontal(lipgloss.Left).Render(historyA),
	)

	// Color grid
	colors := func() string {
		b := strings.Builder{}

		var x int
		for x = 0; x < 200; x++ {
			s := lipgloss.NewStyle().SetString(fmt.Sprintf("%-5d", x))
			b.WriteString(s.String())
		}

		return b.String()
	}()

	// Dialog

	okButton := activeButtonStyle.Render("")
	cancelButton := buttonStyle.Render("Solve")

	questionRow := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF7DB")).
		//Background(lipgloss.Color("#F25D94")).
		MarginRight(2).
		Align(lipgloss.Left).
		PaddingRight(3)

	questionStyle := lipgloss.NewStyle().Width(width - 20).MaxWidth(width).Align(lipgloss.Left).PaddingBottom(1)

	// questionStyleCorrect := questionStyle.Foreground(lipgloss.Color("#1bd191")).Background(lipgloss.Color("#1bd191"))
	questionStyleMarked := questionStyle.
		Foreground(lipgloss.Color("#cfc61f")).
		Underline(true)

	markedOption := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#cfc61f")).
		Underline(true)

	marker1 := questionRow.Render(lipgloss.NewStyle().Render("✅"))
	marker2 := questionRow.Render(markedOption.Render("( O )"))
	marker3 := questionRow.Render(lipgloss.NewStyle().Render("❌"))
	marker4 := questionRow.Render(lipgloss.NewStyle().Render("[ X ]"))
	marker5 := questionRow.Render(lipgloss.NewStyle().Render("[   ]"))

	unmarkedQuestion := questionStyle.Render("Are you sure you want to eat marmalade?Are you sure you want to eat marmalade?Are you sure you want to eat marmalade?Are you sure you want to eat marmalade?Are you sure you want to eat marmalade? Are you sure you want to eat marmalade? \nTEST MULTILINE")
	markedQuestion := questionStyleMarked.Render("Are you sure you want to eat marmalade?Are you sure you want to eat marmalade?Are you sure you want to eat marmalade?Are you sure you want to eat marmalade?Are you sure you want to eat marmalade? Are you sure you want to eat marmalade? \nTEST MULTILINE")

	question1 := lipgloss.JoinHorizontal(lipgloss.Left, marker1, unmarkedQuestion)
	question2 := lipgloss.JoinHorizontal(lipgloss.Left, marker2, markedQuestion)
	question3 := lipgloss.JoinHorizontal(lipgloss.Left, marker3, unmarkedQuestion)
	question4 := lipgloss.JoinHorizontal(lipgloss.Left, marker4, unmarkedQuestion)
	question5 := lipgloss.JoinHorizontal(lipgloss.Left, marker5, unmarkedQuestion)

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, okButton, cancelButton)
	ui := lipgloss.JoinVertical(lipgloss.Center, question1, question2, question3, question4, question5, buttons)
	_ = ui

	//#####
	dialog := lipgloss.Place(width, 15,
		lipgloss.Center, lipgloss.Center,
		dialogBoxStyle.Render(ui),
		lipgloss.WithWhitespaceChars("猫咪"),
		lipgloss.WithWhitespaceForeground(subtle),
	)

	// Status bar
	w := lipgloss.Width
	statusKey := statusStyle.Render("27 / 80")
	encoding := encodingStyle.Render("70% = passed")
	fishCake := fishCakeStyle.Render("")
	statusVal := statusText.
		Width(width - w(statusKey) - w(encoding) - w(fishCake)).
		Render("Correct 15 (12%) | Incorrect 24 (70%) | Unanswered 10 (12%	)")

	bar := lipgloss.JoinHorizontal(lipgloss.Top,
		statusKey,
		statusVal,
		encoding,
		fishCake,
	)

	statusbar := statusBarStyle.Width(width + 30).Render(bar)

	sideView := lipgloss.NewStyle().Width(25).MaxWidth(25).PaddingLeft(1).Render(rainbow(lipgloss.NewStyle(), colors, blends))
	sideView = "\n" + sideView
	ss := lipgloss.JoinVertical(lipgloss.Left, questionSpace, dialog+"")
	view := lipgloss.JoinHorizontal(lipgloss.Top, ss, sideView+"\n\n")

	total := lipgloss.JoinVertical(lipgloss.Left, view, statusbar)

	//doc.WriteString(total)

	// if physicalWidth > 0 {
	// 	docStyle = docStyle.MaxWidth(physicalWidth)
	// }

	// Okay, let's print it
	//fmt.Println(docStyle.Render(doc.String()))

	return total
}

func (m Model) HeaderView() string {
	title := modelTitleStyle.Render("Mr. Pager")
	line := strings.Repeat("─", max2(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m Model) FooterView() string {
	info := modelInfoStyle.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max2(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

func max2(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func rainbow(base lipgloss.Style, s string, colors []color.Color) string {
	var str string
	for i, ss := range s {
		color, _ := colorful.MakeColor(colors[i%len(colors)])
		str = str + base.Foreground(lipgloss.Color(color.Hex())).Render(string(ss))
	}
	return str
}
