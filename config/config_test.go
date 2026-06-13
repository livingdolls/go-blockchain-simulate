package config

import (
	"strings"
	"testing"
	"time"

	"github.com/livingdolls/go-blockchain-simulate/app/entity"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validBaseCfg membuat Config dengan nilai yang lolos semua validasi
// kecuali field yang di-override oleh test case.
func validBaseCfg() Config {
	return Config{
		JWT: JWTConfig{
			UserSecret:  strings.Repeat("a", 32),
			AdminSecret: strings.Repeat("b", 32),
			UserTTL:     24 * time.Hour,
			AdminTTL:    24 * time.Hour,
		},
		Database: DatabaseConfig{DSN: "user:pass@tcp(localhost:3306)/db"},
		Redis:    RedisConfig{Addr: "localhost:6379"},
		RabbitMQ: RabbitMQConfig{URL: "amqp://guest:guest@localhost:5672/"},
		Server: ServerConfig{
			Port:            3010,
			ShutdownTimeout: 30 * time.Second,
		},
	}
}

func TestValidate_HappyPath(t *testing.T) {
	cfg := validBaseCfg()
	assert.NoError(t, cfg.validate())
}

func TestValidate_EmptyUserSecret(t *testing.T) {
	cfg := validBaseCfg()
	cfg.JWT.UserSecret = ""
	err := cfg.validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user_secret wajib diisi")
}

func TestValidate_UserSecretTooShort(t *testing.T) {
	cfg := validBaseCfg()
	cfg.JWT.UserSecret = "short"
	err := cfg.validate()
	require.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrJWTSecretTooShort)
	assert.Contains(t, err.Error(), "user_secret 5 byte")
}

func TestValidate_UserSecretExactly32_OK(t *testing.T) {
	cfg := validBaseCfg()
	cfg.JWT.UserSecret = strings.Repeat("x", 32) // exactly 32
	assert.NoError(t, cfg.validate())
}

func TestValidate_UserSecret31Byte_Fail(t *testing.T) {
	cfg := validBaseCfg()
	cfg.JWT.UserSecret = strings.Repeat("x", 31) // 1 byte short
	err := cfg.validate()
	require.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrJWTSecretTooShort)
}

func TestValidate_EmptyAdminSecret(t *testing.T) {
	cfg := validBaseCfg()
	cfg.JWT.AdminSecret = ""
	err := cfg.validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "admin_secret wajib diisi")
}

func TestValidate_AdminSecretTooShort(t *testing.T) {
	cfg := validBaseCfg()
	cfg.JWT.AdminSecret = "tooshort"
	err := cfg.validate()
	require.Error(t, err)
	assert.ErrorIs(t, err, entity.ErrJWTSecretTooShort)
}

func TestValidate_EmptyDSN(t *testing.T) {
	cfg := validBaseCfg()
	cfg.Database.DSN = ""
	err := cfg.validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database.dsn")
}

func TestValidate_EmptyRabbitMQ(t *testing.T) {
	cfg := validBaseCfg()
	cfg.RabbitMQ.URL = ""
	err := cfg.validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rabbitmq.url")
}

func TestValidate_EmptyRedis(t *testing.T) {
	cfg := validBaseCfg()
	cfg.Redis.Addr = ""
	err := cfg.validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis.addr")
}

func TestValidate_InvalidServerPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"zero", 0},
		{"negative", -1},
		{"too high", 65536},
		{"way too high", 100000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validBaseCfg()
			cfg.Server.Port = tt.port
			err := cfg.validate()
			require.Error(t, err)
			assert.Contains(t, err.Error(), "server.port")
		})
	}

	// Boundary: 1 dan 65535 harus valid
	for _, port := range []int{1, 65535} {
		cfg := validBaseCfg()
		cfg.Server.Port = port
		assert.NoError(t, cfg.validate(), "port %d harus valid", port)
	}
}

func TestValidate_InvalidShutdownTimeout(t *testing.T) {
	cfg := validBaseCfg()
	cfg.Server.ShutdownTimeout = 0
	err := cfg.validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "shutdown_timeout")

	cfg.Server.ShutdownTimeout = -1 * time.Second
	err = cfg.validate()
	require.Error(t, err)
}

func TestSetDefaults_NoWeakJWTSecret(t *testing.T) {
	// Regression: setDefaults() menambahkan default JWT secret kosong ("")
	// agar viper AllKeys() mengenali kunci dan AutomaticEnv bisa bekerja.
	// Nilai "" TIDAK lolos validasi 32-byte, jadi deploy tanpa secret
	// akan gagal di validate().
	v := viper.New()
	setDefaults(v)
	assert.Equal(t, "", v.GetString("jwt.user_secret"),
		"jwt.user_secret harus kosong (validation akan trigger)")
	assert.Equal(t, "", v.GetString("jwt.admin_secret"),
		"jwt.admin_secret harus kosong (validation akan trigger)")
}
