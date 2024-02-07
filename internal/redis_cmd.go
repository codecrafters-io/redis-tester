package internal

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var store = make(map[string]Value)

func Ping() []byte {
	data, err := NewSimpleStringValue("PONG").Encode()
	if err != nil {
		return SendError(err)
	}
	return data
}

func Echo(v Value) []byte {
	data, err := v.Encode()
	if err != nil {
		return SendError(err)
	}
	return data
}

func Set(args []Value) []byte {
	var k, v Value
	var opt string
	var expiry int
	var err error

	if len(args) >= 2 {
		k = args[0]
		v = args[1]
	}

	if len(args) == 4 {
		opt = args[2].String()
		expiry, err = args[3].Integer()
		if err != nil {
			return SendError(err)
		}
	}

	var response []byte

	if old, ok := store[k.String()]; ok {
		response, err = NewBulkStringValue(old.String()).Encode()
		if err != nil {
			return SendError(err)
		}
	} else {
		response, err = NewSimpleStringValue("OK").Encode()
		if err != nil {
			return SendError(err)
		}
	}
	if v.Type() != BULK_STRING {
		v = NewBulkStringValue(v.String())
	}

	store[k.String()] = v

	if strings.ToUpper(opt) == "PX" && expiry > 0 {
		go func() {
			ch := time.After(time.Duration(expiry) * time.Millisecond)
			for {
				select {
				case <-ch:
					delete(store, k.String())
				}
			}
		}()
	}
	return response
}

func Get(k Value) []byte {
	if data, ok := store[k.String()]; ok {
		bytes, err := data.Encode()
		if err != nil {
			return SendError(err)
		}
		return bytes
	}
	return SendNil()
}

func Info(v Value, replicaMode bool) []byte {
	_, err := v.Encode() // Replication header
	if err != nil {
		return SendError(err)
	}
	var role string

	if replicaMode == true {
		role = "slave"
	} else {
		role = "master"
	}

	var info string
	info += "replication" + "\n"
	info += "role:" + role + "\n"
	info += "master_replid:" + "foo" + "\n"
	info += "master_repl_offset:" + "0"
	data, err := NewBulkStringValue(info).Encode()
	if err != nil {
		return SendError(err)
	}
	return data
}

func Replconf(args []Value) []byte {
	data, err := NewSimpleStringValue("OK").Encode()
	if err != nil {
		return SendError(err)
	}
	return data
}

func Psync(args []Value) ([]byte, bool) {
	data, err := NewSimpleStringValue("FULLRESYNC c00d0def8c1d916ed06e2d2e69b8b658532a07ef 0").Encode()
	if err != nil {
		return SendError(err), false
	}
	return data, true
}

func SendRDBFile() []byte {
	hexStr := "524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2"
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		fmt.Printf("Encountered %s while deconding hex string", err.Error())
		return SendError(err)
	}
	resp := []byte("$" + strconv.Itoa(len(bytes)) + "\r\n")
	return (append(resp, bytes...))
}
