package test_cases

import (
	"sort"
	"strconv"

	"github.com/codecrafters-io/redis-tester/internal/instrumented_resp_connection"
	"github.com/codecrafters-io/redis-tester/internal/resp_assertions"
	"github.com/codecrafters-io/tester-utils/logger"
)

type zsetMember struct {
	Score float64
	Name  string
}

type ZsetTestCase struct {
	members []zsetMember
	key     string
}

func NewZsetTestCase(key string) *ZsetTestCase {
	return &ZsetTestCase{
		key: key,
	}
}

func (t *ZsetTestCase) AddMember(name string, score float64) *ZsetTestCase {
	t.members = append(t.members, zsetMember{
		Name:  name,
		Score: score,
	})
	return t
}

func (t *ZsetTestCase) RemoveMember(name string) *ZsetTestCase {
	for i, m := range t.members {
		if m.Name == name {
			t.members = append(t.members[:i], t.members[i+1:]...)
		}
	}
	return t
}

func (t *ZsetTestCase) RunZaddAll(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	zaddTestCase := MultiCommandTestCase{
		CommandWithAssertions: make([]CommandWithAssertion, len(t.members)),
	}
	for i, m := range t.members {
		scoreStr := strconv.FormatFloat(m.Score, 'f', -1, 64)
		zaddTestCase.CommandWithAssertions[i] = CommandWithAssertion{
			Command:   []string{"ZADD", t.key, scoreStr, m.Name},
			Assertion: resp_assertions.NewIntegerAssertion(1),
		}
	}
	return zaddTestCase.RunAll(client, logger)
}

func (t *ZsetTestCase) RunZrankAll(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger) error {
	sortedMembers := sortZsetMembers(t.members)
	zrangeTestCase := MultiCommandTestCase{
		CommandWithAssertions: make([]CommandWithAssertion, len(sortedMembers)),
	}
	for i, m := range sortedMembers {
		zrangeTestCase.CommandWithAssertions[i] = CommandWithAssertion{
			Command:   []string{"ZRANK", t.key, m.Name},
			Assertion: resp_assertions.NewIntegerAssertion(i),
		}
	}
	return zrangeTestCase.RunAll(client, logger)
}

func (t *ZsetTestCase) RunZrange(client *instrumented_resp_connection.InstrumentedRespConnection, logger *logger.Logger, startIdx int, endIdx int) error {
	var assertion resp_assertions.RESPAssertion

	cardinality := len(t.members)

	/* Translate */
	if startIdx < 0 {
		startIdx += cardinality
	}
	if endIdx < 0 {
		endIdx += cardinality
	}

	if startIdx > endIdx || startIdx >= cardinality {
		assertion = resp_assertions.NewOrderedArrayAssertion(nil)
	} else {
		/* Clip */
		if endIdx >= cardinality {
			endIdx = cardinality - 1
		}
		if startIdx < 0 {
			startIdx = 0
		}
		// we can do a if (dirty) then sort logic here
		// dirty flag will be set whenever a member is added or removed
		// i'll remove this comment on non-draft PR
		sortedMembers := sortZsetMembers(t.members)

		expectedArrayLen := endIdx - startIdx + 1
		expectedArray := make([]string, expectedArrayLen)
		for i := range expectedArrayLen {
			expectedArray[i] = sortedMembers[startIdx+i].Name
		}

		assertion = resp_assertions.NewOrderedStringArrayAssertion(expectedArray)
	}

	sendCommandTestCase := SendCommandTestCase{
		Command:   "ZRANGE",
		Args:      []string{t.key, strconv.Itoa(startIdx), strconv.Itoa(endIdx)},
		Assertion: assertion,
	}

	return sendCommandTestCase.Run(client, logger)
}

// sortZsetMembers returns a new slice of ZsetMember sorted by score
// If the score is same, they are sorted lexicographically
func sortZsetMembers(members []zsetMember) []zsetMember {
	sortedMembers := make([]zsetMember, len(members))
	copy(sortedMembers, members)

	sort.Slice(sortedMembers, func(i, j int) bool {
		si := sortedMembers[i].Score
		sj := sortedMembers[j].Score
		if si != sj {
			return si < sj
		}
		return sortedMembers[i].Name < sortedMembers[j].Name
	})

	return sortedMembers
}
