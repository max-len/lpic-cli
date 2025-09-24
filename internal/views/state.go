package views

import (
	"fmt"

	"github.com/SqiSch/lpic-cli/internal/types"
	"github.com/rivo/tview"
)

// QuestionStateOverview shows a compact progress (current/total) plus aggregate statistics.
// The per-question colored list is intentionally omitted for brevity.
func QuestionStateOverview(certSet []*types.Question, tv *tview.TextView, currentIndex int) {
	total := len(certSet)
	if total == 0 {
		tv.SetText("0 / 0").SetDynamicColors(true)
		return
	}

	correct := 0
	incorrect := 0
	unknown := 0
	for _, q := range certSet {
		switch q.AnsweredState {
		case types.AnsweredTrue:
			correct++
		case types.AnsweredFalse:
			incorrect++
		default:
			unknown++
		}
	}

	answered := correct + incorrect
	pct := func(v int) float64 { return (float64(v) / float64(total)) * 100 }

	// currentIndex is zero-based; shown as one-based
	header := fmt.Sprintf("[yellow::b]%d / %d[-]", currentIndex+1, total)
	stats := fmt.Sprintf("Answered: %d (%.0f%%)\n[green]Correct[-]: %d ( %.0f%% )\n[red]Incorrect[-]: %d ( %.0f%% )\n[white]Unknown[-]: %d ( %.0f%% )",
		answered, pct(answered), correct, pct(correct), incorrect, pct(incorrect), unknown, pct(unknown))

	tv.SetText(header + "\n\n" + stats).SetDynamicColors(true)
}
