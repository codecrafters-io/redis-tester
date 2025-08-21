package internal

import (
	location_ds "github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/data_structures/sorted_set"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	testerutils_random "github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testGeospatialDecodeCoordinates(stageHarness *test_case_harness.TestCaseHarness) error {
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

	locationSet := location_ds.GenerateRandomLocationSet(testerutils_random.RandomInt(2, 4))

	// We use ZADD to add the locations so the user implementation requires that they decode
	// co-ordinates from the score
	for _, location := range locationSet.GetLocations() {
		zaddTestCase := test_cases.ZaddTestCase{
			Key: locationKey,
			Member: sorted_set.SortedSetMember{
				Name:  location.Name,
				Score: float64(location.GetGeoCode()),
			},
			ExpectedAddedMembersCount: 1,
		}

		if err := zaddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	geoPosTestCase := test_cases.GeoPosTestCase{
		Key:       locationKey,
		Locations: locationSet.GetLocations(),
	}

	return geoPosTestCase.Run(client, logger)
}
