package internal

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/smallnest/resp3"
)

var replicaMode bool = false
var masterAddr string = ""
var replicas []net.Conn
var propagationEnabled bool = false
var commandChannel chan []byte

func handleConnection(conn net.Conn, reader *resp3.Reader, writer *resp3.Writer) {
	defer conn.Close()

	for {
		req, err := Decode(reader)
		if err != nil {
			if err == io.EOF {
				continue
			}
			writer.Write(SendError(err))
			writer.Flush()
			continue
		}

		if len(req.Array()) == 0 {
			writer.Write(SendError(fmt.Errorf("expected command to be RESP Array")))
			writer.Flush()
			continue
		}

		// input is a valid RESP arrray
		command := req.Array()[0].String()
		args := req.Array()[1:]
		for _, v := range req.Array() {
			fmt.Println(v.String())
		}

		var response []byte
		var fullResyncRequired bool

		switch strings.ToUpper(command) {
		case "PING":
			response = Ping()
		case "ECHO":
			response = Echo(args[0])
		case "SET":
			response = Set(args)
		case "GET":
			response = Get(args[0])
		case "INFO":
			response = Info(args[0], replicaMode)
		case "REPLCONF":
			response = Replconf(args)
		case "PSYNC":
			response, fullResyncRequired = Psync(args)
		}

		writer.Write(response)
		writer.Flush()
		if fullResyncRequired == true {
			// This handler is serving as Master
			response = SendRDBFile()
			writer.Write(response)
			writer.Flush()
			fullResyncRequired = false
			go pushToReplica(commandChannel, writer)
			propagationEnabled = true
			fmt.Println("Started new goroutine for pushToReplica.")
			// replicas = append(replicas, ())
		}

		if propagationEnabled == true {
			// fmt.Println("Inside chan")
			request := recreateRESPMessage(req)
			// fmt.Println(string(request))
			commandChannel <- request
			fmt.Println("Recreated RESP message, and sent to channel CommandChannel.")
		}
	}
}

func connectToMaster(masterAddr string) (*resp3.Reader, *resp3.Writer, error) {
	conn, err := net.Dial("tcp", masterAddr)
	if err != nil {
		return nil, nil, err
	}
	reader := resp3.NewReader(conn)
	writer := resp3.NewWriter(conn)
	return reader, writer, nil
}

func sendMessage(writer *resp3.Writer, message string) error {
	if _, err := writer.WriteString(message); err != nil {
		return err
	}
	writer.Flush()
	return nil
}
func sendPing(writer *resp3.Writer) error {
	arr := make([]Value, 1)
	arr[0] = NewBulkStringValue("PING")
	v, _ := NewArrayValue(arr).Encode()
	return sendMessage(writer, string(v))
}
func sendReplconfPort(writer *resp3.Writer) error {
	arr := make([]Value, 3)
	arr[0] = NewBulkStringValue("REPLCONF")
	arr[1] = NewBulkStringValue("listening-port")
	arr[2] = NewBulkStringValue("6380")
	v, _ := NewArrayValue(arr).Encode()
	return sendMessage(writer, string(v))
}
func sendReplconfCapa(writer *resp3.Writer) error {
	arr := make([]Value, 5)
	arr[0] = NewBulkStringValue("REPLCONF")
	arr[1] = NewBulkStringValue("capa")
	arr[2] = NewBulkStringValue("eof")
	arr[3] = NewBulkStringValue("capa")
	arr[4] = NewBulkStringValue("eof")
	v, _ := NewArrayValue(arr).Encode()
	return sendMessage(writer, string(v))
}
func sendPsync(writer *resp3.Writer) error {
	arr := make([]Value, 3)
	arr[0] = NewBulkStringValue("PSYNC")
	arr[1] = NewBulkStringValue("?")
	arr[2] = NewBulkStringValue("-1")
	v, _ := NewArrayValue(arr).Encode()
	return sendMessage(writer, string(v))
}

func performSyncHandshakeFromReplica(reader *resp3.Reader, writer *resp3.Writer) error {
	err := sendPing(writer)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	fmt.Printf("Response from master: %s", response)

	err = sendReplconfPort(writer)
	response, err = reader.ReadString('\n')
	if err != nil {
		return err
	}
	fmt.Printf("Response from master: %s", response)

	err = sendReplconfCapa(writer)
	response, err = reader.ReadString('\n')
	if err != nil {
		return err
	}
	fmt.Printf("Response from master: %s", response)

	err = sendPsync(writer)
	response, err = reader.ReadString('\n')
	if err != nil {
		return err
	}
	fmt.Printf("Response from master: %s", response)

	return nil
}

func replicate(masterAddr string) {
	fmt.Println("Starting replication")
	reader, writer, err := connectToMaster(masterAddr)
	if err != nil {
		fmt.Printf("Failed to connect to master: %s\n", err)
		return
	}
	if err := performSyncHandshakeFromReplica(reader, writer); err != nil {
		fmt.Printf("Sync handshake failed: %s\n", err)
		return
	}
}

func NewRedisServer() {
	port := flag.String("port", "6379", "Port number for the server")
	var replicaHost, replicaPort string
	replicaOfFlag := flag.Bool("replicaof", false, "Use following two arguments for replica host and port")

	flag.Parse()

	// Manual parsing if --replicaof was provided
	if *replicaOfFlag {
		args := flag.Args() // Get non-flag command-line arguments
		replicaHost, replicaPort = args[0], args[1]
	}

	if replicaHost != "" && replicaPort != "" {
		fmt.Printf("Configuring as replica of: %s, Port: %s\n", replicaHost, replicaPort)
		replicaMode = true
		masterAddr = replicaHost + ":" + replicaPort
		go replicate(masterAddr)
	}

	addr := "0.0.0.0:" + *port
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		reader := resp3.NewReader(conn)
		writer := resp3.NewWriter(conn)
		go handleConnection(conn, reader, writer)
	}
}

func main() {
	commandChannel = make(chan []byte)
	NewRedisServer()
}

func recreateRESPMessage(req Value) []byte {
	n := len(req.Array())
	arr := make([]Value, n)
	for i, v := range req.Array() {
		arr[i] = NewBulkStringValue(v.String())
	}
	v, _ := NewArrayValue(arr).Encode()
	return v
}

// pushToReplica listens for commands on the channel and pushes them to the replica
func pushToReplica(commandChannel <-chan []byte, writer *resp3.Writer) {
	for command := range commandChannel {
		writer.Write(command)
		writer.Flush()
		// fmt.Printf("Pushing command to replica: %s\n", command)
	}
}
