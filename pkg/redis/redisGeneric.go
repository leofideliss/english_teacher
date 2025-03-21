package pkg

import (
    "context"
    "github.com/redis/go-redis/v9"
)

type respositoryRedis struct{
    r *redis.Client
}

var ctx = context.Background()
var repository respositoryRedis

func init(){
    connectRedis()
}

func connectRedis() {
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",              
        DB:       0,                
    })
    repository.r = rdb
}

func (r respositoryRedis) PushRedis( key , value string) error{
    err := r.r.RPush(ctx , key , value).Err()
    return err
}

