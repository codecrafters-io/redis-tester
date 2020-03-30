package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStart(t *testing.T) {
	err := NewExecutable("/blah").Start()
	assertErrorContains(t, err, "no such file")
	assertErrorContains(t, err, "/blah")

	err = NewExecutable("./test_helpers/executable_test/stdout_echo.sh").Start()
	assert.NoError(t, err)
}

func assertErrorContains(t *testing.T, err error, expectedMsg string) {
	assert.Contains(t, err.Error(), expectedMsg)
}

func TestRun(t *testing.T) {
	e := NewExecutable("./test_helpers/executable_test/stdout_echo.sh")
	result, err := e.Run("hey")
	assert.NoError(t, err)
	assert.Equal(t, "hey\n", string(result.Stdout))
}

func TestOutputCapture(t *testing.T) {
	// Stdout capture
	e := NewExecutable("./test_helpers/executable_test/stdout_echo.sh")
	result, err := e.Run("hey")

	assert.NoError(t, err)
	assert.Equal(t, "hey\n", string(result.Stdout))
	assert.Equal(t, "", string(result.Stderr))

	// Stderr capture
	e = NewExecutable("./test_helpers/executable_test/stderr_echo.sh")
	result, err = e.Run("hey")

	assert.NoError(t, err)
	assert.Equal(t, "", string(result.Stdout))
	assert.Equal(t, "hey\n", string(result.Stderr))
}

func TestExitCode(t *testing.T) {
	e := NewExecutable("./test_helpers/executable_test/exit_with.sh")

	result, _ := e.Run("0")
	assert.Equal(t, 0, result.ExitCode)

	result, _ = e.Run("1")
	assert.Equal(t, 1, result.ExitCode)

	result, _ = e.Run("2")
	assert.Equal(t, 2, result.ExitCode)
}

func TestExecutableStartNotAllowedIfInProgress(t *testing.T) {
	e := NewExecutable("./test_helpers/executable_test/sleep_for.sh")

	// Run once
	err := e.Start("0.01")
	assert.NoError(t, err)

	// Starting again when in progress should throw an error
	err = e.Start("0.01")
	assertErrorContains(t, err, "process already in progress")

	// Running again when in progress should throw an error
	_, err = e.Run("0.01")
	assertErrorContains(t, err, "process already in progress")

	e.Wait()

	// Running again once finished should be fine
	err = e.Start("0.01")
	assert.NoError(t, err)
}

func TestSuccessiveExecutions(t *testing.T) {
	e := NewExecutable("./test_helpers/executable_test/stdout_echo.sh")

	result, _ := e.Run("1")
	assert.Equal(t, "1\n", string(result.Stdout))

	result, _ = e.Run("2")
	assert.Equal(t, "2\n", string(result.Stdout))
}
