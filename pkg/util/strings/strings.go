/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package strings

import (
	"regexp"
	"strings"
)

const (
	lettersMatch       = "[a-zA-Z]+"
	numbersMatch       = "[0-9]+"
	signedNumbersMatch = "^[-+]?[0-9]+"
)

// Contains returns true, if the specified slice contains the specified element, false otherwise
func Contains(slice []string, s string) bool {
	for _, elem := range slice {
		if elem == s {
			return true
		}
	}
	return false
}

func EscapeSingleQuote(s string) string {
	r := strings.NewReplacer(`'`, `'"'"'`)
	return r.Replace(s)
}

func checkForMatch(matcher string, strValue string) bool {
	match := regexp.MustCompile(matcher)
	return match.FindString(strValue) != ""
}

// HasLetters returns true when string contains a letter [a-zA-Z]
func HasLetters(yourString string) bool {
	return checkForMatch(lettersMatch, yourString)
}

// HasOnlyLetters returns true when string contains only letters
func HasOnlyLetters(yourString string) bool {
	return checkForMatch(lettersMatch, yourString) &&
		!checkForMatch(numbersMatch, yourString)
}

// HasLetters returns true when string contains a letter [0-9]
func HasNumbers(yourString string) bool {
	return checkForMatch(numbersMatch, yourString)
}

// HasOnlyNumbers returns true when string contains only numbers
func HasOnlyNumbers(yourString string) bool {
	return checkForMatch(numbersMatch, yourString) &&
		!checkForMatch(lettersMatch, yourString)
}

func getOnlyMatch(matcher string, strValue string) string {
	reg, _ := regexp.Compile(matcher)
	return reg.FindString(strValue)
}

// GetOnlyLetters returns a string containing only letters from given string
func GetOnlyLetters(yourString string) string {
	return getOnlyMatch(lettersMatch, yourString)
}

// GetOnlyNumbers returns a string containing only numbers from given string
func GetOnlyNumbers(yourString string) string {
	return getOnlyMatch(numbersMatch, yourString)
}

// GetSignedNumbers returns a string containing only positive and negative numbers from given string
func GetSignedNumbers(yourString string) string {
	return getOnlyMatch(signedNumbersMatch, yourString)
}
