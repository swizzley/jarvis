package util

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	// StringEmpty is the empty string
	StringEmpty string = ""

	// RuneSpace is a single rune representing a space.
	RuneSpace rune = ' '

	// RuneNewline is a single rune representing a newline.
	RuneNewline rune = '\n'

	// LowerA is the ascii int value for 'a'
	LowerA uint = uint('a')
	// LowerZ is the ascii int value for 'z'
	LowerZ uint = uint('z')
)

var (
	lowerDiff = (LowerZ - LowerA)

	// LowerLetters is a runset of lowercase letters.
	LowerLetters = []rune("abcdefghijklmnopqrstuvwxyz")

	// UpperLetters is a runset of uppercase letters.
	UpperLetters = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")

	// Letters is a runset of both lower and uppercase letters.
	Letters = append(LowerLetters, UpperLetters...)

	// Numbers is a runset of numeric characters.
	Numbers = []rune("0123456789")

	// LettersAndNumbers is a runset of letters and numeric characters.
	LettersAndNumbers = append(Letters, Numbers...)

	// Symbols is a runset of symbol characters.
	Symbols = []rune(`!@#$%^&*()_+-=[]{}\|:;`)

	// LettersNumbersAndSymbols is a runset of letters, numbers and symbols.
	LettersNumbersAndSymbols = append(LettersAndNumbers, Symbols...)
)

var (
	// String is a namesapce for string utility functions.
	String = stringUtil{}
)

type stringUtil struct{}

// IsEmpty returns if a string is empty.
func (su stringUtil) IsEmpty(input string) bool {
	return len(input) == 0
}

// EmptyCoalesce returns the first non-empty string.
func (su stringUtil) EmptyCoalesce(inputs ...string) string {
	for _, input := range inputs {
		if !su.IsEmpty(input) {
			return input
		}
	}
	return StringEmpty
}

// CaseInsensitiveEquals compares two strings regardless of case.
func (su stringUtil) CaseInsensitiveEquals(a, b string) bool {
	aLen := len(a)
	bLen := len(b)
	if aLen != bLen {
		return false
	}

	for x := 0; x < aLen; x++ {
		charA := uint(a[x])
		charB := uint(b[x])

		if charA-LowerA <= lowerDiff {
			charA = charA - 0x20
		}
		if charB-LowerA <= lowerDiff {
			charB = charB - 0x20
		}
		if charA != charB {
			return false
		}
	}

	return true
}

// HasPrefixCaseInsensitive returns if a corpus has a prefix regardless of casing.
func (su stringUtil) HasPrefixCaseInsensitive(corpus, prefix string) bool {
	corpusLen := len(corpus)
	prefixLen := len(prefix)

	if corpusLen < prefixLen {
		return false
	}

	for x := 0; x < prefixLen; x++ {
		charCorpus := uint(corpus[x])
		charPrefix := uint(prefix[x])

		if charCorpus-LowerA <= lowerDiff {
			charCorpus = charCorpus - 0x20
		}

		if charPrefix-LowerA <= lowerDiff {
			charPrefix = charPrefix - 0x20
		}
		if charCorpus != charPrefix {
			return false
		}
	}
	return true
}

// HasSuffixCaseInsensitive returns if a corpus has a suffix regardless of casing.
func (su stringUtil) HasSuffixCaseInsensitive(corpus, suffix string) bool {
	corpusLen := len(corpus)
	suffixLen := len(suffix)

	if corpusLen < suffixLen {
		return false
	}

	for x := 0; x < suffixLen; x++ {
		charCorpus := uint(corpus[corpusLen-(x+1)])
		charSuffix := uint(suffix[suffixLen-(x+1)])

		if charCorpus-LowerA <= lowerDiff {
			charCorpus = charCorpus - 0x20
		}

		if charSuffix-LowerA <= lowerDiff {
			charSuffix = charSuffix - 0x20
		}
		if charCorpus != charSuffix {
			return false
		}
	}
	return true
}

