package conjurapi

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cyberark/authn-jwt-gitlab/internal/conjurapi/authn"
	"github.com/cyberark/authn-jwt-gitlab/internal/conjurapi/response"
	"github.com/stretchr/testify/assert"
)

func TestClient_LoadPolicy(t *testing.T) {
	config := &Config{}
	config.mergeEnv()

	apiKey := os.Getenv("CONJUR_AUTHN_API_KEY")
	login := os.Getenv("CONJUR_AUTHN_LOGIN")

	conjur, err := NewClientFromKey(*config, authn.LoginPair{Login: login, APIKey: apiKey})
	assert.NoError(t, err)

	randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))

	t.Run("Successfully load policy", func(t *testing.T) {
		username := "alice"
		policy := fmt.Sprintf(`
- !user %s
`, username)

		resp, err := conjur.LoadPolicy(
			PolicyModePut,
			"root",
			strings.NewReader(policy),
		)

		assert.NoError(t, err)
		assert.GreaterOrEqual(t, resp.Version, uint32(1))
	})

	t.Run("A new role is reported in the policy load response", func(t *testing.T) {
		const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
		result := make([]byte, 12)
		for i := range result {
			result[i] = chars[randomizer.Intn(len(chars))]
		}

		username := string(result)
		policy := fmt.Sprintf(`
- !user %s
`, username)

		resp, err := conjur.LoadPolicy(
			PolicyModePut,
			"root",
			strings.NewReader(policy),
		)

		assert.NoError(t, err)
		createdRole, ok := resp.CreatedRoles["cucumber:user:"+username]
		assert.NotEmpty(t, createdRole.ID)
		assert.NotEmpty(t, createdRole.APIKey)
		assert.True(t, ok)
	})

	t.Run("Given invalid login credentials", func(t *testing.T) {
		login = "invalid-user"

		t.Run("Returns 401", func(t *testing.T) {
			conjur, err := NewClientFromKey(*config, authn.LoginPair{Login: login, APIKey: apiKey})
			assert.NoError(t, err)

			resp, err := conjur.LoadPolicy(PolicyModePut, "root", strings.NewReader(""))

			assert.Error(t, err)
			assert.Nil(t, resp)
			assert.IsType(t, &response.ConjurError{}, err)
			conjurError := err.(*response.ConjurError)
			assert.Equal(t, 401, conjurError.Code)
		})

	})
}
