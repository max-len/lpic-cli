package database

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/SqiSch/lpic-cli/internal/types"
)

func LoadFullData(filename string) ([]*types.CertificationSet, error) {
	var certSets []*types.CertificationSet
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&certSets); err != nil {
		return nil, fmt.Errorf("failed to decode certification sets: %w", err)
	}
	return certSets, nil

}

func LoadDatabaseFromFile(filename string, certID string) (*types.CertificationSet, error) {
	var certSets []types.CertificationSet
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&certSets); err != nil {
		return nil, fmt.Errorf("failed to decode certification sets: %w", err)
	}

	for _, cs := range certSets {
		if cs.CertificationID == certID {
			// Sort the questions by ID
			keys := make([]*types.Question, 0, len(cs.Questions))
			for _, v := range cs.Questions {
				keys = append(keys, v)
			}
			sort.Slice(keys, func(i, j int) bool {
				return keys[i].ID < keys[j].ID
			})
			return &cs, nil
		}
	}

	return nil, fmt.Errorf("certification set with ID %s not found", certID)
}
