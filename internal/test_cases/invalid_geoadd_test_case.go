package test_cases

import (
	"github.com/codecrafters-io/redis-tester/internal/data_structures/location"
	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type InvalidGeoAddTestCase struct {
	Key                     string
	Location                location.Location
	ExpectedErrorBeginsWith string
	ExpectedErrorContains   string
}

func (t *InvalidGeoAddTestCase) Run(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	args := []string{t.Key, t.Location.LongitudeAsRedisCommandArg(), t.Location.LatitudeAsRedisCommandArg(), t.Location.Name}

	sendCommandTestCase := SendCommandTestCase{
		Command: "GEOADD",
		Args:    args,
		Assertion: resp_assertions.PrefixAndSubstringsAssertion{
			Logger:       logger,
			ExpectedType: resp_value.ERROR,
			PrefixPredicate: &resp_assertions.PrefixPredicate{
				Prefix:        t.ExpectedErrorBeginsWith,
				CaseSensitive: true,
			},
			HasSubstringPredicates: []resp_assertions.HasSubstringPredicate{{
				Substring: t.ExpectedErrorContains,
			}},
		},
	}

	return sendCommandTestCase.Run(client, logger)
}
