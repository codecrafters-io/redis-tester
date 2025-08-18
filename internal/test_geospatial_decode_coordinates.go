package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures"
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

	locationSet := data_structures.GenerateRandomLocationSet(testerutils_random.RandomInt(2, 4))

	// We use ZADD to add the locations so the user implementation requires that they decode
	// co-ordinates from the score
	for _, location := range locationSet.GetLocations() {
		zaddTestCase := test_cases.ZaddTestCase{
			Key: locationKey,
			Member: data_structures.SortedSetMember{
				Name:  location.Name,
				Score: float64(location.GetGeoCode()),
			},
			ExpectedAddedMembersCount: 1,
		}
		if err := zaddTestCase.Run(client, logger); err != nil {
			return err
		}
	}

	geoPosTestCase := test_cases.NewGeoPosTestCase(locationKey)
	geoPosTestCase.AddLocations(locationSet.GetLocations(), false)

	return geoPosTestCase.Run(client, logger)
}
