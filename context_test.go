package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiresBinaryPath(t *testing.T) {
	_, err := GetContext([]string{})
	if !assert.Error(t, err) {
		t.FailNow()
	}

	assert.Contains(t, err.Error(), "--binary-path")
	assert.Contains(t, err.Error(), "must be specified")
}

func TestRequiresConfigPath(t *testing.T) {
	_, err := GetContext([]string{"--binary-path", "dummy"})
	if !assert.Error(t, err) {
		t.FailNow()
	}

	assert.Contains(t, err.Error(), "--config-path")
	assert.Contains(t, err.Error(), "must be specified")
}

func TestSuccessParse(t *testing.T) {
	context, err := GetContext([]string{
		"--binary-path",
		"dummy",
		"--config-path",
		"./test_helpers/valid_config.yml",
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, context.binaryPath, "dummy")
	assert.Equal(t, context.currentStageIndex, 3)
}
