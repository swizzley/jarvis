package jarvis

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

func Random(messages []string) string {
	return messages[rand.Intn(len(messages))]
}

func IsDM(channelId string) bool {
	return strings.HasPrefix(channelId, "D")
}

func IsChannel(channelId string) bool {
	return strings.HasPrefix(channelId, "C")
}

func IsMention(message, botId string) bool {
	return Like(message, fmt.Sprintf("<@%s>", botId))
}

func IsSalutation(message string) bool {
	return LikeAny(message, []string{"^hello", "^hi", "^greetings", "^hey", "^yo"})
}

func IsAsking(message string) bool {
	return LikeAny(message, []string{"would it be possible", "can you", "would you", "is it possible", "([^.?!]*)\\?"})
}

func IsPolite(message string) bool {
	return LikeAny(message, []string{"please", "thanks"})
}

func IsVulgar(message string) bool {
	return LikeAny(message, []string{"fuck", "shit", "ass", "cunt"}) //yep.
}

func IsAngry(message string) bool {
	return LikeAny(message, []string{"stupid", "worst", "terrible", "horrible", "cunt", "suck", "awful", "asinine"}) //yep.
}

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
			if c == rune(":"[0]) {
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

func LessSpecificMention(message, userId string) string {
	output := ""
	workingUserId := ""
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
			if c == rune(":"[0]) {
				state = 2
			} else {
				state = 0
				output = output + string(c)
			}
		case 2:
			if c == rune(">"[0]) {
				if workingUserId != userId {
					state = 0
					output = output + fmt.Sprintf("<:%s>", workingUserId)
				} else {
					state = 3
				}
				workingUserId = ""
			} else {
				workingUserId = workingUserId + string(c)
			}
		case 3:
			state = 0
			if c != rune(" "[0]) {
				output = output + string(c)
			}
		}
	}
	return output
}

func ExtractTags(message string) string {
	output := ""
	for _, c := range message {
		if !(c == rune("<"[0]) || c == rune(">"[0])) {
			output = output + string(c)
		}
	}
	return output
}

func LessFirst(message string) string {
	queryPieces := strings.Split(message, " ")[1:]
	return strings.Join(queryPieces, " ")
}

func Last(message string) string {
	pieces := strings.Split(message, " ")
	if len(pieces) != 0 {
		return pieces[len(pieces)-1]
	} else {
		return ""
	}
}

func Like(corpus, expr string) bool {
	if !strings.HasPrefix(expr, "(?i)") {
		expr = "(?i)" + expr
	}
	matched, _ := regexp.Match(expr, []byte(corpus))
	return matched
}

func LikeAny(corpus string, exprs []string) bool {
	for _, expr := range exprs {
		if Like(corpus, expr) {
			return true
		}
	}
	return false
}

func Any(value string, values []string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
