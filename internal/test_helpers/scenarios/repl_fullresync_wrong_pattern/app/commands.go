package main

import (
	"strings"
)

type CommandProcessor struct {
	resp *RESPCodec
}

func NewCommandProcessor(resp *RESPCodec) *CommandProcessor {
	return &CommandProcessor{resp: resp}
}

func (cp *CommandProcessor) ProcessCommand(args []string) []byte {
	if len(args) == 0 {
		return cp.resp.EncodeError("ERR no command provided")
	}

	command := strings.ToLower(args[0])

	switch command {
	case "ping":
		return cp.handlePing()
	case "replconf":
		return cp.handleReplConf(args[1:])
	case "psync":
		return cp.handlePsync(args[1:])
	default:
		return cp.resp.EncodeError("ERR unknown command '" + command + "'")
	}
}

func (cp *CommandProcessor) handlePing() []byte {
	return cp.resp.EncodeSimpleString("PONG")
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

// handlePsync returns a wrong-format response on purpose so vm3 fails with RegexAssertion.
// Expected pattern is FULLRESYNC [A-Za-z0-9]{40} 0; we send "FULLRESYNC wrong 0" (repl ID not 40 chars).
func (cp *CommandProcessor) handlePsync(args []string) []byte {
	if len(args) != 2 {
		return cp.resp.EncodeError("ERR wrong number of arguments for 'psync' command")
	}

	if args[0] != "?" || args[1] != "-1" {
		return cp.resp.EncodeError("ERR only full resync is supported")
	}

	// Intentionally wrong: replication ID must be 40 alphanumeric chars; "wrong" is 5 chars.
	return cp.resp.EncodeSimpleString("FULLRESYNC 4d3e1bc3e18c 0")
}
