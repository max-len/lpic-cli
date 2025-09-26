package views

import (
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
}

func NewVerticalProgressBar() *VerticalProgressBar { return &VerticalProgressBar{Box: tview.NewBox()} }

func (v *VerticalProgressBar) SetQuestions(questions []*types.Question) {
    total := len(questions)
    if total == 0 {
        v.total, v.correct, v.incorrect, v.unknown = 0, 0, 0, 0
        return
    }
    correct, incorrect, unknown := 0, 0, 0
    for _, q := range questions {
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
}

func (v *VerticalProgressBar) Draw(screen tcell.Screen) {
    v.Box.DrawForSubclass(screen, v)
    if v.total == 0 { return }
    x, y, width, height := v.GetInnerRect()
    if width <= 0 || height <= 0 { return }

    h := height
    correctHeight := v.correct * h / v.total
    incorrectHeight := v.incorrect * h / v.total
    used := correctHeight + incorrectHeight
    if used > h { used = h }
    unknownHeight := h - used

    bottomY := y + height - 1

    drawSegment := func(startY int, segHeight int, style tcell.Style) {
        for row := 0; row < segHeight; row++ {
            yy := startY - row
            if yy < y || yy >= y+height { continue }
            for cx := 0; cx < width; cx++ {
                screen.SetContent(x+cx, yy, ' ', nil, style)
            }
        }
    }

    greenStyle := tcell.StyleDefault.Background(tcell.ColorGreen)
    redStyle := tcell.StyleDefault.Background(tcell.ColorRed)
    whiteStyle := tcell.StyleDefault.Background(tcell.ColorWhite)

    drawSegment(bottomY, correctHeight, greenStyle)
    drawSegment(bottomY-correctHeight, incorrectHeight, redStyle)
    drawSegment(bottomY-correctHeight-incorrectHeight, unknownHeight, whiteStyle)
}
