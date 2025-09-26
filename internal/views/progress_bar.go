package views

import (
    "fmt"

    "github.com/SqiSch/lpic-cli/internal/types"
    "github.com/gdamore/tcell/v2"
    "github.com/rivo/tview"
)

// VerticalProgressBar shows overall session progress (bottom-up growth of answered states).
// Bottom -> top: Green (correct), Red (incorrect), White (unanswered).
type VerticalProgressBar struct {
    *tview.Box
    total     int
    correct   int
    incorrect int
    unknown   int
    states    []types.AnsweredState // per-question state for high-res mode
}

func NewVerticalProgressBar() *VerticalProgressBar { return &VerticalProgressBar{Box: tview.NewBox()} }

func (v *VerticalProgressBar) SetQuestions(questions []*types.Question) {
    total := len(questions)
    if total == 0 {
        v.total, v.correct, v.incorrect, v.unknown = 0, 0, 0, 0
        v.states = nil
        return
    }
    correct, incorrect, unknown := 0, 0, 0
    states := make([]types.AnsweredState, 0, total)
    for _, q := range questions {
        states = append(states, q.AnsweredState)
        switch q.AnsweredState {
        case types.AnsweredTrue:
            correct++
        case types.AnsweredFalse:
            incorrect++
        default:
            unknown++
        }
    }
    v.total, v.correct, v.incorrect, v.unknown = total, correct, incorrect, unknown
    v.states = states
}

func (v *VerticalProgressBar) Draw(screen tcell.Screen) {
    v.Box.DrawForSubclass(screen, v)
    if v.total == 0 { return }
    x, y, width, height := v.GetInnerRect()
    if width <= 0 || height <= 0 { return }

    styleCorrect := tcell.StyleDefault.Background(tcell.ColorGreen)
    styleIncorrect := tcell.StyleDefault.Background(tcell.ColorRed)
    styleUnknown := tcell.StyleDefault.Background(tcell.ColorBlack)
    stylePartial := styleCorrect.Dim(true)

    capacity := width * height
    bottomY := y + height - 1

    // Clear area to unknown
    for py := 0; py < height; py++ {
        for px := 0; px < width; px++ {
            screen.SetContent(x+px, y+py, ' ', nil, styleUnknown)
        }
    }

    if v.total <= capacity && len(v.states) == v.total {
        for q := 0; q < v.total; q++ {
            row := q / width
            col := q % width
            dy := bottomY - row
            dx := x + col
            style := styleUnknown
            switch v.states[q] {
            case types.AnsweredTrue:
                style = styleCorrect
            case types.AnsweredFalse:
                style = styleIncorrect
            }
            screen.SetContent(dx, dy, ' ', nil, style)
        }
        return
    }

    // Aggregated mode: map contiguous question ranges to each cell.
    for cell := 0; cell < capacity; cell++ {
        start := cell * v.total / capacity
        end := (cell + 1) * v.total / capacity
        if end > v.total { end = v.total }
        if start >= end { continue }
        anyIncorrect := false
        anyCorrect := false
        anyUnknown := false
        for qi := start; qi < end; qi++ {
            switch v.states[qi] {
            case types.AnsweredTrue:
                anyCorrect = true
            case types.AnsweredFalse:
                anyIncorrect = true
            default:
                anyUnknown = true
            }
            if anyIncorrect { break }
        }
        style := styleUnknown
        if anyIncorrect {
            style = styleIncorrect
        } else if anyCorrect && anyUnknown {
            style = stylePartial
        } else if anyCorrect {
            style = styleCorrect
        }
        row := cell / width
        col := cell % width
        dy := bottomY - row
        dx := x + col
        screen.SetContent(dx, dy, ' ', nil, style)
    }

    // Debug overlay: show capacity (number of character cells) inside bar (top-left).
    // This updates automatically on resize because Draw is called again.
    capStr := fmt.Sprintf("%d", capacity)
    if height > 0 && width >= len(capStr) {
        debugY := y // top row
        for i, r := range capStr {
            screen.SetContent(x+i, debugY, r, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
        }
    }
}
