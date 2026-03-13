package internal

import (
	"github.com/codecrafters-io/redis-tester/internal/redis_executable"
	resp_value "github.com/codecrafters-io/redis-tester/internal/resp/value"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/redis-tester/internal/test_cases"
	"github.com/codecrafters-io/tester-utils/random"
	"github.com/codecrafters-io/tester-utils/test_case_harness"
)

func testPubSubSubscribe3(stageHarness *test_case_harness.TestCaseHarness) error {
	b := redis_executable.NewRedisExecutable(stageHarness)
	if err := b.Run(); err != nil {
		return err
	}

	logger := stageHarness.Logger
	clientsSpawner := ClientsSpawner{
		Addr:         "localhost:6379",
		StageHarness: stageHarness,
	}
	client, err := clientsSpawner.SpawnClientWithPrefix("client")
	if err != nil {
		return err
	}

	channels := random.RandomWords(2)
	subscribeTestCase1 := test_cases.SendCommandTestCase{
		Command:   "SUBSCRIBE",
		Args:      []string{channels[0]},
		Assertion: resp_assertions.NewSubscribeResponseAssertion(channels[0], 1),
	}
	if err := subscribeTestCase1.Run(client, logger); err != nil {
		return err
	}

	/* Test against ECHO/SET/GET (Unallowed commands) */

	/* SET */
	keyAndValue := random.RandomWords(2)
	setTestCase := test_cases.SendCommandTestCase{
		Command: "SET",
		Args:    keyAndValue,
		Assertion: resp_assertions.PrefixAndSubstringsAssertion{
			ExpectedType: resp_value.ERROR,
			PrefixPredicate: &resp_assertions.PrefixPredicate{
				Prefix:        "ERR ",
				CaseSensitive: true,
			},
			HasSubstringPredicates: []resp_assertions.HasSubstringPredicate{{
				Substring:     "can't execute 'set'",
				CaseSensitive: false,
			}},
		},
	}

	if err := setTestCase.Run(client, logger); err != nil {
		return err
	}

	/* GET */
	getTestCase := test_cases.SendCommandTestCase{
		Command: "GET",
		Args:    keyAndValue[1:],
		Assertion: resp_assertions.PrefixAndSubstringsAssertion{
			ExpectedType: resp_value.ERROR,
			PrefixPredicate: &resp_assertions.PrefixPredicate{
				Prefix:        "ERR ",
				CaseSensitive: true,
			},
			HasSubstringPredicates: []resp_assertions.HasSubstringPredicate{{
				Substring:     "can't execute 'get'",
				CaseSensitive: false,
			}},
		},
	}
	if err := getTestCase.Run(client, logger); err != nil {
		return err
	}

	/* ECHO */
	echoTestCase := test_cases.SendCommandTestCase{
		Command: "ECHO",
		Args:    keyAndValue[1:],
		Assertion: resp_assertions.PrefixAndSubstringsAssertion{
			ExpectedType: resp_value.ERROR,
			PrefixPredicate: &resp_assertions.PrefixPredicate{
				Prefix:        "ERR ",
				CaseSensitive: true,
			},
			HasSubstringPredicates: []resp_assertions.HasSubstringPredicate{{
				Substring:     "can't execute 'echo'",
				CaseSensitive: false,
			}},
		},
	}
	if err := echoTestCase.Run(client, logger); err != nil {
		return err
	}

	subscribeTestCase2 := test_cases.SendCommandTestCase{
		Command:   "SUBSCRIBE",
		Args:      []string{channels[1]},
		Assertion: resp_assertions.NewSubscribeResponseAssertion(channels[1], 2),
	}
	return subscribeTestCase2.Run(client, logger)
}
