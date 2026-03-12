package internal

import (
	"fmt"
	"net"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

// ClientsSpawner automatically takes care of logging friendly error and registering
// client.Close() as a teardown function while spawning multiple clients
// It spawns n client on each SpawnClients() call with prefix "clients-n" style prefix
type ClientsSpawner struct {
	Addr         string
	StageHarness *test_case_harness.TestCaseHarness
	Logger       *logger.Logger

	// internal state to keep track of how many clients have been spawned so far
	clientsSpawned int
}

func (s *ClientsSpawner) SpawnClients(clientsCount int) ([]*instrumented_resp_connection.InstrumentedRespConnection, error) {
	var clients []*instrumented_resp_connection.InstrumentedRespConnection

	for range clientsCount {
		client, err := instrumented_resp_connection.NewFromAddr(s.Logger, s.Addr, fmt.Sprintf("client-%d", s.clientsSpawned+1))

		if err != nil {
			logFriendlyError(s.Logger, err)
			return nil, err
		}

		s.clientsSpawned += 1

		// Auto-close the client on teardown
		s.StageHarness.RegisterTeardownFunc(func() {
			client.Close()
		})

		clientLogger := client.GetLogger()
		clientPort := client.Conn.LocalAddr().(*net.TCPAddr).Port
		serverPort := client.Conn.RemoteAddr().(*net.TCPAddr).Port
		clientLogger.Debugf("Connected (port %d -> port %d)", clientPort, serverPort)

		clients = append(clients, client)
	}

	return clients, nil
}
