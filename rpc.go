package main

import (
	"fmt"
	"github.com/datatogether/task-mgmt/tasks"
	"net"
	"net/rpc"
)

// if cfg.RpcPort is specified listenRpc opens up a
// Remote Procedure call listener to communicate with
// other servers
func listenRpc() error {
	if cfg.RpcPort == "" {
		log.Infoln("no rpc port specified, rpc disabled")
		return nil
	}

	taskRequests := &tasks.TaskRequests{
		AmqpUrl: cfg.AmqpUrl,
		Store:   store,
	}
	if err := rpc.Register(taskRequests); err != nil {
		log.Infof("register RPC Users error: %s", err)
		return err
	}
	// if err := rpc.Register(GroupsRequests); err != nil {
	// 	log.Infof("register RPC Groups error: %s", err)
	// 	return err
	// }

	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.RpcPort))
	if err != nil {
		log.Infof("listen on port %s error: %s", cfg.RpcPort, err)
		return err
	}

	log.Infof("accepting RPC requests on port %s", cfg.RpcPort)
	rpc.Accept(ln)
	return nil
}
