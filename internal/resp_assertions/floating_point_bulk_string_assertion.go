package resp_assertions

import (
	"fmt"
	"math"
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
	floatValue, err := strconv.ParseFloat(stringValue, 64)

	if err != nil {
		return fmt.Errorf("Expected %q to be a floating point number", stringValue)
	}

	diff := math.Abs(floatValue - a.ExpectedValue)

	if diff > a.Tolerance {
		// display full precision
		expectedValueStr := strconv.FormatFloat(a.ExpectedValue, 'f', -1, 64)
		toleranceStr := ""

		if a.Tolerance != 0 {
			toleranceStr = fmt.Sprintf(" (± %g)", a.Tolerance)
		}

		expectedStr := fmt.Sprintf("%s%s", expectedValueStr, toleranceStr)
		return fmt.Errorf("Expected %s, got %s", expectedStr, stringValue)
	}

	return nil
}
