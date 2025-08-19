package test_cases

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type GeoSearchTestCase struct {
	Key                   string
	FromLongitude         float64
	FromLatitude          float64
	Radius                float64
	ExpectedLocationNames []string
}

func (t *GeoSearchTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	longitudeStr := strconv.FormatFloat(t.FromLongitude, 'f', -1, 64)
	latitudeStr := strconv.FormatFloat(t.FromLatitude, 'f', -1, 64)
	radiusStr := strconv.FormatFloat(t.Radius, 'f', -1, 64)

	// we use unordered assertion because we do not cover ASC/DESC option for GEOSEARCH
	sendCommandTestCase := SendCommandTestCase{
		Command:   "GEOSEARCH",
		Args:      []string{t.Key, "FROMLONLAT", longitudeStr, latitudeStr, "BYRADIUS", radiusStr, "m"},
		Assertion: resp_assertions.NewUnorderedStringArrayAssertion(t.ExpectedLocationNames),
	}

	return sendCommandTestCase.Run(client, logger)
}
