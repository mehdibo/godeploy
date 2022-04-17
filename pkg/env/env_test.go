package env

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestLoadDotEnv(t *testing.T) {
	LoadDotEnv()

	assert.Equal(t, "a_value", os.Getenv("VAR_A"))
	assert.Equal(t, "b_value", os.Getenv("VAR_B"))
	assert.Equal(t, "c_value_local", os.Getenv("VAR_C"))
	assert.Equal(t, "d_value_local", os.Getenv("VAR_D"))
	assert.Equal(t, "e_value_env", os.Getenv("VAR_E"))
	assert.Equal(t, "f_value_env", os.Getenv("VAR_F"))
	assert.Equal(t, "g_value_env_local", os.Getenv("VAR_G"))
	assert.Equal(t, "h_value_env_local", os.Getenv("VAR_H"))
}

func TestGet(t *testing.T) {
	LoadDotEnv()

	assert.Equal(t, "a_value", Get("VAR_A"))
	assert.Equal(t, "b_value", Get("VAR_B"))
	assert.Equal(t, "c_value_local", Get("VAR_C"))
	assert.Equal(t, "d_value_local", Get("VAR_D"))
	assert.Equal(t, "e_value_env", Get("VAR_E"))
	assert.Equal(t, "f_value_env", Get("VAR_F"))
	assert.Equal(t, "g_value_env_local", Get("VAR_G"))
	assert.Equal(t, "h_value_env_local", Get("VAR_H"))
}

func TestGetDefault(t *testing.T) {
	LoadDotEnv()

	assert.Equal(t, "a_value", GetDefault("VAR_A", "default_value"))
	assert.Equal(t, "default_value", GetDefault("NON_EXISTING", "default_value"))
	assert.Equal(t, "", GetDefault("EMPTY_VAR", "something"))

}
