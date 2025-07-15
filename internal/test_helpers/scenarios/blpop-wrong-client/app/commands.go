package main

import (
	"net"
	"strconv"
	"strings"
	"sync"
)

type CommandProcessor struct {
	store          *Store
	resp           *RESPCodec
	blocker        *BlockingManager
	blockedClients int
	mu             sync.Mutex
}

func NewCommandProcessor(store *Store, resp *RESPCodec, blocker *BlockingManager) *CommandProcessor {
	return &CommandProcessor{
		store:   store,
		resp:    resp,
		blocker: blocker,
	}
}

func (cp *CommandProcessor) ProcessCommand(args []string, conn net.Conn) []byte {
	if len(args) == 0 {
		return cp.resp.EncodeError("ERR no command provided")
	}

	command := strings.ToLower(args[0])

	switch command {
	case "ping":
		return cp.handlePing(args[1:])
	case "echo":
		return cp.handleEcho(args[1:])
	case "rpush":
		return cp.handleRPush(args[1:])
	case "blpop":
		return cp.handleBLPop(args[1:], conn)
	default:
		return cp.resp.EncodeError("ERR unknown command '" + command + "'")
	}
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

func (cp *CommandProcessor) handleRPush(args []string) []byte {
	if len(args) < 2 {
		return cp.resp.EncodeError("ERR wrong number of arguments for 'rpush' command")
	}

	key := args[0]
	values := args[1:]

	length := cp.store.RPush(key, values...)
	return cp.resp.EncodeInteger(length)
}

func (cp *CommandProcessor) handleBLPop(args []string, conn net.Conn) []byte {
	if len(args) < 2 {
		return cp.resp.EncodeError("ERR wrong number of arguments for 'blpop' command")
	}

	key := args[0]
	timeoutStr := args[1]

	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		return cp.resp.EncodeError("ERR timeout is not an integer or out of range")
	}

	if timeout != 0 {
		return cp.resp.EncodeError("ERR timeout not supported yet")
	}

	value := cp.store.LPop(key)
	if value != nil {
		return cp.resp.EncodeArray([]string{key, *value})
	}

	// Always ignore the first client: Bug introduced to send response to wrong client on stage ec3
	cp.mu.Lock()
	if cp.blockedClients == 0 {
		cp.blockedClients += 1
	} else {
		cp.blocker.WaitForElement(key, conn)
	}
	cp.mu.Unlock()
	return nil
}
