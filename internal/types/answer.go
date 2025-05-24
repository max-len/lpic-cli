package types

type Answer struct {
	Text      string `bson:"text"`
	IsCorrect bool   `bson:"isCorrect"`
	AnswerID  string `bson:"answerId"`
	isMarked  bool
}

func (a *Answer) GetIsMarked() bool {
	return a.isMarked
}
func (a *Answer) SetIsMarked(isMarked bool) {
	a.isMarked = isMarked
}
