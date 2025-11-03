package main

import (
	"strings"
)

type CommandProcessor struct {
	resp      *RESPCodec
	userStore *UserStore
}

func NewCommandProcessor(resp *RESPCodec) *CommandProcessor {
	return &CommandProcessor{
		resp:      resp,
		userStore: NewUserStore(),
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

func (cp *CommandProcessor) handleAcl(args []string) []byte {
	if len(args) == 0 {
		return cp.resp.Encode(Error("ERR wrong number of arguments for 'acl' command"))
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "getuser":
		return cp.handleAclGetUser(args[1:])
	case "setuser":
		return cp.handleAclSetUser(args[1:])
	default:
		return cp.resp.Encode(Error("ERR unknown ACL subcommand '" + subcommand + "'"))
	}
}

func (cp *CommandProcessor) handleAclGetUser(args []string) []byte {
	if len(args) == 0 {
		return cp.resp.Encode(Error("ERR wrong number of arguments for 'acl getuser' command"))
	}

	username := args[0]
	user, exists := cp.userStore.GetUser(username)

	var flagsArray []RESPValue
	if !exists || user.HasNopass() {
		flagsArray = []RESPValue{BulkString("nopass")}
	} else {
		flagsArray = []RESPValue{}
	}

	var passwordsArray []RESPValue
	if exists {
		passwords := user.GetPasswords()
		if len(passwords) > 0 {
			passwordsArray = make([]RESPValue, len(passwords))
			for i, password := range passwords {
				passwordsArray[i] = BulkString(password)
			}
		}
	}

	response := Array(
		BulkString("flags"),
		Array(flagsArray...),
		BulkString("passwords"),
		Array(passwordsArray...),
	)

	return cp.resp.Encode(response)
}

func (cp *CommandProcessor) handleAclSetUser(args []string) []byte {
	if len(args) == 0 {
		return cp.resp.Encode(Error("ERR wrong number of arguments for 'acl setuser' command"))
	}

	username := args[0]
	user := cp.userStore.GetOrCreateUser(username)

	// Process remaining arguments
	for i := 1; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, ">") {
			// Add password (with SHA1 hash)
			password := arg[1:]
			user.AddPassword(password)
		}
		// Ignore other ACL rules for now
	}

	return cp.resp.Encode(SimpleString("OK"))
}
