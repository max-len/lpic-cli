package types

type QuestionStateDB struct {
	QuestionID    int
	MarkedAnswers []string
	AnsweredState AnsweredState
}
