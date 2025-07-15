package main

import (
	"fmt"
	"strconv"
	"strings"
)

type CommandProcessor struct {
	store *Store
	resp  *RESPCodec
	repl  *ReplicationManager
}

func NewCommandProcessor(store *Store, resp *RESPCodec, repl *ReplicationManager) *CommandProcessor {
	return &CommandProcessor{
		store: store,
		resp:  resp,
		repl:  repl,
	}
}

func (cp *CommandProcessor) ProcessCommand(args []string) []byte {
	if len(args) == 0 {
		return cp.resp.EncodeError("ERR no command provided")
	}

	command := strings.ToLower(args[0])

	var response []byte

	switch command {
	case "ping":
		response = cp.handlePing(args[1:])
	case "echo":
		response = cp.handleEcho(args[1:])
	case "set":
		response = cp.handleSet(args[1:])
	case "get":
		response = cp.handleGet(args[1:])
	case "info":
		response = cp.handleInfo(args[1:])
	case "replconf":
		response = cp.handleReplConf(args[1:])
	case "psync":
		response = cp.handlePsync(args[1:])
	default:
		response = cp.resp.EncodeError("ERR unknown command '" + command + "'")
	}

	// only propagate write commands to replicas
	if cp.isWriteCommand(command) {
		cp.propagateToReplicas(args)
	}

	return response
}

func (cp *CommandProcessor) handlePing(_ []string) []byte {
	return cp.resp.EncodeSimpleString("PONG")
}

func (cp *CommandProcessor) handleEcho(args []string) []byte {
	if len(args) == 0 {
		return cp.resp.EncodeError("ERR wrong number of arguments for 'echo' command")
	}
	return cp.resp.EncodeBulkString(args[0])
}

func (cp *CommandProcessor) handleSet(args []string) []byte {
	if len(args) < 2 {
		return cp.resp.EncodeError("ERR wrong number of arguments for 'set' command")
	}

	key := args[0]
	value := args[1]

	// Check for expiry options
	var expiryMs *int
	if len(args) >= 4 && strings.ToLower(args[2]) == "px" {
		expiry, err := strconv.Atoi(args[3])
		if err != nil {
			return cp.resp.EncodeError("ERR value is not an integer or out of range")
		}
		expiryMs = &expiry
	}

	cp.store.Set(key, value, expiryMs)

	return cp.resp.EncodeSimpleString("OK")
}

func (cp *CommandProcessor) handleGet(args []string) []byte {
	if len(args) != 1 {
		return cp.resp.EncodeError("ERR wrong number of arguments for 'get' command")
	}

	key := args[0]
	value := cp.store.Get(key)

	if value == nil {
		return cp.resp.EncodeNil()
	}

	return cp.resp.EncodeBulkString(*value)
}

func (cp *CommandProcessor) handleInfo(args []string) []byte {
	if len(args) == 0 {
		return cp.resp.EncodeError("ERR wrong number of arguments for 'info' command")
	}

	section := strings.ToLower(args[0])
	if section != "replication" {
		return cp.resp.EncodeError("ERR unsupported info section")
	}

	role := cp.repl.GetRole()
	replID := cp.repl.GetReplicationID()
	replOffset := cp.repl.GetReplicationOffset()

	info := fmt.Sprintf("role:%s\r\nmaster_replid:%s\r\nmaster_repl_offset:%d\r\n",
		role, replID, replOffset)

	return cp.resp.EncodeBulkString(info)
}

func (cp *CommandProcessor) handleReplConf(args []string) []byte {
	if len(args) < 1 {
		return cp.resp.EncodeError("ERR wrong number of arguments for 'replconf' command")
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "listening-port":
		return cp.resp.EncodeSimpleString("OK")
	case "capa":
		return cp.resp.EncodeSimpleString("OK")
	default:
		return cp.resp.EncodeError("ERR unsupported REPLCONF subcommand")
	}
}

func (cp *CommandProcessor) handlePsync(args []string) []byte {
	if len(args) != 2 {
		return cp.resp.EncodeError("ERR wrong number of arguments for 'psync' command")
	}

	replID := args[0]
	offset := args[1]

	if replID != "?" || offset != "-1" {
		return cp.resp.EncodeError("ERR only full resync is supported")
	}

	response := fmt.Sprintf("FULLRESYNC %s %d", cp.repl.GetReplicationID(), cp.repl.GetReplicationOffset())

	fullResync := cp.resp.EncodeSimpleString(response)

	// hardcoded RDB file
	emptyRdbHex := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"

	emptyRdbBytes := make([]byte, len(emptyRdbHex)/2)
	for i := 0; i < len(emptyRdbHex); i += 2 {
		val, _ := strconv.ParseUint(emptyRdbHex[i:i+2], 16, 8)
		emptyRdbBytes[i/2] = byte(val)
	}

	rdbLength := len(emptyRdbBytes)
	rdbHeader := fmt.Sprintf("$%d\r\n", rdbLength)

	result := append(fullResync, []byte(rdbHeader)...)
	result = append(result, emptyRdbBytes...)

	return result
}

func (cp *CommandProcessor) isWriteCommand(command string) bool {
	writeCommands := map[string]bool{
		"set":     true,
		"del":     true,
		"incr":    true,
		"decr":    true,
		"lpush":   true,
		"rpush":   true,
		"lpop":    true,
		"rpop":    true,
		"xadd":    true,
		"expire":  true,
		"pexpire": true,
	}
	return writeCommands[command]
}

func (cp *CommandProcessor) propagateToReplicas(args []string) {
	if cp.repl.GetRole() != "master" {
		return
	}

	encodedCommand := cp.resp.EncodeArray(args)

	replicas := cp.repl.GetReplicas()

	for _, replica := range replicas {
		replica.Conn.Write(encodedCommand)
	}
}