// IsLetter returns if a rune is in the ascii letter range.
func (su stringUtil) IsLetter(c rune) bool {
	return su.IsUpper(c) || su.IsLower(c)
}

// IsUpper returns if a rune is in the ascii upper letter range.
func (su stringUtil) IsUpper(c rune) bool {
	return c >= rune('A') && c <= rune('Z')
}

// IsLower returns if a rune is in the ascii lower letter range.
func (su stringUtil) IsLower(c rune) bool {
	return c >= rune('a') && c <= rune('z')
}

// IsSymbol returns if the rune is in the symbol range.
func (su stringUtil) IsSymbol(c rune) bool {
	return c >= rune(' ') && c <= rune('/')
}

// CombinePathComponents combines string components of a path.
func (su stringUtil) CombinePathComponents(components ...string) string {
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

// RandomString returns a new random string composed of letters from the `letters` collection.
func (su stringUtil) RandomString(length int) string {
	return su.RandomRunes(Letters, length)
}

// RandomNumbers returns a random string of chars from the `numbers` collection.
func (su stringUtil) RandomNumbers(length int) string {
	return su.RandomRunes(Numbers, length)
}

// RandomStringWithNumbers returns a random string composed of chars from the `lettersAndNumbers` collection.
func (su stringUtil) RandomStringWithNumbers(length int) string {
	return su.RandomRunes(LettersAndNumbers, length)
}

// RandomStringWithNumbersAndSymbols returns a random string composed of chars from the `lettersNumbersAndSymbols` collection.
func (su stringUtil) RandomStringWithNumbersAndSymbols(length int) string {
	return su.RandomRunes(LettersNumbersAndSymbols, length)
}

// RandomRunes returns a random selection of runes from the set.
func (su stringUtil) RandomRunes(runeset []rune, length int) string {
	runes := make([]rune, length)
	for index := range runes {
		runes[index] = runeset[provider.Intn(len(runeset))]
	}
	return string(runes)
}

// CombineRunsets combines given runsets into a single runset.
func (su stringUtil) CombineRunsets(runesets ...[]rune) []rune {
	output := []rune{}
	for _, set := range runesets {
		output = append(output, set...)
	}
	return output
}

// IsValidInteger returns if a string is an integer.
func (su stringUtil) IsValidInteger(input string) bool {
	_, convCrr := strconv.Atoi(input)
	return convCrr == nil
}

// IsNumber returns if a string represents a number
func (su stringUtil) IsNumber(input string) bool {
	_, err := strconv.ParseFloat(input, 64)
	return err == nil
}

// RegexMatch returns if a string matches a regexp.
func (su stringUtil) RegexMatch(input string, exp string) string {
	regexp := regexp.MustCompile(exp)
	matches := regexp.FindStringSubmatch(input)
	if len(matches) != 2 {
		return StringEmpty
	}
	return strings.TrimSpace(matches[1])
}

// RegexExtract returns all matches of a regex expr.
func (su stringUtil) RegexExtract(corpus, expr string) []string {
	re := regexp.MustCompile(expr)
	return re.FindAllString(corpus, -1)
}

// RegexExtractSubMatches returns sub matches for an expr because go's regexp library is weird.
func (su stringUtil) RegexExtractSubMatches(corpus, expr string) []string {
	re := regexp.MustCompile(expr)
	allResults := re.FindAllStringSubmatch(corpus, -1)
	results := []string{}
	for _, resultSet := range allResults {
		for _, result := range resultSet {
			results = append(results, result)
		}
	}

	return results
}

// ParseFloat64 parses a float64
func (su stringUtil) ParseFloat64(input string) float64 {
	result, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0.0
	}
	return result
}

// ParseFloat32 parses a float32
func (su stringUtil) ParseFloat32(input string) float32 {
	result, err := strconv.ParseFloat(input, 32)
	if err != nil {
		return 0.0
	}
	return float32(result)
}

