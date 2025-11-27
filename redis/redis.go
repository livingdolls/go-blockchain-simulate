package redis

import (
	"context"
	"fmt"
	"net/url"

	"github.com/redis/go-redis/v9"
)

func NewRedisMemory() (*redis.Client, error) {
	buildUrl := fmt.Sprintf("redis://:%s@%s:%s/%d", "", "localhost", "6379", 0)

	u, err := url.Parse(buildUrl)

	if err != nil {
		return nil, err
	}

	addr := u.Host
	password, _ := u.User.Password()

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return rdb, nil

}
