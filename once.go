package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// 单例模式
// 无论调用多少次，只会初始化一次

type Instance struct {
	Client *http.Client
}

var (
	instance *Instance
	once     sync.Once
)

func GetClient() *Instance {
	once.Do(func() {
		instance = &Instance{}
	})
	return instance
}

// 数据库初始化
var (
	db     *sql.DB
	dbOnce sync.Once
)

func InitDB() {
	once.Do(func() {
		db, _ = sql.Open("mysql", "xxx")
	})
}

// Lazy initialization
var (
	cache    map[string]string
	lazyOnce sync.Once
)

func GetCache() map[string]string {
	lazyOnce.Do(func() {
		cache = make(map[string]string)
	})
	return cache
}

// 后台一个协程

var workerOnce sync.Once

func StartWorker() {
	workerOnce.Do(func() {
		go func() {
			for {
				fmt.Println("running...")
				time.Sleep(time.Second)
			}
		}()
	})
}

func main() {
	client := GetClient()
	fmt.Println(client)
	InitDB()
	fmt.Println(db)
	cache := GetCache()
	fmt.Println(cache)
	StartWorker()
}
