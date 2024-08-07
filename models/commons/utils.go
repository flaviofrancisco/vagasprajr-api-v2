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
	local_time := time.Now()
	gmt_location := time.FixedZone("GMT", -3*60*60)
	gmt_time := local_time.In(gmt_location)
	return gmt_time
}