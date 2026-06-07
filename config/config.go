package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config adalah struktur konfigurasi utama aplikasi.
// Semua nilai dapat dioverride lewat environment variable.
// Lihat config.local.yaml.example untuk daftar lengkap environment variable.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Logger   LoggerConfig   `mapstructure:"logger"`
	Server   ServerConfig   `mapstructure:"server"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Version     string `mapstructure:"version"`
	Environment string `mapstructure:"environment"`
}

type DatabaseConfig struct {
	Driver          string        `mapstructure:"driver"`
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Addr         string `mapstructure:"addr"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type RabbitMQConfig struct {
	URL      string `mapstructure:"url"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	UserSecret   string        `mapstructure:"user_secret"`
	AdminSecret  string        `mapstructure:"admin_secret"`
	UserTTL      time.Duration `mapstructure:"user_ttl"`
	AdminTTL     time.Duration `mapstructure:"admin_ttl"`
	CookieMaxAge time.Duration `mapstructure:"cookie_max_age"`
}

type LoggerConfig struct {
	Level string `mapstructure:"level"`
	Path  string `mapstructure:"path"`
}

type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	AllowedOrigins  []string      `mapstructure:"allowed_origins"`
}

// Load memuat konfigurasi dari file dan environment variable.
// Prioritas: env var > config.local.yaml > config.yaml > default
func Load(configPath string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// Baca config.local.yaml jika ada (untuk override lokal, biasanya berisi secret).
	v.SetConfigName("config.local")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")
	if err := v.MergeInConfig(); err != nil {
		// File local tidak wajib ada, abaikan error file-not-found.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Untuk path eksplisit, error harus dikembalikan.
			if configPath != "" {
				return nil, fmt.Errorf("gagal membaca konfigurasi: %w", err)
			}
		}
	}
	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// File config utama tidak ditemukan, lanjut dengan env + default.
		} else {
			return nil, fmt.Errorf("gagal membaca config.yaml: %w", err)
		}
	}

	// Override via environment variable.
	// Contoh: APP_NAME, DATABASE_DSN, JWT_USER_SECRET, REDIS_ADDR, dsb.
	v.SetEnvPrefix("")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("gagal parsing konfigurasi: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validasi konfigurasi gagal: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "blockchain")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.environment", "development")

	v.SetDefault("database.driver", "mysql")
	v.SetDefault("database.dsn", "yurina:hirate@tcp(127.0.0.1:3306)/blockchain?parseTime=true&loc=Local")
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.max_idle_conns", 20)
	v.SetDefault("database.conn_max_lifetime", "5m")

	v.SetDefault("redis.addr", "localhost:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.pool_size", 10)
	v.SetDefault("redis.min_idle_conns", 5)

	v.SetDefault("rabbitmq.url", "amqp://guest:guest@localhost:5672/")
	v.SetDefault("rabbitmq.pool_size", 10)

	v.SetDefault("jwt.user_secret", "change-me-in-production")
	v.SetDefault("jwt.admin_secret", "change-me-in-production-admin")
	v.SetDefault("jwt.user_ttl", "24h")
	v.SetDefault("jwt.admin_ttl", "24h")
	v.SetDefault("jwt.cookie_max_age", "24h")

	v.SetDefault("logger.level", "debug")
	v.SetDefault("logger.path", "")

	v.SetDefault("server.host", "")
	v.SetDefault("server.port", 3010)
	v.SetDefault("server.shutdown_timeout", "30s")
	v.SetDefault("server.allowed_origins", []string{
		"http://localhost:3000",
		"http://localhost:3001",
	})
}

func (c *Config) validate() error {
	if c.JWT.UserSecret == "" {
		return fmt.Errorf("jwt.user_secret wajib diisi")
	}
	if c.JWT.AdminSecret == "" {
		return fmt.Errorf("jwt.admin_secret wajib diisi")
	}
	if c.Database.DSN == "" {
		return fmt.Errorf("database.dsn wajib diisi")
	}
	if c.RabbitMQ.URL == "" {
		return fmt.Errorf("rabbitmq.url wajib diisi")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis.addr wajib diisi")
	}
	return nil
}

// Addr mengembalikan alamat HTTP server dalam format "host:port".
func (s ServerConfig) Addr() string {
	host := s.Host
	if host == "" {
		host = "0.0.0.0"
	}
	return fmt.Sprintf("%s:%d", host, s.Port)
}
