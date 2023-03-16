package redis

import "github.com/gomodule/redigo/redis"

func (r *Redis) GetString(key string) (string, error) {
	c := r.pool.Get()
	defer c.Close()

	return redis.String(c.Do("GET", key))
}
