package app

import (
	"github.com/livingdolls/go-blockchain-simulate/database"
	"github.com/livingdolls/go-blockchain-simulate/logger"
	"github.com/livingdolls/go-blockchain-simulate/rabbitmq"
	"github.com/livingdolls/go-blockchain-simulate/redis"
	"github.com/livingdolls/go-blockchain-simulate/security"
)

// initializeInfrastructure membuka koneksi database, redis, rabbitmq,
// dan adapter JWT berdasarkan konfigurasi. Bila ada yang gagal, seluruh
// resource yang sudah dibuka akan di-close untuk mencegah leak.
func (d *AppDependencies) initializeInfrastructure() error {
	// Database
	db, err := database.NewDBConn(database.Config{
		Driver:          d.Config.Database.Driver,
		DSN:             d.Config.Database.DSN,
		MaxOpenConns:    d.Config.Database.MaxOpenConns,
		MaxIdleConns:    d.Config.Database.MaxIdleConns,
		ConnMaxLifetime: d.Config.Database.ConnMaxLifetime,
	})
	if err != nil {
		return err
	}
	d.DBConn = db

	// Redis
	redisClient, err := redis.NewRedisClient(redis.Config{
		Addr:         d.Config.Redis.Addr,
		Password:     d.Config.Redis.Password,
		DB:           d.Config.Redis.DB,
		PoolSize:     d.Config.Redis.PoolSize,
		MinIdleConns: d.Config.Redis.MinIdleConns,
	})
	if err != nil {
		_ = db.Close()
		return err
	}
	d.RedisClient = redisClient

	redisServices, err := redis.NewMemoryAdapter(redisClient, 1024)
	if err != nil {
		_ = redisClient.Close()
		_ = db.Close()
		return err
	}
	d.RedisServices = redisServices

	// RabbitMQ
	rmqClient, err := rabbitmq.NewClient(d.Config.RabbitMQ.URL, d.Config.RabbitMQ.PoolSize)
	if err != nil {
		_ = redisClient.Close()
		_ = db.Close()
		return err
	}
	d.RMQClient = rmqClient

	// JWT
	d.JWT = security.NewJWTAdapter(d.Config.JWT.UserSecret, d.Config.JWT.UserTTL)
	d.JWTAdmin = security.NewAdminJWTAdapter(d.Config.JWT.AdminSecret, d.Config.JWT.AdminTTL)

	logger.LogInfo("infrastruktur berhasil diinisialisasi")
	return nil
}
