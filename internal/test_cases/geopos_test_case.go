package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type GeoPosTestCase struct {
	Key                 string
	LocationNames       []string
	Tolerance           float64
	ExpectedCoordinates []*data_structures.Coordinates
}

// Run : I'll remove this comment later
// I wonder if this can be improved, stuck in a dilemma.
// From its use in `test_geospatial_geopos.go`, it is not apparent that nil is converted to NilAssertion()
// While if we use an ExpectedResponseArray like TransactionTestCase, that will either result in:
// 1. repitition of the logic: line 25-33 below.
// 2. if we delegate this logic to some utility function, it introduces one more step, resulting in the same issue as line 19 above
func (t *GeoPosTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	assertion := make([]resp_assertions.RESPAssertion, len(t.ExpectedCoordinates))
	for i, coordinate := range t.ExpectedCoordinates {
		if coordinate == nil {
			assertion[i] = resp_assertions.NewNilAssertion()
		} else {
			coordinatesAssertion := resp_assertions.NewOrderedArrayAssertion([]resp_assertions.RESPAssertion{
				resp_assertions.NewFloatingPointBulkStringAssertion(coordinate.Longitude, t.Tolerance),
				resp_assertions.NewFloatingPointBulkStringAssertion(coordinate.Latitude, t.Tolerance),
			})
			assertion[i] = coordinatesAssertion
		}
	}
	sendCommandTestCase := SendCommandTestCase{
		Command:   "GEOPOS",
		Args:      append([]string{t.Key}, t.LocationNames...),
		Assertion: resp_assertions.NewOrderedArrayAssertion(assertion),
	}

	return sendCommandTestCase.Run(client, logger)
}
