package test_cases

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/data_structures"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type GeoAddTestCase struct {
	key      string
	location data_structures.Location

	isExpectingError bool

	// Only one of them is used depending on the validity of location's coordinate
	expectedAddedLocationsCount int
	expectedPattern             string
}

func NewGeoAddTestCaseWithValidCoordinates(key string, location *data_structures.Location, expectedAddedMembersCount int) *GeoAddTestCase {
	return &GeoAddTestCase{
		key:                         key,
		location:                    *location,
		expectedAddedLocationsCount: expectedAddedMembersCount,
	}
}

func NewGeoAddTestCaseWithInvalidCoordinates(key string, location *data_structures.Location, expectedPattern string) *GeoAddTestCase {
	return &GeoAddTestCase{
		key:              key,
		location:         *location,
		isExpectingError: true,
		expectedPattern:  expectedPattern,
	}
}

func (t *GeoAddTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	longitudeStr := strconv.FormatFloat(t.location.GetLongitude(), 'f', -1, 64)
	latitudeStr := strconv.FormatFloat(t.location.GetLatitude(), 'f', -1, 64)

	var assertion resp_assertions.RESPAssertion
	if t.isExpectingError {
		assertion = resp_assertions.NewRegexErrorAssertion(t.expectedPattern)
	} else {
		assertion = resp_assertions.NewIntegerAssertion(t.expectedAddedLocationsCount)
	}
	sendCommandTestCase := SendCommandTestCase{
		Command:   "GEOADD",
		Args:      []string{t.key, longitudeStr, latitudeStr, t.location.Name},
		Assertion: assertion,
	}
	return sendCommandTestCase.Run(client, logger)
}
