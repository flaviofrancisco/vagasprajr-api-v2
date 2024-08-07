package commons

import (
	"math/rand"
	"errors"
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

func ValidatePassword(password string) error {
	var (
		lowercaseLetter = regexp.MustCompile(`[a-z]`)
		uppercaseLetter = regexp.MustCompile(`[A-Z]`)
		digit           = regexp.MustCompile(`[0-9]`)
		specialChar     = regexp.MustCompile(`[\W]`)
	)

	if len(password) < 10 ||
		!lowercaseLetter.MatchString(password) ||
		!uppercaseLetter.MatchString(password) ||
		!digit.MatchString(password) ||
		!specialChar.MatchString(password) {
		return errors.New("sua senha deve: conter no mínimo 10 caracteres, conter pelo menos uma letra maiúscula, uma letra minúscula, um número e um caractere especial")
	}

	return nil
}

func GetValidationToken() string {
	
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	b := make([]rune, 32)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}

	return string(b)
}