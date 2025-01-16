package state

import (
	"encoding/json"
	"os"

	"github.com/phoops/bellatrix/internal/core/entities"
	"go.uber.org/zap"
)

type Parser struct {
	logger *zap.Logger
}

func NewParser(logger *zap.Logger) *Parser {
	return &Parser{logger: logger}
}

func (p *Parser) ParseSubscriptionFile(path string) (*entities.SubscriptionsRequestedState, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		p.logger.Debug("could not read the file provided", zap.Error(err), zap.String("file_path", path))
		return nil, err
	}

	var subsState *entities.SubscriptionsRequestedState

	err = json.Unmarshal(content, &subsState)
	if err != nil {
		p.logger.Debug("could not unmarshal the file content", zap.Error(err), zap.String("file_path", path))
		return nil, err
	}

	return subsState, nil
}
