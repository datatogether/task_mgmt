package main

import (
	"encoding/json"
	"fmt"
	"github.com/datatogether/task-mgmt/tasks"
	"github.com/garyburd/redigo/redis"
)

// Main redis connection
var rpool *redis.Pool

func connectRedis() (err error) {
	// var netConn net.Conn

	if cfg.RedisUrl == "" {
		return fmt.Errorf("no redis url specified")
	}

	// for i := 0; i <= 1000; i++ {
	// 	netConn, err = net.Dial("tcp", cfg.RedisUrl)
	// 	if err != nil {
	// 		log.Infof("error connecting to redis: %s", err.Error())
	// 		time.Sleep(time.Second)
	// 		continue
	// 	}
	// 	break
	// }

	// if netConn == nil {
	// 	return fmt.Errorf("no net connection after 1000 tries")
	// }

	// rconn = redis.NewConn(netConn, time.Second*20, time.Second*20)

	rpool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", cfg.RedisUrl)
		return c, err
	}, 3)
	// defer rpool.Close()

	return nil
}

func PublishTaskProgress(pool *redis.Pool, t *tasks.Task) error {
	c := pool.Get()
	defer c.Close()

	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	// log.Infof("publishing on channel %s: %s", t.PubSubChannelName(), string(data))
	_, err = c.Do("PUBLISH", t.PubSubChannelName(), data)
	return err
}
