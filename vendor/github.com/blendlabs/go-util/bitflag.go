package util

func BitFlagAll(reference, value int) bool {
	return reference&value == value
}

func BitFlagAny(reference, value int) bool {
	return reference&value > 0
}

func BitFlagCombine(values ...int) int {
	outputFlag := 0
	for _, value := range values {
		outputFlag = outputFlag | value
	}
	return outputFlag
}
