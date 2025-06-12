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
			UntilStageSlug:      "jm1",
			CodePath:            "./test_helpers/scenarios/bind/failure",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/bind/failure",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"bind_timeout": {
			UntilStageSlug:      "jm1",
			CodePath:            "./test_helpers/scenarios/bind/timeout",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/bind/timeout",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"bind_success": {
			UntilStageSlug:      "jm1",
			CodePath:            "./test_helpers/scenarios/bind/success",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/bind/success",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"ping_pong_eof": {
			UntilStageSlug:      "rg2",
			CodePath:            "./test_helpers/scenarios/ping-pong/eof",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/eof",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"ping_pong_without_crlf": {
			UntilStageSlug:      "rg2",
			CodePath:            "./test_helpers/scenarios/ping-pong/without_crlf",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/without_crlf",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"ping_pong_slow_response": {
			UntilStageSlug:      "rg2",
			CodePath:            "./test_helpers/scenarios/ping-pong/slow_response",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/slow_response",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"ping_pong_without_read_multiple_pongs": {
			UntilStageSlug:      "rg2",
			CodePath:            "./test_helpers/scenarios/ping-pong/without_read_multiple_pongs",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/without_read_multiple_pongs",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"expiry_pass": {
			UntilStageSlug:      "yz1",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/expiry/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"rdb_pass": {
			UntilStageSlug:      "sm4",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/rdb-read-value-with-expiry/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"repl_pass": {
			UntilStageSlug:      "na2",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/repl-wait/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"streams_pass": {
			UntilStageSlug:      "xu1",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/streams/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		"transactions_pass": {
			UntilStageSlug:      "jf8",
			CodePath:            "./test_helpers/pass_all",
			ExpectedExitCode:    0,
			StdoutFixturePath:   "./test_helpers/fixtures/transactions/pass",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
	}

	tester_utils_testing.TestTesterOutput(t, testerDefinition, testCases)
}

func normalizeTesterOutput(testerOutput []byte) []byte {
	replacements := map[string][]*regexp.Regexp{
		"tcp_port":                {regexp.MustCompile(`read tcp 127.0.0.1:\d+->127.0.0.1:6379: read: connection reset by peer`)},
		" tmp_dir ":               {regexp.MustCompile(` /private/var/folders/[^ ]+ `), regexp.MustCompile(` /tmp/[^ ]+ `)},
		"$length\\r\\ntmp_dir\\r": {regexp.MustCompile(`\$\d+\\r\\n/private/var/folders/[^ ]+\\r\\n`), regexp.MustCompile(`\$\d+\\r\\n/tmp/[^ ]+\\r\\n`)},
		"\"tmp_dir\"":             {regexp.MustCompile(`"/private/var/folders/[^"]+"`), regexp.MustCompile(`"/tmp/[^"]+"`)},
		"timestamp":               {regexp.MustCompile(`\d{2}:\d{2}:\d{2}\.\d{3}`)},
		"info_replication":        {regexp.MustCompile(`"# Replication\\r\\n[^"]+"`)},
		"replication_id":          {regexp.MustCompile(`FULLRESYNC [A-Za-z0-9]+ 0`)},
		"wait_timeout":            {regexp.MustCompile(`WAIT command returned after [0-9]+ ms`)},
		"xadd_id":                 {regexp.MustCompile(`\d{13}-\d+`)},
		"rdb_bytes":               {regexp.MustCompile(`"\$[0-9]+\\r\\nREDIS.*"`)},
		"info_replication_bytes":  {regexp.MustCompile(`"\$[0-9]+\\r\\n# Replication\\r\\n[^"]+"`)},
		"rdb_keys":                {regexp.MustCompile(`\[tester::#JW4\] .*Received .*`)},
		"hexdump":                 {regexp.MustCompile(`[0-9a-fA-F]{4} \| [0-9a-fA-F ]{47} \| .{0,16}`)},
	}

	for replacement, regexes := range replacements {
		for _, regex := range regexes {
			testerOutput = regex.ReplaceAll(testerOutput, []byte(replacement))
		}
	}

	return testerOutput
}
