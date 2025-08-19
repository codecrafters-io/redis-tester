package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type GeoAddTestCase struct {
	key      string
	location location.Location
	isValid  bool

	// only one of these is used depending on the value of isValid
	expectedErrorPattern        string
	expectedAddedLocationsCount int
}

func NewGeoAddTestCaseWithAddedLocation(key string, location location.Location, expectedAddedLocationsCount int) *GeoAddTestCase {
	return &GeoAddTestCase{
		key:                         key,
		location:                    location,
		isValid:                     true,
		expectedAddedLocationsCount: expectedAddedLocationsCount,
	}
}

func NewGeoAddTestCaseWithError(key string, location location.Location, errorPattern string) *GeoAddTestCase {
	return &GeoAddTestCase{
		key:                  key,
		location:             location,
		isValid:              false,
		expectedErrorPattern: errorPattern,
	}
}

func (t *GeoAddTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	var assertion resp_assertions.RESPAssertion
	if t.isValid {
		assertion = resp_assertions.NewIntegerAssertion(t.expectedAddedLocationsCount)
	} else {
		assertion = resp_assertions.NewRegexErrorAssertion(t.expectedErrorPattern)
	}
	sendCommandTestCase := SendCommandTestCase{
		Command:   "GEOADD",
		Args:      append([]string{t.key}, t.location.AsRedisCommandArgs()...),
		Assertion: assertion,
	}
	return sendCommandTestCase.Run(client, logger)
}
