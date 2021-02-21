package internal

import (
	"testing"
)

func TestStagesMatchYAML(t *testing.T) {
	testerDefinition.TestAgainstYAML(t, "test_helpers/course_definition.yml")
}
