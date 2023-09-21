package internal

import (
	"testing"

	tester_utils_testing "github.com/codecrafters-io/tester-utils/testing"
)

func TestStagesMatchYAML(t *testing.T) {
	tester_utils_testing.ValidateTesterDefinitionAgainstYAML(t, testerDefinition, "test_helpers/course_definition.yml")
}
