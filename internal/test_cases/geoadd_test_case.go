package test_cases

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type GeoAddTestCase struct {
	Key                         string
	Location                    location.Location
	ExpectedAddedLocationsCount int
}

func (t *GeoAddTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	longitudeStr := strconv.FormatFloat(t.Location.GetLongitude(), 'f', -1, 64)
	latitudeStr := strconv.FormatFloat(t.Location.GetLatitude(), 'f', -1, 64)

	sendCommandTestCase := SendCommandTestCase{
		Command:   "GEOADD",
		Args:      []string{t.Key, longitudeStr, latitudeStr, t.Location.Name},
		Assertion: resp_assertions.NewIntegerAssertion(t.ExpectedAddedLocationsCount),
	}

	return sendCommandTestCase.Run(client, logger)
}
