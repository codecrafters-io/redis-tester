package main

import (
	"bufio"
	"fmt"
	"net"
)

type RedisServer struct {
	resp    *RESPCodec
	cmdProc *CommandProcessor
	config  *Config
}

func NewRedisServer(config *Config) *RedisServer {
	resp := NewRESPCodec()
	cmdProc := NewCommandProcessor(resp)

	server := &RedisServer{
		resp:    resp,
		cmdProc: cmdProc,
		config:  config,
	}

	return server
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
		// read command
		command, err := s.resp.ReadCommand(reader)
		if err != nil {
			return
		}

		// Process the command
		response := s.cmdProc.ProcessCommand(command)

		if _, err := conn.Write(response); err != nil {
			fmt.Println("Error writing response:", err.Error())
			return
		}

	}
}
