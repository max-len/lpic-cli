package repository

import (
	"context"

	"github.com/SqiSch/lpic-cli/internal/types"
)

type QuestionRepository interface {
	UpsertQuestion(ctx context.Context, question *types.Question) error
	GetQuestion(ctx context.Context, id string) (*types.Question, error)
	DeleteQuestion(ctx context.Context, id string) error
	GetAnsweredQuestions() ([]types.QuestionStateDB, error)
}
