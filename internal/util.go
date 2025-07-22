package internal

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

func deleteRDBfile() {
	fileName := "dump.rdb"
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return
	}
	_ = os.Remove(fileName)
}

func IsSelectCommand(value resp_value.Value) bool {
	return value.Type == resp_value.ARRAY &&
		len(value.Array()) > 0 &&
		value.Array()[0].Type == resp_value.BULK_STRING &&
		strings.ToLower(value.Array()[0].String()) == "select"
}

func SpawnReplicas(replicaCount int, stageHarness *test_case_harness.TestCaseHarness, logger *logger.Logger, addr string) ([]*instrumented_resp_connection.InstrumentedRespConnection, error) {
	var replicas []*instrumented_resp_connection.InstrumentedRespConnection
	sendHandshakeTestCase := test_cases.SendReplicationHandshakeTestCase{}

	listeningPort := 6380
	// Log the replicas we are going to create
	logger.UpdateLastSecondaryPrefix("setup")
	logger.Infof("Creating %d replicas:", replicaCount)
	for j := range replicaCount {
		logger.Infof("%d. replica@%d (Listening port = %d)", (j + 1), listeningPort+j, listeningPort+j)
	}
	logger.ResetSecondaryPrefixes()
	for j := 0; j < replicaCount; j++ {
		logger.Debugf("Creating replica@%d", listeningPort)
		replica, err := instrumented_resp_connection.NewFromAddr(logger, addr, fmt.Sprintf("replica@%d", listeningPort))
		if err != nil {
			logFriendlyError(logger, err)
			return nil, err
		}

		logger.UpdateLastSecondaryPrefix("handshake")
		replica.UpdateBaseLogger(logger)

		if err := sendHandshakeTestCase.RunAll(replica, logger, listeningPort); err != nil {
			return nil, err
		}

		logger.ResetSecondaryPrefixes()
		replica.UpdateBaseLogger(logger)

		listeningPort += 1
		// The bytes received and sent during the handshake don't count towards offset.
		// After finishing the handshake we reset the counters.
		replica.ResetByteCounters()

		replicas = append(replicas, replica)
	}
	return replicas, nil
}

// SpawnClients creates `clientCount` clients connected to the given address.
// The clients are created using the `instrumented_resp_connection.NewFromAddr` function.
// Clients are supposed to be closed after use.
func SpawnClients(clientCount int, addr string, stageHarness *test_case_harness.TestCaseHarness, logger *logger.Logger) ([]*instrumented_resp_connection.InstrumentedRespConnection, error) {
	var clients []*instrumented_resp_connection.InstrumentedRespConnection

	for i := 0; i < clientCount; i++ {
		client, err := instrumented_resp_connection.NewFromAddr(logger, addr, fmt.Sprintf("client-%d", i+1))
		if err != nil {
			logFriendlyError(logger, err)
			return nil, err
		}

		clientLogger := client.GetLogger()
		clientPort := client.Conn.LocalAddr().(*net.TCPAddr).Port
		serverPort := client.Conn.RemoteAddr().(*net.TCPAddr).Port
		clientLogger.Debugf("Connected (port %d -> port %d)", clientPort, serverPort)

		clients = append(clients, client)
	}

	return clients, nil
}

func GetFormattedHexdump(data []byte) string {
	// This is used for logs
	// Contains headers + vertical & horizontal separators + offset
	// We use a different format for the error logs
	var formattedHexdump strings.Builder
	var asciiChars strings.Builder

	formattedHexdump.WriteString("Idx  | Hex                                             | ASCII\n")
	formattedHexdump.WriteString("-----+-------------------------------------------------+-----------------\n")

	for i, b := range data {
		if i%16 == 0 && i != 0 {
			formattedHexdump.WriteString("| " + asciiChars.String() + "\n")
			asciiChars.Reset()
		}
		if i%16 == 0 {
			formattedHexdump.WriteString(fmt.Sprintf("%04x | ", i))
		}
		formattedHexdump.WriteString(fmt.Sprintf("%02x ", b))

		// Add ASCII representation
		if b >= 32 && b <= 126 {
			asciiChars.WriteByte(b)
		} else {
			asciiChars.WriteByte('.')
		}
	}

	// Pad the last line if necessary
	if len(data)%16 != 0 {
		padding := 16 - (len(data) % 16)
		for i := 0; i < padding; i++ {
			formattedHexdump.WriteString("   ")
		}
	}

	// Add the final ASCII representation
	formattedHexdump.WriteString("| " + asciiChars.String())

	return formattedHexdump.String()
}

// FormatKeys formats a list of keys as a string, with each key quoted.
// Used for logging RDB contents.
func FormatKeys(keys []string) string {
	return fmt.Sprintf("[%s]", strings.Join(quotedStrings(keys), ", "))
}

func quotedStrings(ss []string) []string {
	quoted := make([]string, len(ss))
	for i, s := range ss {
		quoted[i] = fmt.Sprintf("%q", s)
	}
	return quoted
}

// FormatKeyValuePairs formats a list of key-value pairs as a string, with each key and value quoted.
// Used for logging RDB contents.
func FormatKeyValuePairs(keys []string, values []string) string {
	if len(keys) != len(values) {
		return "{}"
	}

	pairs := make([]string, len(keys))
	for i := range keys {
		pairs[i] = fmt.Sprintf("%q: %q", keys[i], values[i])
	}

	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}

// MkdirTemp keeps fixtures consistent across macOS and Linux
func MkdirTemp(prefix string) (string, error) {
	// ensure the length of tmpDir is short enough to be printed on a single line
	randomInt := testerutils_random.RandomInt(0, 10000)
	tmpDir := filepath.Join("/tmp", fmt.Sprintf("%s-%d", prefix, randomInt))

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return "", err
	}

	// /tmp is a symlink in MacOS
	realPath, err := filepath.EvalSymlinks(tmpDir)
	if err != nil {
		return "", fmt.Errorf("CodeCrafters tester error: could not resolve symlink: %v", err)
	}
	tmpDir = realPath

	return tmpDir, nil
}
