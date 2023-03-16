package cache

import (
	"fmt"
	"github.com/xnile/muxwaf/pkg/redis"
	"time"
)

type UserCache interface {
	AddUserToken(id int64, token string) error
	GetUserToken(id int64) (string, error)
}

type userCache struct {
	redis *redis.Redis
}

func NewUserCache(redis *redis.Redis) UserCache {
	return &userCache{
		redis: redis,
	}
}

func (c *userCache) AddUserToken(id int64, token string) error {
	key := fmt.Sprintf("muxwaf:user:token:%s:string", id)
	ttl := time.Second * 3600
	err := c.redis.Set(key, token, ttl)
	return err
}

func (c *userCache) GetUserToken(id int64) (string, error) {
	key := fmt.Sprintf("muxwaf:user:token:%s:string", id)

	return c.redis.GetString(key)
}
