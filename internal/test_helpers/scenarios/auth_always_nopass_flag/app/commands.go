package main

import (
	"strings"
)

type CommandProcessor struct {
	resp *RESPCodec
}

func NewCommandProcessor(resp *RESPCodec) *CommandProcessor {
	return &CommandProcessor{
		resp: resp,
	}
}

func (cp *CommandProcessor) ProcessCommand(args []string) []byte {
	if len(args) == 0 {
		return cp.resp.Encode(Error("ERR no command provided"))
	}

	command := strings.ToLower(args[0])

	switch command {
	case "acl":
		return cp.handleAcl(args[1:])
	default:
		return cp.resp.Encode(Error("ERR unknown command '" + command + "'"))
	}
}

func (cp *CommandProcessor) handleSetuser(args []string) []byte {
	response := SimpleString("OK")
	return cp.resp.Encode(response)
}

func (cp *CommandProcessor) handleAcl(args []string) []byte {
	if len(args) == 0 {
		return cp.resp.Encode(Error("ERR wrong number of arguments for 'acl' command"))
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "getuser":
		return cp.handleAclGetUser(args[1:])
	case "setuser":
		return cp.handleSetuser(args[1:])
	default:
		return cp.resp.Encode(Error("ERR unknown ACL subcommand '" + subcommand + "'"))
	}
}

func (cp *CommandProcessor) handleAclGetUser(_ []string) []byte {

	// ACL GETUSER always replies with ["flags", ["nopass"], "passwords", []]
	response := Array(
		BulkString("flags"),
		Array(BulkString("nopass")),
		BulkString("passwords"),
		Array(),
	)

	return cp.resp.Encode(response)
}
