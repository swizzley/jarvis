package collections

import (
	"strings"

	"github.com/wcharczuk/jarvis-cli/Godeps/_workspace/src/github.com/blendlabs/go-util"
)

type StringArray []string

func (sa StringArray) Contains(elem string) bool {
	for _, arrayElem := range sa {
		if arrayElem == elem {
			return true
		}
	}
	return false
}

func (sa StringArray) ContainsLower(elem string) bool {
	for _, arrayElem := range sa {
		if strings.ToLower(arrayElem) == elem {
			return true
		}
	}
	return false
}

func (sa StringArray) GetByLower(elem string) string {
	for _, arrayElem := range sa {
		if strings.ToLower(arrayElem) == elem {
			return arrayElem
		}
	}
	return util.EMPTY
}
