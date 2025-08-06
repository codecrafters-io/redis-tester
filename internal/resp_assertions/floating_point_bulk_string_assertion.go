package resp_assertions

import (
	"fmt"
	"strconv"

	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
)

type FloatingPointBulkStringAssertion struct {
	ExpectedValue float64
	Tolerance     float64
}

func NewFloatingPointBulkStringAssertion(expectedValue float64, tolerance float64) RESPAssertion {
	return FloatingPointBulkStringAssertion{
		ExpectedValue: expectedValue,
		Tolerance:     tolerance,
	}
}

func (a FloatingPointBulkStringAssertion) Run(value resp_value.Value) error {
	if value.Type != resp_value.BULK_STRING {
		return fmt.Errorf("Expected bulk string, got %s", value.Type)
	}

	stringValue := value.String()
	f, err := strconv.ParseFloat(stringValue, 64)
	if err != nil {
		return fmt.Errorf("Expected %q to be a floating point number", stringValue)
	}

	diff := f - a.ExpectedValue
	if diff < 0 {
		diff = -diff
	}

	if diff > a.Tolerance {
		expectedStr := fmt.Sprintf("%g ± %g", a.ExpectedValue, a.Tolerance)
		return fmt.Errorf("Expected %s, got %g", expectedStr, f)
	}

	return nil
}
