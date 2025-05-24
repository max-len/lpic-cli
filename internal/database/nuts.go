package database

// import (
// 	"encoding/json"
// 	"fmt"

// 	"github.com/SqiSch/lpic-cli/internal/types"
// 	"github.com/nutsdb/nutsdb"
// )

// func LoadAnsweredQuestionsDB(db *nutsdb.DB) ([]types.QuestionStateDB, error) {
// 	var questionStates []types.QuestionStateDB

// 	err := db.View(func(tx *nutsdb.Tx) error {
// 		_, values, err := tx.GetAll("answered_questions")
// 		if err != nil {
// 			return fmt.Errorf("failed to retrieve certification set: %w", err)
// 		}

// 		for _, k := range values {
// 			var questionState types.QuestionStateDB
// 			if err := json.Unmarshal(k, &questionState); err != nil {
// 				return fmt.Errorf("failed to unmarshal certification set: %w", err)
// 			}
// 			questionStates = append(questionStates, questionState)

// 		}

// 		return nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	return questionStates, nil
// }

// func SaveAnsweredQuestion(db *nutsdb.DB, question *types.Question) error {

// 	markedAnswers := make([]string, 0)
// 	for _, answer := range question.Answers {
// 		if answer.GetIsMarked() {
// 			markedAnswers = append(markedAnswers, answer.AnswerID)
// 		}
// 	}
// 	state := types.QuestionStateDB{
// 		QuestionID:    question.ID,
// 		MarkedAnswers: markedAnswers,
// 		AnsweredState: question.AnsweredState,
// 	}

// 	return db.Update(func(tx *nutsdb.Tx) error {
// 		key := fmt.Sprintf("%d", question.ID)
// 		value, err := json.Marshal(state)
// 		if err != nil {
// 			return fmt.Errorf("failed to marshal question state: %w", err)
// 		}
// 		return tx.Put("answered_questions", []byte(key), value, 0)
// 	})
// }
