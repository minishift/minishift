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
	"bufio"
	"regexp"
	"strings"
)

const (
	lettersMatch       = "[a-zA-Z]+"
	numbersMatch       = "[0-9]+"
	signedNumbersMatch = "^[-+]?[0-9]+"
	punctuationMatch   = "[.,/#!$%^&*;:{}=-_`~()]"
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

// Remove takes a slice of strings and returns a slice with the first occurrence of the string value removed.
func Remove(slice []string, value string) []string {
	for i, s := range slice {
		if s == value {
			slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}
	return slice
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
		!checkForMatch(numbersMatch, yourString) &&
		!checkForMatch(punctuationMatch, yourString)
}

// HasNumbers returns true when string contains a letter [0-9]
func HasNumbers(yourString string) bool {
	return checkForMatch(numbersMatch, yourString)
}

// HasOnlyNumbers returns true when string contains only numbers
func HasOnlyNumbers(yourString string) bool {
	return checkForMatch(numbersMatch, yourString) &&
		!checkForMatch(lettersMatch, yourString) &&
		!checkForMatch(punctuationMatch, yourString)
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

// SplitAndTrim split the string based on the separator passed
func SplitAndTrim(s string, separator string) ([]string, error) {
	// Trims the spaces and then splits
	trimmed := strings.TrimSpace(s)
	split := strings.Split(trimmed, separator)
	cleanSplit := make([]string, len(split))
	for i, val := range split {
		cleanSplit[i] = strings.TrimSpace(val)
	}

	return cleanSplit, nil
}

func ConvertSlashes(input string) string {
	return strings.Replace(input, "\\", "/", -1)
}

func ParseLines(stdout string) []string {
	resp := []string{}

	s := bufio.NewScanner(strings.NewReader(stdout))
	for s.Scan() {
		resp = append(resp, s.Text())
	}

	return resp
}
