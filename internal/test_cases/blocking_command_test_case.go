package test_cases

import (
	"time"

	resp_client "github.com/codecrafters-io/redis-tester/internal/resp/connection"
	"github.com/codecrafters-io/tester-utils/logger"
)

type BlockingCommandTestCase struct {
	*SendCommandTestCase
	timeoutDuration time.Duration
	resumeChannel   chan struct{}
	resultChan      chan error
}

/*
BlockingCommandTestCase should be used whenever the client issues a command that is blocking (eg. BLPOP)

Flow:
Create using NewBlockingCommandTestCase()
Run()

If the command should respond after timeout, call WaitForResult()
If it should respond after another command has been executed, call ResumeAndWaitForResult()
*/

func NewBlockingCommandTestCase(SendCommandTestCase *SendCommandTestCase, timeout *time.Duration) *BlockingCommandTestCase {
	var t time.Duration
	if timeout == nil {
		t = time.Second * 10
	} else {
		t = *timeout
	}
	return &BlockingCommandTestCase{
		SendCommandTestCase: SendCommandTestCase,
		resumeChannel:       make(chan struct{}, 1),
		resultChan:          make(chan error, 1),
		timeoutDuration:     t,
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
	// maintains order for logs in fixtures
	select {
	case t.resumeChannel <- struct{}{}:
	default:
		panic("Codecrafters Internal Error - blocking test case already resumed")
	}
}

func (t *BlockingCommandTestCase) ResumeAndWaitForResult() error {
	t.Resume()
	return t.WaitForResult()
}
