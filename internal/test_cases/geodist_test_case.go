package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

const _GEODIST_TOLERANCE = 10e-4

type GeoDistTestCase struct {
	Key              string
	Location1        location.Location
	Location2        location.Location
	ExpectedDistance float64
}

func (t *GeoDistTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	distance := t.Location1.DistanceFrom(t.Location2)
	geodistTestCase := SendCommandTestCase{
		Command:   "GEODIST",
		Args:      []string{t.Key, t.Location1.Name, t.Location2.Name},
		Assertion: resp_assertions.NewFloatingPointBulkStringAssertion(distance, _GEODIST_TOLERANCE),
	}
	return geodistTestCase.Run(client, logger)
}
