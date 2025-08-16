// Package validator provides custom, reusable validation rules for the application.
package validator

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

var (
	IsISBN    = &isbnRule{message: "must be a valid ISBN-10 or ISBN-13"}
	isbnRegex = regexp.MustCompile(`[^0-9X]`)
)

type isbnRule struct {
	message string
}

func (r *isbnRule) Validate(value any) error {
	isbn, ok := value.(string)
	if !ok {
		return errors.New("must be a string")
	}

	cleanISBN := strings.ToUpper(isbnRegex.ReplaceAllString(isbn, ""))

	switch len(cleanISBN) {
	case 10:
		if r.validateISBN10(cleanISBN) {
			return nil
		}
	case 13:
		if r.validateISBN13(cleanISBN) {
			return nil
		}
	}
	return errors.New(r.message)
}

func (r *isbnRule) Error(message string) *isbnRule {
	return &isbnRule{message: message}
}

func (r *isbnRule) validateISBN10(isbn string) bool {
	var sum int
	for i := 0; i < 9; i++ {
		digit, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false
		}
		sum += digit * (i + 1)
	}

	checksum := sum % 11
	lastChar := string(isbn[9])

	if checksum == 10 {
		return lastChar == "X"
	}
	return strconv.Itoa(checksum) == lastChar
}

func (r *isbnRule) validateISBN13(isbn string) bool {
	var sum int
	for i := 0; i < 12; i++ {
		digit, err := strconv.Atoi(string(isbn[i]))
		if err != nil {
			return false
		}
		if (i+1)%2 == 0 {
			sum += digit * 3
		} else {
			sum += digit
		}
	}

	checksum := (10 - (sum % 10)) % 10
	lastDigit, err := strconv.Atoi(string(isbn[12]))
	if err != nil {
		return false
	}

	return checksum == lastDigit
}
