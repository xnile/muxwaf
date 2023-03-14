package cache

import "github.com/xnile/muxwaf/pkg/redis"

type Cache struct {
	User UserCache
}

func New(redis *redis.Redis) *Cache {
	return &Cache{
		User: NewUserCache(redis),
	}
}
