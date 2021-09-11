
package internal

import (
	"bytes"
	"fmt"
	tester_utils "github.com/codecrafters-io/tester-utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"
)

type TesterOutputTestConfiguration struct {
	CodePath string
	ExpectedExitCode int
	StageName string
	StdoutFixturePath string
}

func TestStages(t *testing.T) {
	tests := map[string]TesterOutputTestConfiguration{
		"bind_failure": {
			StageName: "init",
			CodePath: "./test_helpers/scenarios/bind/failure",
			ExpectedExitCode: 1,
			StdoutFixturePath: "./test_helpers/fixtures/bind/failure",
		},
		"bind_success": {
			StageName: "init",
			CodePath: "./test_helpers/scenarios/bind/success",
			ExpectedExitCode: 0,
			StdoutFixturePath: "./test_helpers/fixtures/bind/success",
		},
		"ping_pong_failure": {
			StageName: "ping-pong",
			CodePath: "./test_helpers/scenarios/ping-pong/eof",
			ExpectedExitCode: 1,
			StdoutFixturePath: "./test_helpers/fixtures/ping-pong/failure",
		},
	}

	m := NewStdIOMocker()
	defer m.End()

	for testName, config := range tests {
		t.Run(testName, func(t *testing.T) {
			m.Start()

			exitCode := runCLIStage(config.StageName, config.CodePath)
			if !assert.Equal(t, config.ExpectedExitCode, exitCode) {
				failWithMockerOutput(t, m)
			}

			m.End()
			CompareOutputWithFixture(t, m.ReadStdout(), config.StdoutFixturePath)
		})
	}
}

func CompareOutputWithFixture(t *testing.T, testerOutput []byte, fixturePath string) {
	shouldRecordFixture := os.Getenv("CODECRAFTERS_RECORD_FIXTURES")

	if shouldRecordFixture == "true" {
		if err := os.MkdirAll(filepath.Dir(fixturePath), os.ModePerm); err != nil {
			panic(err)
		}

		if err := os.WriteFile(fixturePath, testerOutput, 0644); err != nil {
			panic(err)
		}

		return
	}

	fixtureContents, err := os.ReadFile(fixturePath)
	if err != nil {
		panic(err)
	}

	testerOutput = normalizeTesterOutput(testerOutput)
	fixtureContents = normalizeTesterOutput(fixtureContents)

	if bytes.Compare(testerOutput, fixtureContents) != 0 {
		diffExecutablePath, err := exec.LookPath("diff")
		if err != nil {
			panic(err)
		}

		diffExecutable := tester_utils.NewExecutable(diffExecutablePath)

		tmpFile, err := ioutil.TempFile("", "")
		if err != nil {
			panic(err)
		}

		if _, err = tmpFile.Write(testerOutput); err != nil {
			panic(err)
		}

		result, err := diffExecutable.Run(fixturePath, tmpFile.Name())
		if err != nil {
			panic(err)
		}

		os.Stdout.Write(result.Stdout)
		t.FailNow()
	}
}

func normalizeTesterOutput(testerOutput []byte) []byte {
	re, _ := regexp.Compile("read tcp 127.0.0.1:\\d+->127.0.0.1:6379: read: connection reset by peer")
	return re.ReplaceAll(testerOutput, []byte("read tcp 127.0.0.1:xxxxx+->127.0.0.1:6379: read: connection reset by peer"))
}

func runCLIStage(slug string, path string) (exitCode int) {
	return RunCLI(map[string]string{
		"CODECRAFTERS_CURRENT_STAGE_SLUG": slug,
		"CODECRAFTERS_COURSE_PAGE_URL": "http://dummy_url",
		"CODECRAFTERS_SUBMISSION_DIR":     path,
	})
}

func failWithMockerOutput(t *testing.T, m *IOMocker) {
	m.End()
	t.Error(fmt.Sprintf("stdout: \n%s\n\nstderr: \n%s", m.ReadStdout(), m.ReadStderr()))
	t.FailNow()
}
