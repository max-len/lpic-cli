package views

import (
	"fmt"
	"log"
	"unicode/utf8"
	"strings"

	"github.com/SqiSch/lpic-cli/internal/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type QuestionsView struct {
	*tview.Box
	explainationView *tview.TextView
	questionTextView *tview.TextView
	markerPosition   int
	currentQuestion  *types.Question
}

func (r *QuestionsView) IncreaseMarkerPosition() {
	r.markerPosition++
	if r.markerPosition >= r.GetOptionsLenght() {
		r.markerPosition = r.GetOptionsLenght() - 1
	}
}

func (r *QuestionsView) DecreaseMarkerPosition() {
	r.markerPosition--
	if r.markerPosition < 0 {
		r.markerPosition = 0
	}
}

// NewRadioButtons returns a new radio button primitive.
func NewQuestionsView(answers []*types.Answer, questionTextView *tview.TextView, explainationView *tview.TextView) *QuestionsView {
	return &QuestionsView{
		Box:              tview.NewBox(),
		questionTextView: questionTextView,
		explainationView: explainationView,
		markerPosition:   0,
	}
}

func (r *QuestionsView) SetQuestion(question *types.Question) {
	r.markerPosition = 0
	r.explainationView.SetText("")
	r.currentQuestion = question
	escapedText := tview.Escape(question.Text)
	r.questionTextView.SetText(fmt.Sprintf("[white]%s[-]",escapedText))
}

func (r *QuestionsView) GetCurrentQuestion() *types.Question {
	return r.currentQuestion
}

func (r *QuestionsView) isMultiAnswer() bool {
	// check if the question has multiple answers
	multipleAnswers := false
	countCorrectAnswers := 0
	for _, option := range r.currentQuestion.Answers {
		if option.IsCorrect {
			countCorrectAnswers++
		}
		if countCorrectAnswers > 1 {
			multipleAnswers = true
			break
		}
	}
	return multipleAnswers
}

// Draw draws this primitive onto the screen.
func (r *QuestionsView) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()

	radioButtonUnchecked := "\u2610" // Unchecked.
	radioButtonChecked := "\u2611"   // Checked.
	if !r.isMultiAnswer() {
		radioButtonUnchecked = "\u25ef"
		radioButtonChecked = "\u25c9"
	}

	underlinedStartStyle := "[blue::b]" // bold only for current selection
	underlinedStopStyle := "[-:-:-:-]"
	correctAnswerStyle := "[:green]"
	incorrectAnswerStyle := "[:red]"
	viewExplaination := false

	if len(r.currentQuestion.Answers) == 0 && r.currentQuestion != nil && len(r.currentQuestion.GetAnsweredOptions()) > 0 {
		log.Println("get answered options")
		r.currentQuestion.Answers = r.currentQuestion.GetAnsweredOptions()
	}


	// helper to draw one line
	drawLine := func(content string, lineY int) {
			if lineY < height {
				// Fallback color white; selection styling injects [blue::bu] style tag for first line only.
				tview.Print(screen, content, x, y+lineY, width, tview.AlignLeft, tcell.ColorWhite)
			}
		}

	blackSpacerStyle := tcell.StyleDefault.Background(tcell.ColorBlack)
	clearLine := func(lineY int) {
		if lineY >= height { return }
		for cx := 0; cx < width; cx++ {
			screen.SetContent(x+cx, y+lineY, ' ', nil, blackSpacerStyle)
		}
	}

	visualLine := 0
	// word-wrapping helper (no style tags inside, purely visual based on runes)
	wrapText := func(text string, width int) []string {
		if width <= 0 {
			return []string{""}
		}
		words := strings.Fields(tview.Escape(text))
		if len(words) == 0 {
			return []string{""}
		}
		lines := []string{}
		var cur strings.Builder
		curLen := 0
		flush := func() {
			if cur.Len() > 0 {
				lines = append(lines, cur.String())
				cur.Reset()
				curLen = 0
			}
		}
		for _, w := range words {
			rlen := utf8.RuneCountInString(w)
			if curLen == 0 {
				// start new line
				if rlen <= width {
					cur.WriteString(w)
					curLen = rlen
				} else {
					// hard split long word
					runes := []rune(w)
					for len(runes) > 0 {
						space := width
						if space > len(runes) { space = len(runes) }
						lines = append(lines, string(runes[:space]))
						runes = runes[space:]
					}
				}
			continue
		}
		// need a space before next word
		if curLen + 1 + rlen <= width {
			cur.WriteByte(' ')
			cur.WriteString(w)
			curLen += 1 + rlen
			continue
		}
		// flush and start new line
		flush()
		if rlen <= width {
			cur.WriteString(w)
			curLen = rlen
		} else {
			// hard split
			runes := []rune(w)
			for len(runes) > 0 {
				space := width
				if space > len(runes) { space = len(runes) }
				lines = append(lines, string(runes[:space]))
				runes = runes[space:]
			}
		}
	}
	flush()
	return lines
	}

	for index, option := range r.currentQuestion.Answers {
		if visualLine >= height {
			break
		}
		answerStyle := "[white]"
		radioButton := radioButtonUnchecked
		textstyleStart := ""
		textstyleStop := ""
		// Constant-width prefix: (radio glyph) + space + marker slot + space
		markerChar := " "
		if r.markerPosition == index {
			textstyleStart = underlinedStartStyle
			textstyleStop = underlinedStopStyle
			markerChar = "»" // selection marker occupies same slot as blank when not selected
		}
		if r.isOptionMarked(index) {
			if option.IsCorrect {
				answerStyle = correctAnswerStyle
			} else {
				answerStyle = incorrectAnswerStyle
				viewExplaination = true
			}
			radioButton = radioButtonChecked
		}
		// Build base prefix (radio + space + marker + space) ensuring fixed width for all answers
		basePrefix := fmt.Sprintf("%s %s ", radioButton, markerChar)
		prefixVisualWidth := utf8.RuneCountInString(basePrefix) // all runes are single width here
		avail := width - prefixVisualWidth
		if avail < 5 { avail = width - prefixVisualWidth } // minimal safeguard
		segments := wrapText(option.Text, avail)
		contPrefix := strings.Repeat(" ", prefixVisualWidth)
		for i, seg := range segments {
			if visualLine >= height { break }
			prefix := basePrefix
			if i > 0 { prefix = contPrefix }
			styled := fmt.Sprintf("%s%s%s%s%s", prefix, answerStyle, textstyleStart, seg, textstyleStop)
			drawLine(styled, visualLine)
			visualLine++
		}
		// Add a blank spacer line between answers (except after last), if room left
		if index < len(r.currentQuestion.Answers)-1 && visualLine < height {
			clearLine(visualLine)
			visualLine++
		}
	}

	if viewExplaination {
		r.explainationView.SetText(fmt.Sprintf("[red]Wrong![-]\n%s", r.currentQuestion.Explanation))
	}
}

