package assistant

func resolvedSearchLimit(value *int) int {
	if value == nil {
		return 5
	}
	if *value < 1 {
		return 1
	}
	if *value > 10 {
		return 10
	}
	return *value
}

func resolvedSearchScoreThreshold(value *float64) float64 {
	if value == nil {
		return 0.78
	}
	if *value < 0 {
		return 0
	}
	if *value > 1 {
		return 1
	}
	return *value
}