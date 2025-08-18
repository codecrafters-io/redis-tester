package test_cases

import (
	"math"

	"github.com/codecrafters-io/redis-tester/internal/data_structures"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

const _GEOPOS_TOLERANCE = 10e-6

type GeoPosTestCase struct {
	key              string
	locations        []data_structures.Location
	missingLocations []string
	// If verifyCoordinates is true, only floating point parsing is checked for existing locations
	verifyCoordinates bool
}

func NewGeoPosTestCase(key string, onlyParseCoordinates bool) *GeoPosTestCase {
	return &GeoPosTestCase{
		key:               key,
		verifyCoordinates: onlyParseCoordinates,
	}
}

func (t *GeoPosTestCase) AddLocations(locations []data_structures.Location) {
	t.locations = append(t.locations, locations...)
}

func (t *GeoPosTestCase) AddMissingLocations(locationNames []string) {
	t.missingLocations = append(t.missingLocations, locationNames...)
}

func (t *GeoPosTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	assertions := make([]resp_assertions.RESPAssertion, len(t.locations)+len(t.missingLocations))

	for i, coordinate := range t.locations {
		tolerance := math.Inf(1)
		if t.verifyCoordinates {
			tolerance = _GEOPOS_TOLERANCE
		}
		coordinatesAssertion := resp_assertions.NewOrderedArrayAssertion([]resp_assertions.RESPAssertion{
			resp_assertions.NewFloatingPointBulkStringAssertion(coordinate.GetLongitude(), tolerance),
			resp_assertions.NewFloatingPointBulkStringAssertion(coordinate.GetLatitude(), tolerance),
		})
		assertions[i] = coordinatesAssertion
	}

	offset := len(t.locations)
	for i := range t.missingLocations {
		assertions[offset+i] = resp_assertions.NewNilAssertion()
	}

	allLocationNames := make([]string, len(t.locations)+len(t.missingLocations))
	for i, location := range t.locations {
		allLocationNames[i] = location.Name
	}
	for i, missingLocation := range t.missingLocations {
		allLocationNames[offset+i] = missingLocation
	}

	sendCommandTestCase := SendCommandTestCase{
		Command:   "GEOPOS",
		Args:      append([]string{t.key}, allLocationNames...),
		Assertion: resp_assertions.NewOrderedArrayAssertion(assertions),
	}

	return sendCommandTestCase.Run(client, logger)
}
