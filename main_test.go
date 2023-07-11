package main

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/cyberark/conjur-api-go/conjurapi"
	"github.com/stretchr/testify/assert"
)

func TestCheckEnvironmentVariables_AllVariableSet(t *testing.T) {
	os.Setenv("CONJUR_APPLIANCE_URL", "https://example.com")
	os.Setenv("CONJUR_ACCOUNT", "myaccount")
	os.Setenv("CONJUR_AUTHN_JWT_SERVICE_ID", "service_id")
	os.Setenv("CONJUR_AUTHN_JWT_TOKEN", "mytoken")
	os.Setenv("CONJUR_SECRET_ID", "validSecretID")

	err := checkEnvironmentVariables()
	assert.NoError(t, err)
	if err != nil {
		t.Errorf("Expected no err, but got: %v", err)
	}
}

type MockConjurClient struct{}

func (m *MockConjurClient) RetrieveSecret(variableIdentifier string) ([]byte, error) {
	if variableIdentifier == "validSecretID" {
		return []byte("mockedSecretValue"), nil
	}
	return nil, fmt.Errorf("Mocked error: secret retrieval failed")
}

func TestMain(t *testing.T) {
	oldConfig := loadConfig
	oldclientFromEnvironment := clientFromEnvironment
	defer func() {
		loadConfig = oldConfig
		clientFromEnvironment = oldclientFromEnvironment
	}()

	home := os.Getenv("HOME")
	os.Setenv("HOME", home)

	loadConfig = func() (conjurapi.Config, error) {
		return conjurapi.Config{Account: "myaccount", ApplianceURL: "example.com", NetRCPath: path.Join(home, ".netrc")}, nil
	}

	clientFromEnvironment = func(config conjurapi.Config) (*conjurapi.Client, error) {
		return &conjurapi.Client{}, nil
	}

	secretValue, err := (&MockConjurClient{}).RetrieveSecret("validSecretID")
	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}

	expectedSecretValue := []byte("mockedSecretValue")
	if string(secretValue) != string(expectedSecretValue) {
		t.Errorf("Expected secret value: %s, but got: %s", string(expectedSecretValue), string(secretValue))
	}

	_, err = (&MockConjurClient{}).RetrieveSecret("invalidSecretID")
	if err == nil {
		t.Error("Expected error during secret retrieval, but got nil")
	}
}

func TestCheckEnvironmentVariables_MissingVariable(t *testing.T) {
	os.Unsetenv("CONJUR_APPLIANCE_URL")

	err := checkEnvironmentVariables()
	if err == nil {
		t.Errorf("Expected an error, but got nil")
	}
}
