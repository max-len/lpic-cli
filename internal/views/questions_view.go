package views

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

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

	underlinedStartStyle := "[yellow::bu]" // bold + underline for current selection
	underlinedStopStyle := "[-:-:-:-]"
	correctAnswerStyle := "[:green]"
	incorrectAnswerStyle := "[:red]"
	viewExplaination := false

	if len(r.currentQuestion.Answers) == 0 && r.currentQuestion != nil && len(r.currentQuestion.GetAnsweredOptions()) > 0 {
		log.Println("get answered options")
		r.currentQuestion.Answers = r.currentQuestion.GetAnsweredOptions()
	}

	// helper: word-aware wrap preserving spaces between words
	wrap := func(text string, avail int) []string {
		if avail <= 4 { // too narrow; avoid panic
			return []string{text}
		}
		escaped := tview.Escape(text)
		words := strings.Fields(escaped)
		if len(words) == 0 {
			return []string{""}
		}
		lines := []string{}
		var current strings.Builder
		curLen := 0
		for i, w := range words {
			wl := utf8.RuneCountInString(w)
			sep := 0
			if current.Len() > 0 {
				sep = 1
			}
			if curLen+sep+wl <= avail {
				if sep == 1 {
					current.WriteByte(' ')
					curLen++
				}
				current.WriteString(w)
				curLen += wl
			} else {
				// flush current
				if current.Len() > 0 {
					lines = append(lines, current.String())
				}
				current.Reset()
				curLen = 0
				// word longer than avail: hard split
				if wl > avail {
					runes := []rune(w)
					start := 0
					for start < len(runes) {
						end := start + avail
						if end > len(runes) {
							end = len(runes)
						}
						lines = append(lines, string(runes[start:end]))
						start = end
					}
				} else {
					current.WriteString(w)
					curLen = wl
				}
			}
			if i == len(words)-1 && current.Len() > 0 {
				lines = append(lines, current.String())
			}
		}
		return lines
	}

	drawLine := func(content string, lineY int) {
		if lineY < height {
			tview.Print(screen, content, x, y+lineY, width, tview.AlignLeft, tcell.ColorYellow)
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
	for index, option := range r.currentQuestion.Answers {
		if visualLine >= height {
			break
		}
		answerStyle := "[white]"
		radioButton := radioButtonUnchecked
		textstyleStart := ""
		textstyleStop := ""
		prefixChar := " "
		if r.markerPosition == index {
			textstyleStart = underlinedStartStyle
			textstyleStop = underlinedStopStyle
			// Use a distinctive UTF-8 glyph instead of '-' to avoid confusion
			// when an answer itself starts with a hyphen.
			prefixChar = "»"
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
		// Build base prefix (radio + space + prefixChar + space)
		basePrefix := fmt.Sprintf("%s %s", radioButton, prefixChar)
		// available width for first line text after styling markers (styles don't consume visual width but we keep simple)
		textAvail := width - 3
		if textAvail < 5 {
			textAvail = width
		}
		wrapped := wrap(option.Text, textAvail)
		for i, seg := range wrapped {
			if visualLine >= height {
				break
			}
			if i == 0 {
				line := fmt.Sprintf(`%s%s%s%s%s`, basePrefix, answerStyle, textstyleStart, seg, textstyleStop)
				drawLine(line, visualLine)
			} else {
				contPrefix := strings.Repeat(" ", len(basePrefix))
				line := fmt.Sprintf(`%s%s`, contPrefix, seg)
				drawLine(line, visualLine)
			}
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
