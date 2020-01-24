package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiresAppDir(t *testing.T) {
	_, err := GetContext(map[string]string{})
	if !assert.Error(t, err) {
		t.FailNow()
	}
}

func TestSuccessParse(t *testing.T) {
	context, err := GetContext(map[string]string{
		"APP_DIR": "./test_helpers/valid_app_dir",
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, context.binaryPath, "test_helpers/valid_app_dir/spawn_redis_server.sh")
	assert.Equal(t, context.currentStageIndex, 2) // 3 - 1, for number -> index
}
