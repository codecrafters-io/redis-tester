package internal

import (
	"os"
	"regexp"
	"testing"

	tester_utils_testing "github.com/codecrafters-io/tester-utils/testing"
)

func TestStages(t *testing.T) {
	os.Setenv("CODECRAFTERS_RANDOM_SEED", "1234567890")

	testCases := map[string]tester_utils_testing.TesterOutputTestCase{
		"bind_failure": {
			UntilStageSlug:      "init",
			CodePath:            "./test_helpers/scenarios/bind/failure",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/bind/failure",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"bind_timeout": {
			UntilStageSlug:      "init",
			CodePath:            "./test_helpers/scenarios/bind/timeout",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/bind/timeout",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"bind_success": {
			UntilStageSlug:      "init",
			CodePath:            "./test_helpers/scenarios/bind/success",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/bind/success",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"ping_pong_eof": {
			UntilStageSlug:      "ping-pong",
			CodePath:            "./test_helpers/scenarios/ping-pong/eof",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/eof",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"ping_pong_without_crlf": {
			UntilStageSlug:      "ping-pong",
			CodePath:            "./test_helpers/scenarios/ping-pong/without_crlf",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/without_crlf",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"ping_pong_without_read_multiple_pongs": {
			UntilStageSlug:      "ping-pong",
			CodePath:            "./test_helpers/scenarios/ping-pong/without_read_multiple_pongs",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/without_read_multiple_pongs",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"expiry_pass": {
			UntilStageSlug:      "expiry",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/expiry/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"rdb_config_pass": {
			UntilStageSlug:      "rdb-config",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/rdb-config/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"rdb_read_key_pass": {
			UntilStageSlug:      "rdb-read-key",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/rdb-read-key/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"rdb_read_string_value_pass": {
			UntilStageSlug:      "rdb-read-string-value",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/rdb-string-value/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"rdb_read_multiple_keys_pass": {
			UntilStageSlug:      "rdb-read-multiple-keys",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/rdb-read-multiple-keys/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"rdb_read_multiple_string_values_pass": {
			UntilStageSlug:      "rdb-read-multiple-string-values",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/rdb-read-multiple-string-values/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"rdb_read_value_with_expiry_pass": {
			UntilStageSlug:      "rdb-read-value-with-expiry",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/rdb-read-value-with-expiry/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
	}

	tester_utils_testing.TestTesterOutput(t, testerDefinition, testCases)
}

func normalizeTesterOutput(testerOutput []byte) []byte {
	replacements := map[string][]*regexp.Regexp{
		"tcp_port":  {regexp.MustCompile("read tcp 127.0.0.1:\\d+->127.0.0.1:6379: read: connection reset by peer")},
		" tmp_dir ": {regexp.MustCompile(" /private/var/folders/[^ ]+ "), regexp.MustCompile(" /tmp/[^ ]+ ")},
		"timestamp": {regexp.MustCompile("\\d{2}:\\d{2}:\\d{2}\\.\\d{3}")},
	}

	for replacement, regexes := range replacements {
		for _, regex := range regexes {
			testerOutput = regex.ReplaceAll(testerOutput, []byte(replacement))
		}
	}

	return testerOutput
}
