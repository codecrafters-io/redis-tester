package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type RedisServer struct {
	store   *Store
	resp    *RESPCodec
	cmdProc *CommandProcessor
	config  *Config
	blocker *BlockingManager
}

func NewRedisServer(config *Config) *RedisServer {
	store := NewStore()
	resp := NewRESPCodec()
	blocker := NewBlockingManager()
	cmdProc := NewCommandProcessor(store, resp, blocker)

	return &RedisServer{
		store:   store,
		resp:    resp,
		cmdProc: cmdProc,
		config:  config,
		blocker: blocker,
	}
}

func (s *RedisServer) Start() error {
	addr := fmt.Sprintf("0.0.0.0:%d", s.config.Port)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind to port %d: %v", s.config.Port, err)
	}

	defer listener.Close()

	fmt.Println("Redis server listening on", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *RedisServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		command, err := s.resp.ReadCommand(reader)
		if err != nil {
			return
		}
		response := s.cmdProc.ProcessCommand(command, conn)
		if response != nil {
			if _, err := conn.Write(response); err != nil {
				return
			}
		}
		if len(command) > 0 && strings.ToLower(command[0]) == "rpush" {
			s.blocker.NotifyAllWaiters(s.store, s.resp)
		}
	}
}
