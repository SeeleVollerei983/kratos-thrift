package data

import (
	"aboveThriftRPC/internal/conf"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

// Data .
type Data struct {
	redis *redis.Pool
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	// Initialize Redis connection
	dialOptions := []redis.DialOption{
		redis.DialReadTimeout(c.Redis.ReadTimeout.AsDuration()),
		redis.DialWriteTimeout(c.Redis.WriteTimeout.AsDuration()),
	}

	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", c.Redis.Addr, dialOptions...)
		},
	}

	cleanup := func() {
		logrus.Infof("closing the data resources")
		pool.Close()
	}
	return &Data{redis: pool}, cleanup, nil
}
