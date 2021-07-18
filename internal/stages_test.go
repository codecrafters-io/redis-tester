
package internal

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBind(t *testing.T) {
	m := NewStdIOMocker()
	m.Start()
	defer m.End()

	fmt.Println("Test failure")
	exitCode := runCLIStage("init", "./test_helpers/stages/bind_failure")
	if !assert.Equal(t, 1, exitCode) {
		failWithMockerOutput(t, m)
	}
	assert.Contains(t, m.ReadStdout(), "Test failed")

	m.Reset()

	fmt.Println("Test success")
	exitCode = runCLIStage("init", "./test_helpers/stages/bind")
	if !assert.Equal(t, 0, exitCode) {
		failWithMockerOutput(t, m)
	}
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