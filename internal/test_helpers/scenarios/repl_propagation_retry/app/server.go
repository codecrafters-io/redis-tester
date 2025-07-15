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
	repl    *ReplicationManager
	config  *Config
}

func NewRedisServer(config *Config) *RedisServer {
	store := NewStore()
	resp := NewRESPCodec()
	repl := NewReplicationManager()
	cmdProc := NewCommandProcessor(store, resp, repl)

	server := &RedisServer{
		store:   store,
		resp:    resp,
		cmdProc: cmdProc,
		repl:    repl,
		config:  config,
	}

	// handle --replicaof
	if config.ReplicaOf != nil {
		repl.SetReplicaMode(config.ReplicaOf)
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

	if s.config.ReplicaOf != nil {
		go s.startReplication()
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err.Error())
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *RedisServer) startReplication() {
	replicaClient := NewReplicaClient(s.config.ReplicaOf, s.resp, s.config.Port)

	if err := replicaClient.Connect(); err != nil {
		fmt.Println("Failed to connect to master:", err.Error())
		return
	}
	defer replicaClient.Close()

	replicaClient.StartReplicating(s.cmdProc)
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

		if s.repl.GetRole() == "master" && len(command) > 0 && strings.ToLower(command[0]) == "psync" {
			s.repl.AddReplica(conn)
			defer s.repl.RemoveReplica(conn)
		}
	}
}
