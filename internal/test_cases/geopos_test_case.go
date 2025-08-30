package test_cases

import (
	"fmt"
	"math"
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/testing"
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
	c.locationAssertions = testerutils_random.ShuffleArray(c.locationAssertions)
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
	if testing.IsRecordingOrEvaluatingFixtures() {
		client.SetReadValueInterceptor(reducePrecision)
		defer client.UnsetReadValueInterceptor()
	}
	locationAssertions := locationAssertionCollection{}

	// Assertions for existing locations
	for _, location := range t.Locations {
		tolerance := 1e-6

		if t.ShouldSkipCoordinatesVerfication {
			tolerance = math.Inf(1)
		}

		// Calculate geogrid center coordinates for assertion
		geoGridCenterCoordinates := location.GetGeoGridCenterCoordinates()
		locationAssertions.append(locationNameWithAssertion{
			LocationName: location.Name,
			Assertion: resp_assertions.NewOrderedArrayAssertion([]resp_assertions.RESPAssertion{
				resp_assertions.NewFloatingPointBulkStringAssertion(geoGridCenterCoordinates.Longitude, tolerance),
				resp_assertions.NewFloatingPointBulkStringAssertion(geoGridCenterCoordinates.Latitude, tolerance),
			}),
		})
	}

	// Assertions for missing locations
	for _, missingLocationName := range t.MissingLocationNames {
		locationAssertions.append(locationNameWithAssertion{
			LocationName: missingLocationName,
			Assertion:    resp_assertions.NewNilArrayAssertion(),
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

func reducePrecision(value resp_value.Value) resp_value.Value {
	switch value.Type {
	case resp_value.ARRAY:
		return reducePrecisionForArray(value)
	case resp_value.BULK_STRING:
		return reducePrecisionForBulkString(value)
	default:
		return value
	}
}

func reducePrecisionForArray(value resp_value.Value) resp_value.Value {
	if value.Type != resp_value.ARRAY {
		return value
	}

	var arrayElements []resp_value.Value

	for _, arrayElement := range value.Array() {
		arrayElements = append(arrayElements, reducePrecision(arrayElement))
	}

	return resp_value.NewArrayValue(arrayElements)
}

func reducePrecisionForBulkString(value resp_value.Value) resp_value.Value {
	if value.Type != resp_value.BULK_STRING {
		return value
	}

	floatValue, err := strconv.ParseFloat(value.String(), 64)

	if err != nil {
		return value
	}

	floatStringWithReducedPrecision := fmt.Sprintf("%.10f", floatValue)
	return resp_value.NewBulkStringValue(floatStringWithReducedPrecision)
}
