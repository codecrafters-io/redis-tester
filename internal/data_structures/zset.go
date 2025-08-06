package data_structures

import (
	"sort"
	"strconv"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

type SortedSetMember struct {
	name  string
	score float64
}

func NewSortedSetMember(name string, score float64) SortedSetMember {
	return SortedSetMember{
		name:  name,
		score: score,
	}
}

func (m SortedSetMember) GetName() string {
	return m.name
}

func (m SortedSetMember) GetScore() float64 {
	return m.score
}

/* SortedSet is a data structure that offers automatically sorting of its elements based on their scores
If the scores are same, the elements are sorted lexicographically
*/

type SortedSet struct {
	members []SortedSetMember
}

func NewSortedSet() *SortedSet {
	return &SortedSet{}
}

func (ss *SortedSet) AddMember(m SortedSetMember) *SortedSet {
	ss.members = append(ss.members, m)
	ss.sort()
	return ss
}

func (ss *SortedSet) RemoveMember(name string) *SortedSet {
	for i, m := range ss.members {
		if m.name == name {
			ss.members = append(ss.members[:i], ss.members[i+1:]...)
			return ss
		}
	}
	return ss
}

func (ss *SortedSet) Size() int {
	return len(ss.members)
}

// GetMembers returns the a copy of all the members
func (ss *SortedSet) GetMembers() []SortedSetMember {
	members := make([]SortedSetMember, len(ss.members))
	copy(members, ss.members)
	return members
}

// GetMemberNames returns a slice containing all member names
func (ss *SortedSet) GetMemberNames() []string {
	memberNames := make([]string, len(ss.members))
	for i, m := range ss.members {
		memberNames[i] = m.name
	}
	return memberNames
}

type ZsetMemberGenerationOption struct {
	Count          int // Total number of members to generate
	SameScoreCount int // Number of members with same score (for testing lexicographic sorting)
}

func GenerateZsetWithRandomMembers(option ZsetMemberGenerationOption) *SortedSet {
	count := option.Count
	sameScoreCount := min(option.SameScoreCount, count)
	differentScoresCount := count - sameScoreCount

	ss := NewSortedSet()

	memberNames := testerutils_random.RandomWords(count)

	// generate members with different scores
	for i := range differentScoresCount {
		score := GetRandomZSetScore()
		ss.AddMember(SortedSetMember{
			name:  memberNames[i],
			score: score,
		})
	}

	// generate members with same score
	baseScore := GetRandomZSetScore()
	for i := range sameScoreCount {
		ss.AddMember(SortedSetMember{
			name:  memberNames[differentScoresCount+i],
			score: baseScore,
		})
	}

	return ss
}

// GetRandomZSetScore returns a random value of score for a sorted set
// We clip digits after 12 decimal places so there are no inconsistencies in tests
// I'll remove this comment later. The cause was: https://github.com/codecrafters-io/redis-tester/actions/runs/16774410706/job/47496959864
// I tried testing this in isolation but couldn't reproduce the bug
// I think we should change this logic in tester utils itself if we were to do this
func GetRandomZSetScore() float64 {
	raw := testerutils_random.RandomFloat64(1, 100)
	clippedStr := strconv.FormatFloat(raw, 'f', 12, 64)
	clippedFloat, _ := strconv.ParseFloat(clippedStr, 64)
	return clippedFloat
}

// sort orders members by ascending value of score
// if scores are same, the members are sorted lexicographically
func (ss *SortedSet) sort() {
	sort.Slice(ss.members, func(i, j int) bool {
		if ss.members[i].score != ss.members[j].score {
			return ss.members[i].score < ss.members[j].score
		}
		return ss.members[i].name < ss.members[j].name
	})
}
