package main

import (
	"strings"
)

type CommandProcessor struct {
	resp   *RESPCodec
	server *RedisServer
}

func NewCommandProcessor(resp *RESPCodec) *CommandProcessor {
	return &CommandProcessor{
		resp: resp,
	}
}

func (cp *CommandProcessor) ProcessCommand(args []string, connIsDefault bool) []byte {
	if len(args) == 0 {
		return cp.resp.Encode(Error("ERR no command provided"))
	}

	command := strings.ToLower(args[0])

	switch command {
	case "ping":
		return cp.resp.Encode(SimpleString("PONG"))
	case "acl":
		return cp.handleAcl(args[1:], connIsDefault)
	default:
		return cp.resp.Encode(Error("ERR unknown command '" + command + "'"))
	}
}

func (cp *CommandProcessor) handleAcl(args []string, connIsDefault bool) []byte {
	if len(args) == 0 {
		return cp.resp.Encode(Error("ERR wrong number of arguments for 'acl' command"))
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "setuser":
		return cp.handleSetuser(args[1:])
	case "whoami":
		return cp.handleWhoami(connIsDefault)
	default:
		return cp.resp.Encode(Error("ERR unknown ACL subcommand '" + subcommand + "'"))
	}
}

func (cp *CommandProcessor) handleSetuser(_ []string) []byte {
	if cp.server != nil {
		cp.server.defaultUserHasPassword = true
	}
	return cp.resp.Encode(SimpleString("OK"))
}

func (cp *CommandProcessor) handleWhoami(connIsDefault bool) []byte {
	if connIsDefault {
		return cp.resp.Encode(BulkString("default"))
	}
	// Intentional bug: should be "NOAUTH Authentication required." but we use "ERR" so
	// PatternedBytesAssertion (expected prefix "NOAUTH") will fail and produce a fixture.
	return cp.resp.Encode(Error("ERR Authentication required."))
}
