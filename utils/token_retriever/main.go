package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/phoops/bellatrix/internal/core/entities"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func mustHaveEnvVariable(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	panic(fmt.Sprintf("Could not start the application env variable %s is missing", key))
}

var (
	wobcomAuthServerLoginURLEnvVariable     = "AUTH_SERVER_LOGIN_URL"
	wobcomUsernameEnvVariable               = "WOBCOM_USERNAME"
	wobcomPasswordEnvVariable               = "WOBCOM_PASSWORD"
	bellatrixInputStateFilePathEnvVariable  = "BELLATRIX_INPUT_STATE_FILE_PATH"
	bellatrixOutputStateFilePathEnvVariable = "BELLATRIX_OUTPUT_STATE_FILE_PATH"
)

func main() {
	wobcomAuthServerLoginURL := mustHaveEnvVariable(wobcomAuthServerLoginURLEnvVariable)
	wobcomUsername := mustHaveEnvVariable(wobcomUsernameEnvVariable)
	wobcomPassword := mustHaveEnvVariable(wobcomPasswordEnvVariable)
	bellatrixInputStateFilePath := mustHaveEnvVariable(bellatrixInputStateFilePathEnvVariable)
	bellatrixOutputStateFilePath := mustHaveEnvVariable(bellatrixOutputStateFilePathEnvVariable)

	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(errors.Wrap(err, "cannot initialize logger"))
	}

	logger.Info(
		"Wobcom token retriver started",
		zap.String("username", wobcomUsername),
		zap.String("auth_server", wobcomAuthServerLoginURL),
	)

	stateFileContent, err := os.ReadFile(bellatrixInputStateFilePath)
	if err != nil {
		logger.Fatal("Could not read state input file", zap.Error(err))
	}

	// decode the state to json

	var stateFile entities.SubscriptionsRequestedState
	err = json.Unmarshal(stateFileContent, &stateFile)
	if err != nil {
		logger.Fatal("Could not unmarshal the input state file", zap.Error(err))
	}

	formPayload := url.Values{
		"username":   []string{wobcomUsername},
		"password":   []string{wobcomPassword},
		"grant_type": []string{"password"},
		"client_id":  []string{"api"},
		"scope":      []string{"entity:read subscription:read subscription:delete subscription:write subscription:create"},
	}

	resp, err := http.PostForm(wobcomAuthServerLoginURL, formPayload)
	if err != nil {
		logger.Fatal("Could not perform http request for login", zap.Error(err))
	}

	if resp.StatusCode != 200 {
		logger.Fatal(
			"Could not perform login request",
			zap.Int("status_code", resp.StatusCode),
		)
	}

	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		logger.Fatal("Could not decode login request response body", zap.Error(err))
	}
	accessToken, found := result["access_token"].(string)
	if !found {
		logger.Fatal(
			"Could not find the access_token inside login body response",
			zap.Any("login_response_body", result),
		)
	}

	stateFile.ClientOptions.AdditionalHeaders["Authorization"] = "Bearer " + accessToken

	// convert the state back to json and save

	outputStateFileContents, err := json.Marshal(stateFile)
	if err != nil {
		logger.Fatal("Could not marshal the output state file", zap.Error(err))
	}

	err = os.WriteFile(bellatrixOutputStateFilePath, outputStateFileContents, 0777)
	if err != nil {
		logger.Fatal("Could not save output state file", zap.String("path", bellatrixOutputStateFilePath), zap.Error(err))
	}

	logger.Info("Token retrieved, bellatrix state updated.")
}
