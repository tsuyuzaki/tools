package main

import (
	"strconv"
	"strings"
)

const methodIdentifierPrefix = "Started "
const actionIdentifier = "Processing by "
const paramsIdentifier = "Parameters: "
const queryIdentifier = "SELECT  "
const resultIdentifier = "Completed "
const limitOne = "LIMIT 1"

var methods = [...]string{"POST", "GET", "PUT", "OPTIONS"}

func replaceNumber(sqlQuery string) string {
	splitted := strings.Split(sqlQuery, " ")
	for i := 0; i < len(splitted); i++ {
		if isNumber(splitted[i]) {
			splitted[i] = "[NUMBER]"
		}
	}
	return strings.Join(splitted, " ")
}

func getStrPartWithPrePostText(line, preText, postText string) (bool, string) {
	ok, str := getStrPartWithPreText(line, preText)
	if !ok {
		return false, ""
	}
	pos := strings.Index(str, postText)
	if pos == -1 {
		return false, ""
	}
	return true, str[:pos]
}

func getStrPartWithPreText(line, preText string) (bool, string) {
	pos := strings.Index(line, preText)
	if pos == -1 {
		return false, ""
	}
	return true, line[pos+len(preText):]
}

func isMethod(line string) bool {
	for _, method := range methods {
		query := methodIdentifierPrefix + method
		if strings.Contains(line, query) {
			return true
		}
	}
	return false
}

func isNumber(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
