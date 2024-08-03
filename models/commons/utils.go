package commons

import (
	"regexp"
	"time"
)

func HandleValueForRegex(value string) string {

	escapedPart := regexp.QuoteMeta(value)

	return escapedPart
}

func GetBrasiliaTime() time.Time {
	currentTime := time.Now()
	currentTime = currentTime.Add(-3 * time.Hour)
	return currentTime
}