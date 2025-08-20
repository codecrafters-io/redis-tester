package test_cases

import (
	"math"

	"github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
	"github.com/codecrafters-io/tester-utils/random"
)

type GeoPosTestCase struct {
	Key                  string
	Locations            []location.Location
	MissingLocationNames []string
	// If ShouldVerifyCoordinates is true, only floating point parsing is checked for existing locations
	ShouldVerifyCoordinates bool
}

func (t *GeoPosTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {

	allLocationsLen := len(t.Locations) + len(t.MissingLocationNames)
	allLocationNames := make([]string, allLocationsLen)
	allAssertions := make([]resp_assertions.RESPAssertion, allLocationsLen)

	tolerance := math.Inf(1)

	if t.ShouldVerifyCoordinates {
		tolerance = 10e-6
	}

	// Populate location names and assertion for existing locations
	for i, loc := range t.Locations {
		allLocationNames[i] = loc.Name

		allAssertions[i] = resp_assertions.NewOrderedArrayAssertion([]resp_assertions.RESPAssertion{
			resp_assertions.NewFloatingPointBulkStringAssertion(loc.Coordinates.Longitude, tolerance),
			resp_assertions.NewFloatingPointBulkStringAssertion(loc.Coordinates.Latitude, tolerance),
		})
	}

	// Populate location names and assertion for missing locations
	offset := len(t.Locations)

	for i, missingLocationName := range t.MissingLocationNames {
		allLocationNames[offset+i] = missingLocationName
		// WILL_CHANGE: NewNilArrayAssertion() after PR#211 is merged
		allAssertions[offset+i] = resp_assertions.NewNilAssertion()
	}

	// Shuffle location names and assertions in same order
	shuffledLocationNames := make([]string, allLocationsLen)
	shuffledAssertions := make([]resp_assertions.RESPAssertion, allLocationsLen)
	shuffledIndexes := random.RandomInts(0, allLocationsLen, allLocationsLen)

	for i, idx := range shuffledIndexes {
		shuffledLocationNames[i] = allLocationNames[idx]
		shuffledAssertions[i] = allAssertions[idx]
	}

	sendCommandTestCase := SendCommandTestCase{
		Command:   "GEOPOS",
		Args:      append([]string{t.Key}, shuffledLocationNames...),
		Assertion: resp_assertions.NewOrderedArrayAssertion(shuffledAssertions),
	}

	return sendCommandTestCase.Run(client, logger)
}
