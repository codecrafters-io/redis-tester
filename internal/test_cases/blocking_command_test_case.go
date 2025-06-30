package test_cases

import (
	"time"

	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type BlockingCommandTestCase struct {
	SendCommandTestCase
	timeoutDuration time.Duration
	resumeChannel   chan struct{}
	resultChan      chan error
}

func NewBlockingCommandTestCase(command string, args []string, assertion resp_assertions.RESPAssertion, timeout *time.Duration) *BlockingCommandTestCase {
	var t time.Duration
	if timeout == nil {
		t = time.Second * 10
	} else {
		t = *timeout
	}
	return &BlockingCommandTestCase{
		SendCommandTestCase: SendCommandTestCase{
			Command:                   command,
			Args:                      args,
			Assertion:                 assertion,
			ShouldSkipUnreadDataCheck: true,
		},
		resumeChannel:   make(chan struct{}, 1),
		resultChan:      make(chan error, 1),
		timeoutDuration: t,
	}
}

func (t *BlockingCommandTestCase) Run(client *resp_client.RespConnection, logger *logger.Logger) {
	go func() {
		responseChannel := make(chan error)
		timeoutChan := time.After(t.timeoutDuration)
		go func() {
			err := t.SendCommandTestCase.Run(client, logger)
			responseChannel <- err
		}()

		select {
		case <-t.resumeChannel:
			t.resultChan <- <-responseChannel
		case <-timeoutChan:
			t.resultChan <- <-responseChannel
		}
	}()
}

func (t *BlockingCommandTestCase) WaitForResult() error {
	return <-t.resultChan
}

func (t *BlockingCommandTestCase) Resume() {
	select {
	case t.resumeChannel <- struct{}{}:
	default:
		panic("Codecrafters Internal Error - blocking test case already resumed")
	}
}
