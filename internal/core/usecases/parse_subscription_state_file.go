package usecases

import (
	"github.com/phoops/bellatrix/internal/core/entities"
	"github.com/pkg/errors"
)

type SubscriptionsFileParser interface {
	ParseSubscriptionFile(path string) (*entities.SubscriptionsRequestedState, error)
}

type ParseSubscriptionsStateFile struct {
	fileParser     SubscriptionsFileParser
	instancePrefix string
}

func NewParseSubscriptionsStateFile(fileParser SubscriptionsFileParser, instancePrefix string) *ParseSubscriptionsStateFile {
	return &ParseSubscriptionsStateFile{fileParser: fileParser, instancePrefix: instancePrefix}
}

func (u *ParseSubscriptionsStateFile) Execute(path string) (*entities.SubscriptionsRequestedState, error) {
	if len(path) == 0 {
		return nil, errors.Errorf("invalid path provided")
	}

	subsState, err := u.fileParser.ParseSubscriptionFile(path)

	if err != nil {
		return nil, errors.Wrap(err, "could not parse subscription file.")
	}

	// Attach the bellatrix prefix, to subs description, in order to distinguish
	// on orion the subs managed by this  program

	for _, subRequest := range subsState.SubscriptionsState {
		for _, subs := range subRequest.Subscriptions {
			fullPrefix := u.instancePrefix + BellatrixManagedSubscriptionsPrefix
			subs.Description = fullPrefix + subs.Description
		}
	}

	return subsState, nil
}
