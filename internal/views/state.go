package views

import (
	"fmt"

	"github.com/SqiSch/lpic-cli/internal/types"
	"github.com/rivo/tview"
)

func QuestionStateOverview(certSet []*types.Question, tview *tview.TextView, currentIndex int) {
	index := 0
	questionCount := ""

	correct := 0
	incorrect := 0
	unknown := 0
	total := len(certSet)

	for _, question := range certSet {
		if index == currentIndex {
			questionCount = fmt.Sprintf("%s [yellow::b]%d[-] ", questionCount, index+1)
			index++
			continue
		}
		switch question.AnsweredState {
		case types.AnsweredUnknown:
			questionCount = fmt.Sprintf("%s %d ", questionCount, index+1)
			unknown++
		case types.AnsweredTrue:
			questionCount = fmt.Sprintf("%s [green]%d[-] ", questionCount, index+1)
			correct++
		case types.AnsweredFalse:
			questionCount = fmt.Sprintf("%s [red]%d[-] ", questionCount, index+1)
			incorrect++
		}
		index++
	}

	done := correct + incorrect
	questionCount = fmt.Sprintf("%s\n\n%d / %d - [red]%d[-] / [green]%d[-] ", questionCount, done, total, incorrect, correct)

	questionsLenght := len(certSet)
	CorrectAnswered := 0
	IncorrectAnswered := 0

	for _, question := range certSet {
		if question.AnsweredState == types.AnsweredTrue {
			CorrectAnswered++
		} else if question.AnsweredState == types.AnsweredFalse {
			IncorrectAnswered++
		}
	}

	precentageCorrect := float64(CorrectAnswered) / float64(questionsLenght) * 100
	precentageIncorrect := float64(IncorrectAnswered) / float64(questionsLenght) * 100
	percentageUnknown := float64(unknown) / float64(questionsLenght) * 100
	questionCount = fmt.Sprintf("%s\n\n[red]Incorrect[-]: %d ( %.0f%%)\n[green]Correct[-]: %d ( %.0f%%)\n[white]Unknown[-]: %d ( %.0f%%)\n\n", questionCount, incorrect, precentageIncorrect, correct, precentageCorrect, unknown, percentageUnknown)

	tview.SetText(questionCount).SetDynamicColors(true)
}