// ParseInt parses an int
func (su stringUtil) ParseInt(input string) int {
	result, err := strconv.Atoi(input)
	if err != nil {
		return 0
	}
	return result
}

// ParseInt64 parses an int64
func (su stringUtil) ParseInt64(input string) int64 {
	result, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		return int64(0)
	}
	return result
}

// IntToString turns an int into a string
func (su stringUtil) IntToString(input int) string {
	return strconv.Itoa(input)
}

// Int64ToString turns an int64 into a string
func (su stringUtil) Int64ToString(input int64) string {
	return strconv.FormatInt(input, 10)
}

// Float32ToString turns an float32 into a string
func (su stringUtil) Float32ToString(input float32) string {
	return strconv.FormatFloat(float64(input), 'f', -1, 32)
}

// Float64ToString turns an float64 into a string
func (su stringUtil) Float64ToString(input float64) string {
	return strconv.FormatFloat(input, 'f', -1, 64)
}

// ToCSVOfInt returns a csv from a given slice of integers.
func (su stringUtil) ToCSVOfInt(input []int) string {
	outputStrings := []string{}
	for _, v := range input {
		outputStrings = append(outputStrings, su.IntToString(v))
	}
	return strings.Join(outputStrings, ",")
}

// StripQuotes removes quote characters from a string.
func (su stringUtil) StripQuotes(input string) string {
	output := []rune{}
	for _, c := range input {
		if !(c == '\'' || c == '"') {
			output = append(output, c)
		}
	}
	return string(output)
}

// TrimTo trims a string to a given length.
func (su stringUtil) TrimTo(val string, length int) string {
	if len(val) > length {
		return val[0:length]
	}
	return val
}

// TrimWhitespace trims spaces and tabs from a string.
func (su stringUtil) TrimWhitespace(input string) string {
	return strings.Trim(input, " \t")
}

// IsCamelCase returns if a string is CamelCased.
// CamelCased in this sense is if a string has both upper and lower characters.
func (su stringUtil) IsCamelCase(input string) bool {
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

// Base64Encode returns a base64 string for a byte array.
func (su stringUtil) Base64Encode(blob []byte) string {
	return base64.StdEncoding.EncodeToString(blob)
}

// Base64Decode returns a byte array for a base64 encoded string.
func (su stringUtil) Base64Decode(blob string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(blob)
}

// FormatFileSize returns a string representation of a file size in bytes.
func (su stringUtil) FormatFileSize(sizeBytes int) string {
	if sizeBytes >= 1<<30 {
		return fmt.Sprintf("%dgB", sizeBytes/(1<<30))
	} else if sizeBytes >= 1<<20 {
		return fmt.Sprintf("%dmB", sizeBytes/(1<<20))
	} else if sizeBytes >= 1<<10 {
		return fmt.Sprintf("%dkB", sizeBytes/(1<<10))
	}
	return fmt.Sprintf("%dB", sizeBytes)
}

var nonTitleWords = map[string]bool{
	"and":     true,
	"the":     true,
	"a":       true,
	"an":      true,
	"but":     true,
	"or":      true,
	"on":      true,
	"in":      true,
	"with":    true,
	"for":     true,
	"either":  true,
	"neither": true,
	"nor":     true,
}

func (su stringUtil) ToTitleCase(corpus string) string {
	output := bytes.NewBuffer(nil)
	runes := []rune(corpus)

	haveSeenLetter := false
	var r rune
	for x := 0; x < len(runes); x++ {
		r = runes[x]

		if unicode.IsLetter(r) {
			if !haveSeenLetter {
				output.WriteRune(unicode.ToUpper(r))
				haveSeenLetter = true
			} else {
				output.WriteRune(unicode.ToLower(r))
			}
		} else {
			output.WriteRune(r)
			haveSeenLetter = false
		}
	}
	return output.String()
}
