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
		// "bind_failure": {
		// 	UntilStageSlug:      "jm1",
		// 	CodePath:            "./test_helpers/scenarios/bind/failure",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/bind/failure",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "bind_timeout": {
		// 	UntilStageSlug:      "jm1",
		// 	CodePath:            "./test_helpers/scenarios/bind/timeout",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/bind/timeout",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "bind_success": {
		// 	UntilStageSlug:      "jm1",
		// 	CodePath:            "./test_helpers/scenarios/bind/success",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/bind/success",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "ping_pong_eof": {
		// 	UntilStageSlug:      "rg2",
		// 	CodePath:            "./test_helpers/scenarios/ping-pong/eof",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/eof",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "ping_pong_without_crlf": {
		// 	UntilStageSlug:      "rg2",
		// 	CodePath:            "./test_helpers/scenarios/ping-pong/without_crlf",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/without_crlf",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "ping_pong_slow_response": {
		// 	UntilStageSlug:      "rg2",
		// 	CodePath:            "./test_helpers/scenarios/ping-pong/slow_response",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/slow_response",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "ping_pong_without_read_multiple_pongs": {
		// 	UntilStageSlug:      "rg2",
		// 	CodePath:            "./test_helpers/scenarios/ping-pong/without_read_multiple_pongs",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/without_read_multiple_pongs",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "ping_pong_string_type_mismatch": {
		// 	StageSlugs:          []string{"rg2"},
		// 	CodePath:            "./test_helpers/scenarios/ping-pong/string_type_mismatch",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/ping-pong/string_type_mismatch",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "invalid_resp_error": {
		// 	StageSlugs:          []string{"rg2"},
		// 	CodePath:            "./test_helpers/scenarios/invalid-resp/",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/invalid-resp/invalid_resp_error",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "expiry_pass": {
		// 	UntilStageSlug:      "yz1",
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/expiry/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "rdb_pass": {
		// 	StageSlugs:          []string{"zg5", "jz6", "gc6", "jw4", "dq3", "sm4"},
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/rdb-read-value-with-expiry/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "repl_propagation_retry": {
		// 	StageSlugs:          []string{"yg4"},
		// 	CodePath:            "./test_helpers/scenarios/repl_propagation_retry",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/repl-wait/repl_propagation_retry",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		"repl_fullresync_wrong_pattern": {
			StageSlugs:          []string{"vm3"},
			CodePath:            "./test_helpers/scenarios/repl_fullresync_wrong_pattern",
			ExpectedExitCode:    1,
			StdoutFixturePath:   "./test_helpers/fixtures/repl-wait/repl_fullresync_wrong_pattern",
			NormalizeOutputFunc: normalizeTesterOutput,
		},
		// "repl_pass": {
		// 	StageSlugs:          []string{"bw1", "ye5", "hc6", "xc1", "gl7", "eh4", "ju6", "fj0", "vm3", "cf8", "zn8", "hd5", "yg4", "xv6", "yd3", "my8", "tu8", "na2"},
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/repl-wait/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "streams_pass": {
		// 	StageSlugs:          []string{"cc3", "cf6", "hq8", "yh3", "xu6", "zx1", "yp1", "fs1", "um0", "ru9", "bs1", "hw1", "xu1"},
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/streams/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "transactions_pass": {
		// 	StageSlugs:          []string{"si4", "lz8", "mk1", "pn0", "lo4", "we1", "rs9", "fy6", "rl9", "sg9", "jf8"},
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/transactions/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "lists_blpop_all_clients": {
		// 	StageSlugs:          []string{"ec3"},
		// 	CodePath:            "./test_helpers/scenarios/blpop-all-clients",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/lists/blpop_all_clients",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "lists_pass": {
		// 	StageSlugs:          []string{"mh6", "tn7", "lx4", "sf6", "ri1", "gu5", "fv6", "ef1", "jp1", "ec3", "xj7"},
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/lists/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "pubsub_pass": {
		// 	StageSlugs:          []string{"mx3", "zc8", "aw8", "lf1", "hf2", "dn4", "ze9"},
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/pubsub/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "zset_pass": {
		// 	StageSlugs:          []string{"ct1", "hf1", "lg6", "ic1", "bj4", "kn4", "gd7", "sq7"},
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/zset/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "geospatial_pass": {
		// 	StageSlugs:          []string{"zt4", "ck3", "tn5", "cr3", "xg4", "hb5", "ek6", "rm9"},
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/geospatial/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "auth_pass": {
		// 	StageSlugs:          []string{"jn4", "gx8", "ql6", "pl7", "uv9", "hz3", "nm2", "ws7"},
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/auth/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "auth_always_nopass_flag": {
		// 	StageSlugs:          []string{"uv9"},
		// 	CodePath:            "./test_helpers/scenarios/auth_always_nopass_flag",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/auth/auth_always_nopass_flag",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "auth_mistake_literal_password": {
		// 	StageSlugs:          []string{"pl7"},
		// 	CodePath:            "./test_helpers/scenarios/auth_mistake_literal_password",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/auth/auth_mistake_literal_password",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "auth_mistake_sha256": {
		// 	StageSlugs:          []string{"uv9"},
		// 	CodePath:            "./test_helpers/scenarios/auth_mistake_sha256",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/auth/auth_mistake_sha256",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "auth_default_user_authentication_wrong_error_pattern": {
		// 	StageSlugs:          []string{"nm2"},
		// 	CodePath:            "./test_helpers/scenarios/auth_acl_whoami_wrong_error_pattern",
		// 	ExpectedExitCode:    1,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/auth/auth_acl_whoami_wrong_error_pattern",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
		// "optimistic_locking_pass": {
		// 	StageSlugs:          []string{"jb7", "jq9", "mh8", "fp0", "uo9", "bn1", "fn4", "hq1"},
		// 	CodePath:            "./test_helpers/pass_all",
		// 	ExpectedExitCode:    0,
		// 	StdoutFixturePath:   "./test_helpers/fixtures/optimistic_locking/pass",
		// 	NormalizeOutputFunc: normalizeTesterOutput,
		// },
	}

	tester_utils_testing.TestTesterOutput(t, testerDefinition, testCases)
}

func normalizeTesterOutput(testerOutput []byte) []byte {
	replacements := map[string][]*regexp.Regexp{
		"tcp_port":                {regexp.MustCompile(`read tcp 127.0.0.1:\d+->127.0.0.1:6379: read: connection reset by peer`)},
		" tmp_dir ":               {regexp.MustCompile(` /private/tmp/[^ ]+ `), regexp.MustCompile(` /tmp/[^ ]+ `)},
		"$length\\r\\ntmp_dir\\r": {regexp.MustCompile(`\$\d+\\r\\n/private/tmp/[^ ]+\\r\\n`), regexp.MustCompile(`\$\d+\\r\\n/tmp/[^ ]+\\r\\n`)},
		"\"tmp_dir\"":             {regexp.MustCompile(`"/private/tmp/[^"]+"`), regexp.MustCompile(`"/tmp/[^"]+"`)},
		"timestamp":               {regexp.MustCompile(`\d{2}:\d{2}:\d{2}\.\d{3}`)},
		"info_replication":        {regexp.MustCompile(`"# Replication\\r\\n[^"]+"`)},
		"replication_id":          {regexp.MustCompile(`FULLRESYNC [A-Za-z0-9]+ 0`)},
		"wait_timeout":            {regexp.MustCompile(`WAIT command returned after [0-9]+ ms`)},
		"xadd_id":                 {regexp.MustCompile(`\d{13}-\d+`)},
		"rdb_bytes":               {regexp.MustCompile(`"\$[0-9]+\\r\\nREDIS.*"`)},
		"info_replication_bytes":  {regexp.MustCompile(`"\$[0-9]+\\r\\n# Replication\\r\\n[^"]+"`)},
		"rdb_keys":                {regexp.MustCompile(`\[tester::#JW4\] .*Received .*`), regexp.MustCompile(`\[tester::#JW4\].*"(apple|orange|banana|pear|grape|pineapple|mango|strawberry|raspberry|blueberry)",?.*`)},
		"hexdump":                 {regexp.MustCompile(`[0-9a-fA-F]{4} \| [0-9a-fA-F ]{47} \| .{0,16}`)},
		"client_connected":        {regexp.MustCompile(`Connected \(port \d+ -> port \d+\)`)},
	}

	for replacement, regexes := range replacements {
		for _, regex := range regexes {
			testerOutput = regex.ReplaceAll(testerOutput, []byte(replacement))
		}
	}

	return testerOutput
}
