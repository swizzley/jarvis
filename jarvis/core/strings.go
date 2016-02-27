package core

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"

	"github.com/blendlabs/go-util"
)

// Random returns a random selection from the input messages slice.
func Random(messages []string) string {
	return messages[rand.Intn(len(messages))]
}

// IsDM returns if a channelID is a DM.
func IsDM(channelID string) bool {
	return strings.HasPrefix(channelID, "D")
}

// IsChannel returns if a channelID is a channel.
func IsChannel(channelID string) bool {
	return strings.HasPrefix(channelID, "C")
}

// IsUserMention returns if a message is a mention for a given userID.
func IsUserMention(message, userID string) bool {
	return Like(message, fmt.Sprintf("<@%s>", userID))
}

// IsMention returns if a message has a mention.
func IsMention(message string) bool {
	return Like(message, "<@(.*)>")
}

// IsSalutation returns if a message has a greeting in it.
func IsSalutation(message string) bool {
	return LikeAny(message, "^hello", "^hi", "^greetings", "^hey", "^yo", "^sup")
}

// IsAsking returns if a message is asking a question.
func IsAsking(message string) bool {
	return LikeAny(message, "would it be possible", "can you", "would you", "is it possible", "([^.?!]*)\\?")
}

// IsPolite returns if a message is polite.
func IsPolite(message string) bool {
	return LikeAny(message, "please", "thanks")
}

// IsVulgar returns if a message is vulgar.
func IsVulgar(message string) bool {
	return LikeAny(message, "fuck", "shit", "ass", "cunt") //yep.
}

// IsAngry returns if a message is angry.
func IsAngry(message string) bool {
	return LikeAny(message, "stupid", "worst", "terrible", "horrible", "cunt", "suck", "awful", "asinine") //yep.
}

// LessMentions removes mentions from a message.
func LessMentions(message string) string {
	output := ""
	state := 0
	for _, c := range message {
		switch state {
		case 0:
			if c == rune("<"[0]) {
				state = 1
			} else {
				output = output + string(c)
			}
		case 1:
			if c == rune(">"[0]) {
				state = 2
			}
		case 2:
			if c == rune(":"[0]) { //chomp one more char
				state = 2
			} else if c == rune(" "[0]) {
				state = 0
			} else {
				state = 0
				output = output + string(c)
			}
		}
	}
	return output
}

// LessSpecificMention removes a specific mention from a message.
func LessSpecificMention(message, userID string) string {
	output := ""
	workingUserID := ""
	tagBuffer := ""
	state := 0
	for _, c := range message {
		switch state {
		case 0:
			if c == rune("<"[0]) {
				state = 1
				tagBuffer = "<"
			} else {
				output = output + string(c)
			}
		case 1:
			if c == rune("@"[0]) {
				tagBuffer = tagBuffer + string(c)
				state = 2
			} else {
				state = 0
				output = output + tagBuffer + string(c)
				tagBuffer = ""
			}
		case 2:
			tagBuffer = tagBuffer + string(c)
			if c == rune(">"[0]) {
				if workingUserID != userID {
					state = 0
					output = output + tagBuffer
				} else {
					state = 3
				}
				workingUserID = ""
				tagBuffer = ""
			}
			workingUserID = workingUserID + string(c)
		case 3:
			if c != rune(" "[0]) && c != rune(":"[0]) {
				state = 0
				output = output + string(c)
			}
		}
	}
	return output
}

// RemoveTags removes tags from a message.
func RemoveTags(message string) string {
	output := ""
	for _, c := range message {
		if !(c == rune("<"[0]) || c == rune(">"[0])) {
			output = output + string(c)
		}
	}
	return output
}

// FixLinks removes the weird slack specific link syntax.
func FixLinks(message string) string {
	output := ""
	state := 0
	tagBuffer := ""
	for _, c := range message {
		switch state {
		case 0: //normal text
			if c == rune("<"[0]) {
				state = 1
				tagBuffer = "<"
				break
			}
			output = output + string(c)
		case 1:
			tagBuffer = tagBuffer + string(c)
			if c == rune("|"[0]) {
				state = 2
				break
			} else if c == rune(">"[0]) {
				state = 0
				output = output + tagBuffer
				break
			}
		case 2:
			if c == rune(">"[0]) {
				state = 0
				break
			}
			output = output + string(c)
		}
	}
	return output
}

// LessFirstWord removes the first word from a message.
func LessFirstWord(message string) string {
	queryPieces := strings.Split(message, " ")[1:]
	return strings.Join(queryPieces, " ")
}

// FirstWord returns the first word from a message.
func FirstWord(message string) string {
	pieces := strings.Split(message, " ")
	return pieces[0]
}

// LastWord returns the last word in a message.
func LastWord(message string) string {
	pieces := strings.Split(message, " ")
	if len(pieces) != 0 {
		return pieces[len(pieces)-1]
	}
	return util.EMPTY
}

// Like returns if a corpus matches a given regex expr.
func Like(corpus, expr string) bool {
	if !strings.HasPrefix(expr, "(?i)") {
		expr = "(?i)" + expr
	}
	matched, _ := regexp.Match(expr, []byte(corpus))
	return matched
}

// Extract returns all matches of a regex expr.
func Extract(corpus, expr string) []string {
	re := regexp.MustCompile(expr)
	return re.FindAllString(corpus, -1)
}

// ExtractSubMatches returns sub matches for an expr because go's regexp library is weird.
func ExtractSubMatches(corpus, expr string) []string {
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

// LikeAny returns true if any of the regex exprs match the corpus.
func LikeAny(corpus string, exprs ...string) bool {
	for _, expr := range exprs {
		if Like(corpus, expr) {
			return true
		}
	}
	return false
}

// EqualsAny returns true if any of the values equal a given value.
func EqualsAny(value string, values ...string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

// ReplaceAny replaces a given set of values in a corpus with a given replacement.
func ReplaceAny(corpus string, replacement string, values ...string) string {
	output := strings.ToLower(corpus)

	for _, thing := range values {
		output = strings.Replace(output, strings.ToLower(thing), strings.ToLower(replacement), -1)
	}

	return output
}

func Mentions(corpus string) []string {
	return ExtractSubMatches(corpus, "<@(.*)>")
}
