package util

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	EMPTY        = ""
	COLOR_RED    = "31"
	COLOR_BLUE   = "94"
	COLOR_GREEN  = "32"
	COLOR_YELLOW = "33"
	COLOR_WHITE  = "37"
	COLOR_GRAY   = "90"
)

var (
	letters           = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	numbers           = []rune("0123456789")
	lettersAndNumbers = append(letters, numbers...)
)

func IsEmpty(input string) bool {
	return len(input) == 0
}

func EmptyCoalesce(inputs ...string) string {
	for _, input := range inputs {
		if !IsEmpty(input) {
			return input
		}
	}
	return EMPTY
}

func CombinePathComponents(components ...string) string {
	slash := "/"
	fullPath := ""
	for index, component := range components {
		workingComponent := component
		if strings.HasPrefix(workingComponent, slash) {
			workingComponent = strings.TrimPrefix(workingComponent, slash)
		}

		if strings.HasSuffix(workingComponent, slash) {
			workingComponent = strings.TrimSuffix(workingComponent, slash)
		}

		if index != len(components)-1 {
			fullPath = fullPath + workingComponent + slash
		} else {
			fullPath = fullPath + workingComponent
		}
	}
	return fullPath
}

func RandomString(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, length)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func RandomStringWithNumbers(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, length)
	for i := range b {
		b[i] = lettersAndNumbers[r.Intn(len(lettersAndNumbers))]
	}
	return string(b)
}

func RandomNumbers(length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, length)
	for i := range b {
		b[i] = numbers[r.Intn(len(numbers))]
	}
	return string(b)
}

func IsValidInteger(input string) bool {
	_, convCrr := strconv.Atoi(input)
	return convCrr == nil
}

func RegexMatch(string_to_parse string, regexp_string string) string {
	regexp := regexp.MustCompile(regexp_string)
	matches := regexp.FindStringSubmatch(string_to_parse)
	if len(matches) != 2 {
		return EMPTY
	}
	return strings.TrimSpace(matches[1])
}

func ParseFloat64(input string) float64 {
	result, conv_err := strconv.ParseFloat(input, 64)
	if conv_err != nil {
		return 0.0
	} else {
		return result
	}
}

func ParseFloat32(input string) float32 {
	result, conv_err := strconv.ParseFloat(input, 32)
	if conv_err != nil {
		return 0.0
	} else {
		return float32(result)
	}
}

func ParseInt(input string) int {
	result, conv_err := strconv.Atoi(input)
	if conv_err != nil {
		return 0
	} else {
		return result
	}
}

func IntToString(input int) string {
	return strconv.Itoa(input)
}

func Float32ToString(input float32) string {
	return fmt.Sprintf("%v", input)
}

func Float64ToString(input float64) string {
	return fmt.Sprintf("%v", input)
}

func ToCSVOfInt(input []int) string {
	outputStrings := []string{}
	for _, v := range input {
		outputStrings = append(outputStrings, IntToString(v))
	}
	return strings.Join(outputStrings, ",")
}

func StripQuotes(input string) string {
	output := []rune{}
	for _, c := range input {
		if !(c == '\'' || c == '"') {
			output = append(output, c)
		}
	}
	return string(output)
}

func TrimWhitespace(input string) string {
	return strings.Trim(input, " \t")
}

func IsCamelCase(input string) bool {
	hasLowers := false
	hasUppers := false

	for _, c := range input {
		if unicode.IsUpper(c) {
			hasUppers = true
		}
		if unicode.IsLower(c) {
			hasLowers = true
		}
	}

	return hasLowers && hasUppers
}

func Base64Encode(blob []byte) string {
	return base64.StdEncoding.EncodeToString(blob)
}

func Base64Decode(blob string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(blob)
}

func Color(input string, colorCode string) string {
	return fmt.Sprintf("\033[%s;01m%s\033[0m", colorCode, input)
}

func ColorFixedWidth(input string, colorCode string, width int) string {
	fixedToken := fmt.Sprintf("%%%d.%ds", width, width)
	fixedMessage := fmt.Sprintf(fixedToken, input)
	return fmt.Sprintf("\033[%s;01m%s\033[0m", colorCode, fixedMessage)
}

func ColorFixedWidthLeftAligned(input string, colorCode string, width int) string {
	fixedToken := fmt.Sprintf("%%-%ds", width)
	fixedMessage := fmt.Sprintf(fixedToken, input)
	return fmt.Sprintf("\033[%s;01m%s\033[0m", colorCode, fixedMessage)
}
