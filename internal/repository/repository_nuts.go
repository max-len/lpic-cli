package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/SqiSch/lpic-cli/internal/types"
	"github.com/nutsdb/nutsdb"
)

var _ QuestionRepository = (*NutsQuestionRepository)(nil)

type NutsQuestionRepository struct {
	db *nutsdb.DB
}

// NewNutsQuestionRepositoryWithDir creates a NutsDB-backed repository at the provided baseDir.
// If baseDir is empty, $HOME is used. Data is always stored inside a directory named ".nutsdb".
// If baseDir already ends with ".nutsdb" it is used directly; otherwise the ".nutsdb" segment is appended.
func NewNutsQuestionRepositoryWithDir(baseDir string) *NutsQuestionRepository {
	var resolved string
	if strings.TrimSpace(baseDir) == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("failed to get user home directory: %v", err)
		}
		resolved = filepath.Join(homeDir, ".nutsdb")
	} else {
		// Expand leading ~
		if strings.HasPrefix(baseDir, "~") {
			homeDir, err := os.UserHomeDir()
			if err == nil {
				baseDir = filepath.Join(homeDir, strings.TrimPrefix(baseDir, "~"))
			}
		}
		// If the user points to an existing .nutsdb directory, keep it; else append
		if filepath.Base(baseDir) == ".nutsdb" {
			resolved = baseDir
		} else {
			resolved = filepath.Join(baseDir, ".nutsdb")
		}
	}

	if err := os.MkdirAll(resolved, 0o755); err != nil {
		log.Fatalf("failed to create nutsdb directory %s: %v", resolved, err)
	}

	db, err := nutsdb.Open(nutsdb.DefaultOptions, nutsdb.WithDir(resolved))
	if err != nil {
		log.Fatalf("failed to open nutsdb at %s: %v", resolved, err)
	}

	ensureBucketExists(db, "answered_questions")

	return &NutsQuestionRepository{db: db}
}

// NewNutsQuestionRepository keeps backwards compatibility with previous no-arg constructor.
// It stores data under $HOME/.nutsdb.
func NewNutsQuestionRepository() *NutsQuestionRepository {
	return NewNutsQuestionRepositoryWithDir("")
}

func ensureBucketExists(db *nutsdb.DB, bucketName string) error {
	return db.Update(func(tx *nutsdb.Tx) error {
		if !tx.ExistBucket(nutsdb.DataStructureBTree, bucketName) {
			return tx.NewBucket(nutsdb.DataStructureBTree, bucketName)
		}
		return nil
	})
}

func (n *NutsQuestionRepository) Close() error {
	if err := n.db.Close(); err != nil {
		return fmt.Errorf("failed to close nutsdb: %w", err)
	}
	return nil
}

// DeleteQuestion implements QuestionRepository.
func (n *NutsQuestionRepository) DeleteQuestion(ctx context.Context, id string) error {
	panic("unimplemented")
}

// GetQuestion implements QuestionRepository.
func (n *NutsQuestionRepository) GetQuestion(ctx context.Context, id string) (*types.Question, error) {
	panic("unimplemented")
}

// UpsertQuestion implements QuestionRepository.
func (n *NutsQuestionRepository) UpsertQuestion(ctx context.Context, question *types.Question) error {
	markedAnswers := make([]string, 0)
	for _, answer := range question.Answers {
		if answer.GetIsMarked() {
			markedAnswers = append(markedAnswers, answer.AnswerID)
		}
	}
	state := types.QuestionStateDB{
		QuestionID:    question.ID,
		MarkedAnswers: markedAnswers,
		AnsweredState: question.AnsweredState,
		Important:     question.GetIsImportant(),
	}

	return n.db.Update(func(tx *nutsdb.Tx) error {
		key := fmt.Sprintf("%d", question.ID)
		value, err := json.Marshal(state)
		if err != nil {
			return fmt.Errorf("failed to marshal question state: %w", err)
		}
		return tx.Put("answered_questions", []byte(key), value, 0)
	})
}

func (n *NutsQuestionRepository) GetAnsweredQuestions() ([]types.QuestionStateDB, error) {
	var questionStates []types.QuestionStateDB

	err := n.db.View(func(tx *nutsdb.Tx) error {
		_, values, err := tx.GetAll("answered_questions")
		if err != nil {
			return fmt.Errorf("failed to retrieve certification set: %w", err)
		}

		for _, k := range values {
			var questionState types.QuestionStateDB
			if err := json.Unmarshal(k, &questionState); err != nil {
				return fmt.Errorf("failed to unmarshal certification set: %w", err)
			}
			questionStates = append(questionStates, questionState)

		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return questionStates, nil
}
