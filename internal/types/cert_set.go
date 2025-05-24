package types

import (
	"fmt"
	"maps"
	"slices"
)

type CertificationSet struct {
	ID                       string             `bson:"_id"`
	CertificationID          string             `bson:"testsetId"`
	CertificationName        string             `bson:"testsetName"`
	CertificationDescription string             `bson:"testsetDescription"`
	Questions                map[int]*Question  `bson:"questions"`
	Testsets                 map[string]Testset `bson:"testset"`
}

func (c *CertificationSet) CertificationSetMapToSlice() []*Question {
	vals := slices.Collect(maps.Values(c.Questions))
	return vals
}

func (c *CertificationSet) GetQuestionsForTestset(id string, filterCorrect bool, stateDB []QuestionStateDB) ([]*Question, error) {
	testset, ok := c.Testsets[id]
	if !ok && id != "" {
		return nil, fmt.Errorf("setset not found")
	}

	markedAnswer := func(question *Question, markedAnswers []QuestionStateDB) {
		for _, questionState := range stateDB {
			if question.ID == questionState.QuestionID {
				question.AnsweredState = questionState.AnsweredState
				for _, answer := range question.Answers {
					for _, markedAnswer := range questionState.MarkedAnswers {
						if answer.AnswerID == markedAnswer {
							answer.SetIsMarked(true)
						}
					}
				}
			}
		}
	}

	if id == "" {
		if len(c.Questions) == 0 {
			return nil, fmt.Errorf("no questions in certification set")
		}
		questions := c.CertificationSetMapToSlice()

		for _, question := range questions {
			markedAnswer(question, stateDB)
		}

		return questions, nil
	}

	if len(testset.QuestionsIds) == 0 {
		return nil, fmt.Errorf("no questions in testset")
	}

	var questions []*Question
	for _, questionID := range testset.QuestionsIds {
		question, ok := c.Questions[questionID]
		if !ok {
			return nil, fmt.Errorf("questionid not found%s", questionID)
		}

		markedAnswer(question, stateDB)

		for _, questionState := range stateDB {
			if question.ID == questionState.QuestionID {
				question.AnsweredState = questionState.AnsweredState
				for _, answer := range question.Answers {
					for _, markedAnswer := range questionState.MarkedAnswers {
						if answer.AnswerID == markedAnswer {
							answer.SetIsMarked(true)
						}
					}
				}
			}
		}

		if filterCorrect && question.AnsweredState == AnsweredTrue {
			continue

		}

		questions = append(questions, question)
	}

	//sort the questions by their ID
	for i := 0; i < len(questions)-1; i++ {
		for j := 0; j < len(questions)-i-1; j++ {
			if questions[j].ID > questions[j+1].ID {
				questions[j], questions[j+1] = questions[j+1], questions[j]
			}
		}
	}

	return questions, nil
}
