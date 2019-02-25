package readordie

import (
	"log"
	"regexp"
)

// RemoveNonAlphanumerical removes non alphanumerical characters from a string
func RemoveNonAlphanumerical(input string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9 ]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(input, "")
}