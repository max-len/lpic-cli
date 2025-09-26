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

    // Define styles
    styleCorrect := tcell.StyleDefault.Background(tcell.ColorGreen)
    styleIncorrect := tcell.StyleDefault.Background(tcell.ColorRed)
    styleUnknown := tcell.StyleDefault.Background(tcell.ColorBlack)

    capacity := width * height
    bottomY := y + height - 1

    // Helper to map linear index to serpentine coordinates (snaking)
    toCoord := func(idx int) (int, int) {
        row := idx / width          // 0 = first (bottom) row logically
        col := idx % width
        // bottom row is visual bottom
        drawY := bottomY - row
        // serpentine: even row left->right, odd row right->left
        if row%2 == 1 {
            col = width - 1 - col
        }
        drawX := x + col
        return drawX, drawY
    }

    // Clear area first (unknown)
    for py := 0; py < height; py++ {
        for px := 0; px < width; px++ {
            screen.SetContent(x+px, y+py, ' ', nil, styleUnknown)
        }
    }

    if v.total <= capacity && len(v.states) == v.total {
        // Direct mapping: each question -> one cell
        for qi, st := range v.states {
            if qi >= capacity { break }
            dx, dy := toCoord(qi)
            style := styleUnknown
            switch st {
            case types.AnsweredTrue:
                style = styleCorrect
            case types.AnsweredFalse:
                style = styleIncorrect
            }
            screen.SetContent(dx, dy, ' ', nil, style)
        }
        return
    }

    // Half-block mode: up to 2 questions per cell using lower half block (first) then full block
    if v.total <= capacity*2 && len(v.states) == v.total {
        for qi, st := range v.states {
            if qi >= capacity*2 { break }
            cell := qi / 2
            half := qi % 2 // 0 first (lower half), 1 second (full block after painting)
            dx, dy := toCoord(cell)
            existingMainRune, existingComb, existingStyle, _ := screen.GetContent(dx, dy)
            _ = existingComb
            // Determine style for this question
            style := styleUnknown
            switch st {
            case types.AnsweredTrue:
                style = styleCorrect
            case types.AnsweredFalse:
                style = styleIncorrect
            }
            if half == 0 { // draw lower half block if answered
                if st == types.AnsweredTrue || st == types.AnsweredFalse {
                    screen.SetContent(dx, dy, '▄', nil, style)
                } // else leave unknown
            } else { // second half; merge result
                // Decompose existing style background
                _, existingBg, _ := existingStyle.Decompose()
                _, correctBg, _ := styleCorrect.Decompose()
                _, incorrectBg, _ := styleIncorrect.Decompose()
                secondIncorrect := st == types.AnsweredFalse
                secondCorrect := st == types.AnsweredTrue
                firstHalfIncorrect := existingMainRune == '▄' && existingBg == incorrectBg
                firstHalfCorrect := existingMainRune == '▄' && existingBg == correctBg
                firstAnsweredIncorrect := existingMainRune == '█' && existingBg == incorrectBg
                if secondIncorrect || firstHalfIncorrect || firstAnsweredIncorrect {
                    screen.SetContent(dx, dy, '█', nil, styleIncorrect)
                } else if secondCorrect && (firstHalfCorrect || (existingMainRune == '█' && existingBg == correctBg)) {
                    screen.SetContent(dx, dy, '█', nil, styleCorrect)
                } else if secondCorrect && existingMainRune == '▄' && existingBg != correctBg {
                    // show half correct (lower half already); keep as half block in correct color
                    screen.SetContent(dx, dy, '▄', nil, styleCorrect)
                } // else unanswered second half -> keep previous
            }
        }
        return
    }

    // Overflow aggregation mode: more than 2x capacity
    for cell := 0; cell < capacity; cell++ {
        start := cell * v.total / capacity
        end := (cell + 1) * v.total / capacity
        if end > v.total { end = v.total }
        if start >= end { continue }
        anyIncorrect := false
        anyCorrect := false
        unanswered := 0
        answered := 0
        for qi := start; qi < end; qi++ {
            switch v.states[qi] {
            case types.AnsweredTrue:
                anyCorrect = true
                answered++
            case types.AnsweredFalse:
                anyIncorrect = true
                answered++
            default:
                unanswered++
            }
            if anyIncorrect { // priority
                break
            }
        }
        style := styleUnknown
        if anyIncorrect {
            style = styleIncorrect
        } else if anyCorrect && unanswered == 0 {
            style = styleCorrect
        } else if anyCorrect && unanswered > 0 {
            // partial progress: dim correct
            style = styleCorrect.Dim(true)
        }
        dx, dy := toCoord(cell)
        screen.SetContent(dx, dy, ' ', nil, style)
    }
}
