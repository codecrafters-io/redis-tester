package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testGeospatialValidateCoordinates(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger

	client, err := instrumented_resp_connection.NewFromAddr(logger, "localhost:6379", "client")
	if err != nil {
		logFriendlyError(logger, err)
		return err
	}
	defer client.Close()

	locationKey := testerutils_random.RandomWord()
	// The co-ordinates of this location are used as a valid longitude/latitude in each test case
	location := data_structures.GenerateRandomLocations(1)[0]

	// Invalid latitude, valid longitude
	errorPatternWrongLatitude := `^ERR.*(?i:latitude)`

	// Latitude greater than max boundary
	locationName := testerutils_random.RandomWord()
	invalidLocation := data_structures.NewLocation(locationName, data_structures.Coordinates{
		Latitude:  testerutils_random.RandomFloat64(data_structures.LATITUDE_MAX+1, 500),
		Longitude: location.GetLongitude(),
	})
	geoAddTestCase := test_cases.NewGeoAddTestCaseWithInvalidCoordinates(locationKey, invalidLocation, errorPatternWrongLatitude)
	if err := geoAddTestCase.Run(client, logger); err != nil {
		return err
	}

	// Latitude smaller than min boundary
	locationName = testerutils_random.RandomWord()
	invalidLocation = data_structures.NewLocation(locationName, data_structures.Coordinates{
		Latitude:  testerutils_random.RandomFloat64(-500, data_structures.LATITUDE_MIN),
		Longitude: location.GetLongitude(),
	})
	geoAddTestCase = test_cases.NewGeoAddTestCaseWithInvalidCoordinates(locationKey, invalidLocation, errorPatternWrongLatitude)
	if err := geoAddTestCase.Run(client, logger); err != nil {
		return err
	}

	// Invalid longitude, but valid latitude
	errorPatternWrongLongitude := `^ERR.*(?i:longitude)`

	// Longitude greater than max boundary
	locationName = testerutils_random.RandomWord()
	invalidLocation = data_structures.NewLocation(locationName, data_structures.Coordinates{
		Latitude:  location.GetLatitude(),
		Longitude: testerutils_random.RandomFloat64(data_structures.LONGITUDE_MAX+1, 500),
	})
	geoAddTestCase = test_cases.NewGeoAddTestCaseWithInvalidCoordinates(locationKey, invalidLocation, errorPatternWrongLongitude)
	if err := geoAddTestCase.Run(client, logger); err != nil {
		return err
	}

	// Longitude smaller than min boundary
	locationName = testerutils_random.RandomWord()
	invalidLocation = data_structures.NewLocation(locationName, data_structures.Coordinates{
		Latitude:  location.GetLatitude(),
		Longitude: testerutils_random.RandomFloat64(-500, data_structures.LONGITUDE_MIN),
	})
	geoAddTestCase = test_cases.NewGeoAddTestCaseWithInvalidCoordinates(locationKey, invalidLocation, errorPatternWrongLongitude)
	return geoAddTestCase.Run(client, logger)
}
