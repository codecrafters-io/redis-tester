package test_cases

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}
