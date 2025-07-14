package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type ReplicaClient struct {
	config     *ReplicaConfig
	conn       net.Conn
	resp       *RESPCodec
	serverPort int
}

func NewReplicaClient(config *ReplicaConfig, resp *RESPCodec, serverPort int) *ReplicaClient {
	return &ReplicaClient{
		config:     config,
		resp:       resp,
		serverPort: serverPort,
	}
}

func (rc *ReplicaClient) Connect() error {
	addr := net.JoinHostPort(rc.config.Host, strconv.Itoa(rc.config.Port))
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to master %s: %v", addr, err)
	}

	rc.conn = conn
	return nil
}

func (rc *ReplicaClient) StartReplicating(s *CommandProcessor) {
	reader := bufio.NewReader(rc.conn)

	/* Handshake first */
	// step 1: Send PING
	if err := rc.sendPing(); err != nil {
		fmt.Println("Ping failed")
		return
	}

	response, err := rc.readResponse(reader)
	if err != nil {
		fmt.Printf("failed to read PONG: %v\n", err)
		return
	}

	if string(response) != "+PONG\r\n" {
		fmt.Printf("expected PONG, got: %s\n", string(response))
		return
	}

	// send REPLCONF listening-port
	if err := rc.sendReplConfListeningPort(); err != nil {
		fmt.Printf("REPLCONF listening-port failed: %v\n", err)
		return
	}

	response, err = rc.readResponse(reader)
	if err != nil {
		fmt.Printf("failed to read REPLCONF listening-port response: %v\n", err)
		return
	}

	if string(response) != "+OK\r\n" {
		fmt.Printf("expected OK for REPLCONF listening-port, got: %s\n", string(response))
		return
	}

	// send REPLCONF capa psync2
	if err := rc.sendReplConfCapa(); err != nil {
		fmt.Printf("REPLCONF capa failed: %v\n", err)
		return
	}

	response, err = rc.readResponse(reader)
	if err != nil {
		fmt.Printf("failed to read REPLCONF capa response: %s\n", err)
		return
	}

	if string(response) != "+OK\r\n" {
		fmt.Printf("expected OK for REPLCONF capa, got: %s\n", string(response))
		return
	}

	// send PSYNC
	if err := rc.sendPsync(); err != nil {
		fmt.Printf("PSYNC failed: %v\n", err)
		return
	}

	// read FULLRESYNC response and RDB file
	_, err = rc.readFullResyncAndRDB(reader)
	if err != nil {
		fmt.Printf("failed to read PSYNC response: %v\n", err)
		return
	}

	rc.ProcessPropagatedCommands(reader, s)
}

func (rc *ReplicaClient) sendPing() error {
	pingCmd := rc.resp.EncodeArray([]string{"PING"})
	_, err := rc.conn.Write(pingCmd)
	return err
}

func (rc *ReplicaClient) sendReplConfListeningPort() error {
	cmd := rc.resp.EncodeArray([]string{"REPLCONF", "listening-port", strconv.Itoa(rc.serverPort)})
	_, err := rc.conn.Write(cmd)
	return err
}

func (rc *ReplicaClient) sendReplConfCapa() error {
	cmd := rc.resp.EncodeArray([]string{"REPLCONF", "capa", "psync2"})
	_, err := rc.conn.Write(cmd)
	return err
}

func (rc *ReplicaClient) sendPsync() error {
	cmd := rc.resp.EncodeArray([]string{"PSYNC", "?", "-1"})
	_, err := rc.conn.Write(cmd)
	return err
}

func (rc *ReplicaClient) readResponse(reader *bufio.Reader) ([]byte, error) {
	var response []byte
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err != nil {
			return nil, err
		}

		response = append(response, buffer[:n]...)

		// Check if we have a complete response
		if len(response) >= 4 && string(response[len(response)-2:]) == "\r\n" {
			break
		}
	}

	return response, nil
}

// readFullResyncAndRDB reads the FULLRESYNC response and the RDB file
// it stops reading after the RDB file, leaving any subsequent commands in the buffer
func (rc *ReplicaClient) readFullResyncAndRDB(reader *bufio.Reader) ([]byte, error) {
	// consume FULLRESYNC response
	_, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading FULLRESYNC: %w", err)
	}

	rdbFileLen, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading RDB file length: %w", err)
	}

	fileLen, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(string(rdbFileLen), "$"), "\r\n"))
	if err != nil {
		return nil, fmt.Errorf("error parsing RDB file length: %w", err)
	}

	rdbFileBuf := make([]byte, fileLen)
	_, err = io.ReadFull(reader, rdbFileBuf)
	if err != nil {
		return nil, fmt.Errorf("error reading RDB File: %w", err)
	}

	return rdbFileBuf, nil
}

func (rc *ReplicaClient) Close() error {
	if rc.conn != nil {
		return rc.conn.Close()
	}
	return nil
}

func (rc *ReplicaClient) GetConnection() net.Conn {
	return rc.conn
}

// ProcessPropagatedCommands reads and processes commands from the master
func (rc *ReplicaClient) ProcessPropagatedCommands(reader *bufio.Reader, cmdProc *CommandProcessor) {
	for {
		// Read command from master
		command, err := rc.resp.ReadCommand(reader)
		if err != nil {
			return
		}

		// Process the command silently (no response sent back to master)
		cmdProc.ProcessCommand(command)
	}
}
