package dbManage

import (
	"log"

	"github.com/go-redis/redis"
)

func RedisInit() (Cli *redis.Client) {
	client := redis.NewClient(&redis.Options{Addr: "127.0.0.1:3478", Password: "*******"})

	pong, err := client.Ping().Result()
	if err != nil {
		log.Println("ping error", err.Error())
	}
	log.Println("ping result:", pong)
	return client
}
