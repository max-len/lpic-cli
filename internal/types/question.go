package types

type Question struct {
	ID            int           `bson:"_id" json:"ID,string"`
	Text          string        `bson:"questionText"`
	Answers       []*Answer     `bson:"answers"`
	Explanation   string        `bson:"explanation,omitempty"`
	AnsweredState AnsweredState `bson:"answeredState,omitempty"`
	isImportant   bool          `bson:"important,omitempty"`
}

func (q *Question) GetAnsweredOptions() []*Answer {
	var result []*Answer
	for _, answer := range q.Answers {
		if answer.GetIsMarked() {
			result = append(result, answer)
		}
	}
	return result
}

func (q *Question) IsSingleAnswer() bool {
	count := 0
	for _, answer := range q.Answers {
		if answer.IsCorrect {
			count++
		}
		if count > 1 {
			return false
		}
	}
	return true
}

func (question *Question) SetAnsweredState(AnsweredState AnsweredState) *Question {
	question.AnsweredState = AnsweredState
	return question
}

func (question *Question) SetIsImportant(isImportant bool) *Question {
	question.isImportant = isImportant
	return question
}

func (question *Question) GetIsImportant() bool {
	return question.isImportant
}
