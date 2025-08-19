package internal

import (
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
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

	// WILL_REMOVE: re-use valid values of latitude and longitude to avoid generation and conversion of coordinates to float everytime
	// we generate a new location, because we're concerned only with invalid values in this stage
	locationName := validLocation.Name
	validLatitude := strconv.FormatFloat(validLocation.GetLatitude(), 'f', -1, 64)
	validLongitude := strconv.FormatFloat(validLocation.GetLongitude(), 'f', -1, 64)

	// Invalid latitude, valid longitude
	errorPatternWrongLatitude := `^ERR.*(?i:latitude)`

	// WILL_REMOVE: I could not use GeoAddTestCase since it is built using Location struct,
	// which is a valid location in mercator projection
	// If we remove the validations in the NewCoordinates(),
	// that way we can re-use GeoAddTestCase for this stage as well
	// however, we'll have to change the implementation of GeoAddTestCase to expect both error and integer (for invalid and valid coordinates respectively)
	// so we have to use two different constructor functions:
	// NewGeoAddTestCaseWithValidLocation() and NewGeoAddTestCaseWithInvalidLocation()
	// and move the coordinates checks in NewCoordinates() to these constructors, which does seem like unification of concern,
	// and (over-engineering?? not sure) given that NewGeoAddTestCaseWithInvalidLocation() is being used only in this stage and
	// NewGeoAddTestCaseWithValidLocation() is being used in all others except this one
	// But, let me know if there is a better way to do this

	// Latitude greater than max boundary
	invalidLatitude := strconv.FormatFloat(testerutils_random.RandomFloat64(location.LATITUDE_MAX, 500), 'f', -1, 64)
	positiveInvalidLatitudeTestCase := test_cases.SendCommandTestCase{
		Command:   "GEOADD",
		Args:      []string{locationKey, validLongitude, invalidLatitude, locationName},
		Assertion: resp_assertions.NewRegexErrorAssertion(errorPatternWrongLatitude),
	}
	if err := positiveInvalidLatitudeTestCase.Run(client, logger); err != nil {
		return err
	}

	invalidLatitude = strconv.FormatFloat(testerutils_random.RandomFloat64(-500, location.LATITUDE_MIN), 'f', -1, 64)
	negativeInvalidLatitudeTestCase := test_cases.SendCommandTestCase{
		Command:   "GEOADD",
		Args:      []string{locationKey, validLongitude, invalidLatitude, locationName},
		Assertion: resp_assertions.NewRegexErrorAssertion(errorPatternWrongLatitude),
	}
	if err := negativeInvalidLatitudeTestCase.Run(client, logger); err != nil {
		return err
	}

	// Invalid longitude, but valid latitude
	errorPatternWrongLongitude := `^ERR.*(?i:longitude)`

	invalidLongitude := strconv.FormatFloat(testerutils_random.RandomFloat64(location.LONGITUDE_MAX+1, 500), 'f', -1, 64)
	positiveInvalidLongitudeTestCase := test_cases.SendCommandTestCase{
		Command:   "GEOADD",
		Args:      []string{locationKey, invalidLongitude, validLatitude, locationName},
		Assertion: resp_assertions.NewRegexErrorAssertion(errorPatternWrongLongitude),
	}
	if err := positiveInvalidLongitudeTestCase.Run(client, logger); err != nil {
		return err
	}

	// Longitude smaller than min boundary
	invalidLongitude = strconv.FormatFloat(testerutils_random.RandomFloat64(-500, location.LONGITUDE_MIN), 'f', -1, 64)
	negativeInvalidLongitudeTestCase := test_cases.SendCommandTestCase{
		Command:   "GEOADD",
		Args:      []string{locationKey, invalidLongitude, validLatitude, locationName},
		Assertion: resp_assertions.NewRegexErrorAssertion(errorPatternWrongLongitude),
	}
	return negativeInvalidLongitudeTestCase.Run(client, logger)
}
