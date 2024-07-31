package commons

import "regexp"

func HandleValueForRegex(value string) string {

	escapedPart := regexp.QuoteMeta(value)

	return escapedPart
}