package views

import (
	"fmt"
	"log"

	"github.com/SqiSch/lpic-cli/internal/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type QuestionsViewModern struct {
	*tview.Box
	explainationView *tview.TextView
	questionTextView *tview.TextView
	markerPosition   int
	currentQuestion  *types.Question
}

func (r *QuestionsViewModern) IncreaseMarkerPosition() {
	r.markerPosition++
	if r.markerPosition >= r.GetOptionsLenght() {
		r.markerPosition = r.GetOptionsLenght() - 1
	}
}

func (r *QuestionsViewModern) DecreaseMarkerPosition() {
	r.markerPosition--
	if r.markerPosition < 0 {
		r.markerPosition = 0
	}
}

// NewRadioButtons returns a new radio button primitive.
func NewQuestionsViewModern(answers []*types.Answer, questionTextView *tview.TextView, explainationView *tview.TextView) *QuestionsViewModern {
	return &QuestionsViewModern{
		Box:              tview.NewBox(),
		questionTextView: questionTextView,
		explainationView: explainationView,
		markerPosition:   0,
	}
}

func (r *QuestionsViewModern) SetQuestion(question *types.Question) {
	r.markerPosition = 0
	r.explainationView.SetText("")
	r.currentQuestion = question
	r.questionTextView.SetText(fmt.Sprintf("[white]%s[-]", question.Text))
}

func (r *QuestionsViewModern) GetCurrentQuestion() *types.Question {
	return r.currentQuestion
}

func (r *QuestionsViewModern) isMultiAnswer() bool {
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
func (r *QuestionsViewModern) Draw(screen tcell.Screen) {
	r.Box.DrawForSubclass(screen, r)
	x, y, width, height := r.GetInnerRect()

	// if multipleAnswers > 1 use check boxes, when not use radio buttons
	radioButtonUnchecked := "\u2610" // Unchecked.
	radioButtonChecked := "\u2611"   // Checked.

	if !r.isMultiAnswer() {
		radioButtonUnchecked = "\u25ef" // Unchecked.
		radioButtonChecked = "\u25c9"   // Checked.
	}

	underlinedStartStyle := "[yellow::u]"
	underlinedStopStyle := "[-:-:-:-]"

	correctAnswerStyle := "[:green]"
	incorrectAnswerStyle := "[:red]"

	viewExplaination := false

	if len(r.currentQuestion.Answers) == 0 && r.currentQuestion != nil && len(r.currentQuestion.GetAnsweredOptions()) > 0 {
		log.Println("get answered options")
		r.currentQuestion.Answers = r.currentQuestion.GetAnsweredOptions()
	}

	for index, option := range r.currentQuestion.Answers {
		if index >= height {
			break
		}
		answerStyle := "[white]"
		radioButton := radioButtonUnchecked

		textstyleStart := ""
		textstyleStop := ""
		prefix := " "

		if r.markerPosition == index {
			log.Println("marker position is", r.markerPosition, "index is", index)
			textstyleStart = underlinedStartStyle
			textstyleStop = underlinedStopStyle
			prefix = "-"
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

		line := fmt.Sprintf(`%s %s%s%s%s%s`, radioButton, prefix, answerStyle, textstyleStart, option.Text, textstyleStop)
		tview.Print(screen, line, x, y+index, width, tview.AlignLeft, tcell.ColorYellow)
	}

	if viewExplaination {
		r.explainationView.SetText(fmt.Sprintf("[red]Wrong![-]\n%s", r.currentQuestion.Explanation))
	}
}

func (r *QuestionsViewModern) indexToAnswerID(index int) string {
	if index < 0 || index >= len(r.currentQuestion.Answers) {
		return ""
	}
	return r.currentQuestion.Answers[index].AnswerID
}

// check if the current option is in markedOptions
func (r *QuestionsViewModern) isOptionMarked(index int) bool {
	return r.currentQuestion.Answers[index].GetIsMarked()
}

func (r *QuestionsViewModern) GetCurrentOptions() []*types.Answer {
	var answers []*types.Answer
	for _, v := range r.currentQuestion.Answers {
		if v.GetIsMarked() {
			answers = append(answers, v)
		}
	}
	return answers
}

func (r *QuestionsViewModern) getOptionByIndex(index int) *types.Answer {
	if index < 0 || index >= len(r.currentQuestion.Answers) {
		return nil
	}
	return r.currentQuestion.Answers[index]
}

func (r *QuestionsViewModern) ToggleCurrentMarkedOption() types.AnsweredState {
	return r.ToggleMarkedOption(r.markerPosition)
}

// toggle the current option and return true if the option is correct
func (r *QuestionsViewModern) ToggleMarkedOption(index int) types.AnsweredState {
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

func (r *QuestionsViewModern) checkAllCorrectMarked() bool {
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
func (r *QuestionsViewModern) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
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
func (r *QuestionsViewModern) removeCurrentOption(index int) {
	for i, v := range r.currentQuestion.Answers {
		if v.AnswerID == r.indexToAnswerID(index) {
			r.currentQuestion.Answers[i].SetIsMarked(false)
		}
	}
}

func (r *QuestionsViewModern) ShowExplanation() {
	if r.currentQuestion.Explanation != "" {
		r.explainationView.SetText(r.currentQuestion.Explanation)
	} else {
		r.explainationView.SetText("[red]No explanation available[-]")
	}
	log.Println("Show explanation")
}

func (r *QuestionsViewModern) GetOptionsLenght() int {
	if r.currentQuestion == nil {
		return 0
	}
	return len(r.currentQuestion.Answers)
}
