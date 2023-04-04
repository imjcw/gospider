package main

import (
	"fmt"

	_ "github.com/imjcw/gospider/config"
	"github.com/imjcw/gospider/db"
)

func main() {
	fmt.Println(1)
	db.InitRedisHandler()
	val, err := db.RedisHandler.Get("mykey").Result()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(val)
}
