package util

func Ternary(condition bool, ifTrue, ifFalse interface{}) interface{} {
	if condition {
		return ifTrue
	} else {
		return ifFalse
	}
}

func TernaryOfString(condition bool, ifTrue, ifFalse string) string {
	if condition {
		return ifTrue
	} else {
		return ifFalse
	}
}

func TernaryOfInt(condition bool, ifTrue, ifFalse int) int {
	if condition {
		return ifTrue
	} else {
		return ifFalse
	}
}

func TernaryOfFloat(condition bool, ifTrue, ifFalse float64) float64 {
	if condition {
		return ifTrue
	} else {
		return ifFalse
	}
}
