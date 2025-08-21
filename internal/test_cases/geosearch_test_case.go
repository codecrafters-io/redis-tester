package test_cases

import (
	"strconv"

	location_ds "github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type GeoSearchTestCase struct {
	Key                   string
	FromCoordinates       location_ds.Coordinates
	Radius                float64
	ExpectedLocationNames []string
}

func (t *GeoSearchTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	longitudeStr := t.FromCoordinates.LongitudeAsRedisCommandArg()
	latitudeStr := t.FromCoordinates.LatitudeAsRedisCommandArg()
	radiusStr := strconv.FormatFloat(t.Radius, 'f', -1, 64)

	// we use unordered assertion because we do not cover ASC/DESC option for GEOSEARCH
	sendCommandTestCase := SendCommandTestCase{
		Command:   "GEOSEARCH",
		Args:      []string{t.Key, "FROMLONLAT", longitudeStr, latitudeStr, "BYRADIUS", radiusStr, "m"},
		Assertion: resp_assertions.NewUnorderedStringArrayAssertion(t.ExpectedLocationNames),
	}

	return sendCommandTestCase.Run(client, logger)
}
