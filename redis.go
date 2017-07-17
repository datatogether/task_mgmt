package main

import (
	"encoding/json"
	"fmt"
	"github.com/datatogether/task-mgmt/tasks"
	"github.com/garyburd/redigo/redis"
	"net"
	"time"
)

// Main redis connection
var rconn redis.Conn

func connectRedis() (err error) {
	var netConn net.Conn

	if cfg.RedisUrl == "" {
		return fmt.Errorf("no redis url specified")
	}

	for i := 0; i <= 1000; i++ {
		netConn, err = net.Dial("tcp", cfg.RedisUrl)
		if err != nil {
			return err
			time.Sleep(time.Second)
			continue
		}
		break
	}

	if netConn == nil {
		return fmt.Errorf("no net connection after 1000 tries")
	}

	rconn = redis.NewConn(netConn, time.Second*20, time.Second*20)
	return nil
}

func PublishProgress(c redis.Conn, p tasks.Progress) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	_, err = c.Do("PUBLISH", p.ChannelName(), data)
	return err
}
