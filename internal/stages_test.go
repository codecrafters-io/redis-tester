
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
	"testing"
)

func TestBindFailure(t *testing.T) {
	m := NewStdIOMocker()
	m.Start()
	defer m.End()

	exitCode := runCLIStage("init", "./test_helpers/scenarios/bind/failure")
	if !assert.Equal(t, 1, exitCode) {
		failWithMockerOutput(t, m)
	}

	CompareOutputWithFixture(t, m.ReadStdout(), "./test_helpers/fixtures/bind/failure")
}

func TestBindSuccess(t *testing.T) {
	m := NewStdIOMocker()
	m.Start()
	defer m.End()

	exitCode := runCLIStage("init", "./test_helpers/scenarios/bind/success")
	if !assert.Equal(t, 0, exitCode) {
		failWithMockerOutput(t, m)
	}

	m.End()

	CompareOutputWithFixture(t, m.ReadStdout(), "./test_helpers/fixtures/bind/success")
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


//func TestBind(t *testing.T) {
//	m := NewStdIOMocker()
//	m.Start()
//	defer m.End()
//
//	fmt.Println("Test failure")
//	exitCode := runCLIStage("init", "./test_helpers/fixtures/failure")
//	if !assert.Equal(t, 1, exitCode) {
//		failWithMockerOutput(t, m)
//	}
//	assert.Contains(t, m.ReadStdout(), "Test failed")
//
//	m.Reset()
//
//	fmt.Println("Test success")
//	exitCode = runCLIStage("init", "./test_helpers/fixtures/success")
//	if !assert.Equal(t, 0, exitCode) {
//		failWithMockerOutput(t, m)
//	}
//}

//func TestRespondToPing(t *testing.T) {
//	m := NewStdIOMocker()
//	m.Start()
//	defer m.End()
//
//	exitCode := runCLIStage("ping-pong", "./test_helpers/fixtures/success")
//	if !assert.Equal(t, 1, exitCode) {
//		failWithMockerOutput(t, m)
//	}
//	assert.Contains(t, m.ReadStdout(), "Test failed")
//
//	m.Reset()
//
//	exitCode = runCLIStage("ping-pong", "./test_helpers/fixtures/success")
//	if !assert.Equal(t, 0, exitCode) {
//		failWithMockerOutput(t, m)
//	}
//}

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
