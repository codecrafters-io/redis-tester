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
	// If ShouldSkipCoordinatesVerfication is true, only floating point parsing is checked for existing locations
	ShouldSkipCoordinatesVerfication bool
}

type locationNameWithAssertion struct {
	LocationName string
	Assertion    resp_assertions.RESPAssertion
}

type locationAssertionCollection struct {
	locationAssertions []locationNameWithAssertion
}

func (c *locationAssertionCollection) append(locationAssertion locationNameWithAssertion) {
	c.locationAssertions = append(c.locationAssertions, locationAssertion)
}

func (c *locationAssertionCollection) shuffle() {
	c.locationAssertions = random.ShuffleArray(c.locationAssertions)
}

func (c *locationAssertionCollection) locationNames() []string {
	locationNames := []string{}

	for _, locationAssertion := range c.locationAssertions {
		locationNames = append(locationNames, locationAssertion.LocationName)
	}

	return locationNames
}

func (c *locationAssertionCollection) assertions() []resp_assertions.RESPAssertion {
	assertions := []resp_assertions.RESPAssertion{}

	for _, locationAssertion := range c.locationAssertions {
		assertions = append(assertions, locationAssertion.Assertion)
	}

	return assertions
}

func (t *GeoPosTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	locationAssertions := locationAssertionCollection{}

	// Assertions for existing locations
	for _, location := range t.Locations {
		tolerance := 10e-6

		if t.ShouldSkipCoordinatesVerfication {
			tolerance = math.Inf(1)
		}

		locationAssertions.append(locationNameWithAssertion{
			LocationName: location.Name,
			Assertion: resp_assertions.NewOrderedArrayAssertion([]resp_assertions.RESPAssertion{
				resp_assertions.NewFloatingPointBulkStringAssertion(location.Coordinates.Longitude, tolerance),
				resp_assertions.NewFloatingPointBulkStringAssertion(location.Coordinates.Latitude, tolerance),
			}),
		})
	}

	// Assertions for missing locations
	for _, missingLocationName := range t.MissingLocationNames {
		locationAssertions.append(locationNameWithAssertion{
			LocationName: missingLocationName,
			Assertion:    resp_assertions.NewNilAssertion(),
		})
	}

	locationAssertions.shuffle()

	sendCommandTestCase := SendCommandTestCase{
		Command:   "GEOPOS",
		Args:      append([]string{t.Key}, locationAssertions.locationNames()...),
		Assertion: resp_assertions.NewOrderedArrayAssertion(locationAssertions.assertions()),
	}

	return sendCommandTestCase.Run(client, logger)
}
