package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiresAppDir(t *testing.T) {
	_, err := GetContext(map[string]string{
		"CODECRAFTERS_CURRENT_STAGE_SLUG": "init",
	})
	if !assert.Error(t, err) {
		t.FailNow()
	}
}

func TestRequiresCurrentStageSlug(t *testing.T) {
	_, err := GetContext(map[string]string{
		"CODECRAFTERS_SUBMISSION_DIR": "./test_helpers/valid_app_dir",
	})
	if !assert.Error(t, err) {
		t.FailNow()
	}
}

func TestSuccessParse(t *testing.T) {
	context, err := GetContext(map[string]string{
		"CODECRAFTERS_CURRENT_STAGE_SLUG": "init",
		"CODECRAFTERS_SUBMISSION_DIR":     "./test_helpers/valid_app_dir",
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, context.binaryPath, "test_helpers/valid_app_dir/spawn_redis_server.sh")
	assert.Equal(t, context.currentStageSlug, "init")
}
