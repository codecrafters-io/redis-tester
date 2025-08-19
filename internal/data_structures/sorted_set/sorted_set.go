package sorted_set

import (
	"sort"

	testerutils_random "github.com/codecrafters-io/tester-utils/random"
)

type SortedSetMember struct {
	Name  string
	Score float64
}

// SortedSet is a data structure that maintains its elements sorted by score.
// If multiple elements have the same score, they are ordered lexicographically.
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
		if m.Name == name {
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
		memberNames[i] = m.Name
	}
	return memberNames
}

type SortedSetMemberGenerationOption struct {
	Count          int // Total number of members to generate
	SameScoreCount int // Number of members with same score (for testing lexicographic sorting)
}

func GenerateSortedSetWithRandomMembers(option SortedSetMemberGenerationOption) *SortedSet {
	count := option.Count
	sameScoreCount := min(option.SameScoreCount, count)
	differentScoresCount := count - sameScoreCount

	ss := NewSortedSet()

	memberNames := testerutils_random.RandomWords(count)

	// generate members with different scores
	for i := range differentScoresCount {
		score := GetRandomSortedSetScore()
		ss.AddMember(SortedSetMember{
			Name:  memberNames[i],
			Score: score,
		})
	}

	// generate members with same score
	baseScore := GetRandomSortedSetScore()
	for i := range sameScoreCount {
		ss.AddMember(SortedSetMember{
			Name:  memberNames[differentScoresCount+i],
			Score: baseScore,
		})
	}

	return ss
}

// GetRandomSortedSetScore returns a random value of score for a sorted set
func GetRandomSortedSetScore() float64 {
	return testerutils_random.RandomFloat64(1, 100)
}

// sort orders members by ascending value of score
// if scores are same, the members are sorted lexicographically
func (ss *SortedSet) sort() {
	sort.Slice(ss.members, func(i, j int) bool {
		if ss.members[i].Score != ss.members[j].Score {
			return ss.members[i].Score < ss.members[j].Score
		}
		return ss.members[i].Name < ss.members[j].Name
	})
}
