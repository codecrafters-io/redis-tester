package internal

import (
	tester_utils "github.com/codecrafters-io/tester-utils"
	"regexp"
	"testing"
)

func TestStages(t *testing.T) {
	testCases := map[string]tester_utils.TesterOutputTestCase{
		"bind_failure": {
			StageName:           "init",
			CodePath:            "./test_helpers/scenarios/bind/failure",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/bind/failure",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"bind_timeout": {
			StageName:           "init",
			CodePath:            "./test_helpers/scenarios/bind/timeout",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/bind/timeout",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"bind_success": {
			StageName:           "init",
			CodePath:            "./test_helpers/scenarios/bind/success",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/bind/success",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"ping_pong_eof": {
			StageName:           "ping-pong",
			CodePath:            "./test_helpers/scenarios/ping-pong/eof",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/eof",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"ping_pong_without_crlf": {
			StageName:           "ping-pong",
			CodePath:            "./test_helpers/scenarios/ping-pong/without_crlf",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/without_crlf",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
	}

	tester_utils.TestTesterOutput(t, testerDefinition, testCases)
}

func normalizeTesterOutput(testerOutput []byte) []byte {
	re, _ := regexp.Compile("read tcp 127.0.0.1:\\d+->127.0.0.1:6379: read: connection reset by peer")
	return re.ReplaceAll(testerOutput, []byte("read tcp 127.0.0.1:xxxxx->127.0.0.1:6379: read: connection reset by peer"))
}