func (r *QuestionsView) indexToAnswerID(index int) string {
	if index < 0 || index >= len(r.currentQuestion.Answers) {
		return ""
	}
	return r.currentQuestion.Answers[index].AnswerID
}

// check if the current option is in markedOptions
func (r *QuestionsView) isOptionMarked(index int) bool {
	return r.currentQuestion.Answers[index].GetIsMarked()
}

func (r *QuestionsView) GetCurrentOptions() []*types.Answer {
	var answers []*types.Answer
	for _, v := range r.currentQuestion.Answers {
		if v.GetIsMarked() {
			answers = append(answers, v)
		}
	}
	return answers
}

func (r *QuestionsView) getOptionByIndex(index int) *types.Answer {
	if index < 0 || index >= len(r.currentQuestion.Answers) {
		return nil
	}
	return r.currentQuestion.Answers[index]
}

func (r *QuestionsView) ToggleCurrentMarkedOption() types.AnsweredState {
	return r.ToggleMarkedOption(r.markerPosition)
}

// toggle the current option and return true if the option is correct
func (r *QuestionsView) ToggleMarkedOption(index int) types.AnsweredState {
	if r.isOptionMarked(index) {
		r.removeCurrentOption(index)
	} else {
		r.currentQuestion.Answers[index].SetIsMarked(true)
	}
	r.markerPosition = index

	if r.currentQuestion.IsSingleAnswer() {
		for i, v := range r.currentQuestion.Answers {
			if i != index {
				v.SetIsMarked(false)
			}
		}
	}

	if r.checkAllCorrectMarked() {
		r.explainationView.SetText(fmt.Sprintf("[green]Correct![-]\n%s", r.currentQuestion.Explanation))
		return types.AnsweredTrue
	}

	isCorrect := r.getOptionByIndex(index).IsCorrect
	if !isCorrect && r.isOptionMarked(index) {
		r.explainationView.SetText(fmt.Sprintf("[red]Wrong![-]\n%s", r.currentQuestion.Text))
		return types.AnsweredFalse
	}
	return types.AnsweredUnknown
}

func (r *QuestionsView) checkAllCorrectMarked() bool {
	for _, answer := range r.currentQuestion.Answers {
		if answer.IsCorrect && !answer.GetIsMarked() {
			return false
		} else if !answer.IsCorrect && answer.GetIsMarked() {
			return false
		}
	}
	return true
}

// MouseHandler returns the mouse handler for this primitive.
func (r *QuestionsView) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return r.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		x, y := event.Position()
		_, rectY, _, _ := r.GetInnerRect()
		if !r.InRect(x, y) {
			return false, nil
		}

		if action == tview.MouseLeftClick {
			setFocus(r)
			index := y - rectY
			if index >= 0 && index < len(r.currentQuestion.Answers) {
				r.ToggleMarkedOption(index)
				consumed = true
			}
		}

		return
	})
}

// remove current option from markedOptions
func (r *QuestionsView) removeCurrentOption(index int) {
	for i, v := range r.currentQuestion.Answers {
		if v.AnswerID == r.indexToAnswerID(index) {
			r.currentQuestion.Answers[i].SetIsMarked(false)
		}
	}
}

func (r *QuestionsView) ShowExplanation() {
	if r.currentQuestion.Explanation != "" {
		r.explainationView.SetText(r.currentQuestion.Explanation)
	} else {
		r.explainationView.SetText("[red]No explanation available[-]")
	}
	log.Println("Show explanation")
}

func (r *QuestionsView) GetOptionsLenght() int {
	if r.currentQuestion == nil {
		return 0
	}
	return len(r.currentQuestion.Answers)
}
