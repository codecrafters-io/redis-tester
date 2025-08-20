package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures/location"
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
	validLocation := location.GenerateRandomLocationSet(1).GetLocations()[0]
	validLatitude := validLocation.GetLatitude()
	validLongitude := validLocation.GetLongitude()
	errorPatternWrongLatitude := `^ERR.*(?i:latitude)`
	errorPatternWrongLongitude := `^ERR.*(?i:longitude)`

	// Invalid latitude, valid longitude
	// Latitude greater than max boundary
	locationWithPositiveInvalidLatitude := location.Location{
		Coordinates: location.NewInvalidCoordinates(
			testerutils_random.RandomFloat64(location.LATITUDE_MAX+1, 500),
			validLongitude,
		),
		Name: testerutils_random.RandomWord(),
	}
	positiveInvalidLatitudeTestCase := test_cases.InvalidGeoAddTestCase{
		Key:                  locationKey,
		Location:             locationWithPositiveInvalidLatitude,
		ExpectedErrorPattern: errorPatternWrongLatitude,
	}
	if err := positiveInvalidLatitudeTestCase.Run(client, logger); err != nil {
		return err
	}

	// Latitude smaller than min boundary
	locationWithNegativeInvalidLatitude := location.Location{
		Coordinates: location.NewInvalidCoordinates(
			testerutils_random.RandomFloat64(-500, location.LATITUDE_MIN-1),
			validLongitude,
		),
		Name: testerutils_random.RandomWord(),
	}
	negativeInvalidLatitudeTestCase := test_cases.InvalidGeoAddTestCase{
		Key:                  locationKey,
		Location:             locationWithNegativeInvalidLatitude,
		ExpectedErrorPattern: errorPatternWrongLatitude,
	}
	if err := negativeInvalidLatitudeTestCase.Run(client, logger); err != nil {
		return err
	}

	// Invalid longitude, but valid latitude
	// Longitude greater than max boundary
	locationWithPositiveInvalidLongitude := location.Location{
		Coordinates: location.NewInvalidCoordinates(
			validLatitude,
			testerutils_random.RandomFloat64(location.LONGITUDE_MAX+1, 500),
		),
		Name: testerutils_random.RandomWord(),
	}
	positiveInvalidLongitudeTestCase := test_cases.InvalidGeoAddTestCase{
		Key:                  locationKey,
		Location:             locationWithPositiveInvalidLongitude,
		ExpectedErrorPattern: errorPatternWrongLongitude,
	}
	if err := positiveInvalidLongitudeTestCase.Run(client, logger); err != nil {
		return err
	}

	// Longitude smaller than min boundary
	locationWithNegativeInvalidLongitude := location.Location{
		Coordinates: location.NewInvalidCoordinates(
			validLatitude,
			testerutils_random.RandomFloat64(-500, location.LONGITUDE_MIN-1),
		),
		Name: testerutils_random.RandomWord(),
	}
	negativeInvalidLongitudeTestCase := test_cases.InvalidGeoAddTestCase{
		Key:                  locationKey,
		Location:             locationWithNegativeInvalidLongitude,
		ExpectedErrorPattern: errorPatternWrongLongitude,
	}
	return negativeInvalidLongitudeTestCase.Run(client, logger)
}
