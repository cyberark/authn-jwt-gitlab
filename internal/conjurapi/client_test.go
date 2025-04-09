package conjurapi

import (
	"testing"
	"fmt"
	"net/http"

	"github.com/cyberark/authn-jwt-gitlab/internal/conjurapi/authn"
	"github.com/stretchr/testify/assert"
)

func TestNewClientFromKey(t *testing.T) {
	t.Run("Has authenticator of type APIKeyAuthenticator", func(t *testing.T) {
		client, err := NewClientFromKey(
			Config{Account: "account", ApplianceURL: "appliance-url"},
			authn.LoginPair{Login: "login", APIKey: "api-key"},
		)

		assert.NoError(t, err)
		assert.IsType(t, &authn.APIKeyAuthenticator{}, client.authenticator)
	})
}

func TestClient_GetConfig(t *testing.T) {
	t.Run("Returns Client Config", func(t *testing.T) {
		expectedConfig := Config{
			Account:      "some-account",
			ApplianceURL: "some-appliance-url",
			NetRCPath:    "some-netrc-path",
			SSLCert:      "some-ssl-cert",
			SSLCertPath:  "some-ssl-cert-path",
		}
		client := Client{
			config: expectedConfig,
		}

		assert.EqualValues(t, client.GetConfig(), expectedConfig)
	})
}

func TestNewClientFromTokenFile(t *testing.T) {
	t.Run("Has authenticator of type TokenFileAuthenticator", func(t *testing.T) {
		client, err := NewClientFromTokenFile(Config{Account: "account", ApplianceURL: "appliance-url"}, "token-file")

		assert.NoError(t, err)
		assert.IsType(t, &authn.TokenFileAuthenticator{}, client.authenticator)
	})
}

func Test_newClientWithAuthenticator(t *testing.T) {
	t.Run("Returns nil and error for invalid config", func(t *testing.T) {
		client, err := newClientWithAuthenticator(Config{}, nil)

		assert.Nil(t, client)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Must specify")
	})

	t.Run("Returns client without error for valid config", func(t *testing.T) {
		client, err := newClientWithAuthenticator(Config{Account: "account", ApplianceURL: "appliance-url"}, nil)

		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestNewClientFromToken(t *testing.T) {
	t.Run("Has authenticator of type TokenAuthenticator", func(t *testing.T) {
		client, err := NewClientFromToken(Config{Account: "account", ApplianceURL: "appliance-url"}, "token")

		assert.NoError(t, err)
		assert.IsType(t, &authn.TokenAuthenticator{}, client.authenticator)
	})
}

func TestHandleHttpStatus(t *testing.T) {
	tests := []struct {
		statusCode    int
		expectedError error
	}{
		{
			statusCode:    http.StatusBadRequest,
			expectedError: fmt.Errorf("Bad request: received 400 status code"),
		},
		{
			statusCode:    http.StatusUnauthorized,
			expectedError: fmt.Errorf("Unauthorized: received 401 status code"),
		},
		{
			statusCode:    http.StatusForbidden,
			expectedError: fmt.Errorf("Forbidden: received 403 status code"),
		},
		{
			statusCode:    http.StatusInternalServerError,
			expectedError: fmt.Errorf("Internal server error: received 500 status code"),
		},
		{
			statusCode:    499,
			expectedError: fmt.Errorf("Unexpected error: received 499 status code"),
		},
		{
			statusCode:    200,
			expectedError: nil,
		},
		{
			statusCode:    300,
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Status code: %d", tt.statusCode), func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
			}

			err := HandleHttpStatus(resp)
			
			if tt.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.expectedError.Error())
			}
		})
	}
}
