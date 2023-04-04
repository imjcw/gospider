package db

import (
	"fmt"

	"github.com/go-redis/redis"
	"github.com/spf13/viper"
)

// RedisHandler .
var RedisHandler *redis.Client

// InitRedisHandler .
func InitRedisHandler() {
	fmt.Println(fmt.Sprintf(`%s:%d`, viper.GetString("redis.host"), viper.GetInt("redis.port")))
	RedisHandler = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf(`%s:%d`, viper.GetString("redis.host"), viper.GetInt("redis.port")),
	})
	_, err := RedisHandler.Ping().Result()
	if err != nil {
		fmt.Println(err)
	}
}
