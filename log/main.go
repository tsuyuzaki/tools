/**
 * golang install (mac)
 * $ brew install go
 *
 * tool の実行
 * go run *.go [log ファイル] [数値 [msec] (省略化)]
 * 数値 [msec] を指定した場合は、その値より遅いリクエストのみ出力します。
 */
package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var logs = requestLogs{}
var optionMethodIDs = map[string]bool{}

func main() {
	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Println("logファイルと数値(msec)を指定してください。")
		fmt.Println("数値(msec)は省略可能で、指定された場合、その値より遅いリクエストのみ出力します。")
		return
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.Open() error [%v]\n", err)
		os.Exit(1)
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		id, ok := getStrPartWithPrePostText(line, " -- : [", "]")
		if !ok {
			continue
		}
		if !fillRequestLog(s, id, line) {
			fmt.Fprintf(os.Stderr, "Cannot fill requestLog line[%s]\n", line)
		}
	}
	msecBoundary := 0
	if len(os.Args) == 3 {
		opt := os.Args[2]
		if log, ok := logs[opt]; ok {
			fmt.Println("hogehoge")
			log.print()
			return
		}
		msecBoundary, _ = strconv.Atoi(opt)
	}
	logs.print(msecBoundary)
}

func fillRequestLog(s *bufio.Scanner, id, line string) bool {
	if optionMethodIDs[id] {
		return true
	}

	log := logs.get(id)
	if isMethod(line) {
		return fillMethod(log, id, line)
	} else if strings.Contains(line, actionIdentifier) {
		return fillStringValue(line, &(log.action), actionIdentifier)
	} else if strings.Contains(line, paramsIdentifier) {
		return fillStringValue(line, &(log.params), paramsIdentifier)
	} else if strings.Contains(line, "Rendered ") && strings.Contains(line, "jbuilder") {
		return fillStringValue(line, &(log.jbuilder), "Rendered ")
	} else if strings.Contains(line, queryIdentifier) {
		return appendSqlQuery(log, line)
	} else if strings.Contains(line, resultIdentifier) {
		return fillResult(log, line)
	} else if strings.Contains(line, "WARN ") && strings.Contains(line, "user: root") {
		return fillEagerLoadLogs(s, log)
	}

	return true
}

func fillMethod(log *requestLog, id, line string) bool {
	method, ok := getStrPartWithPrePostText(line, methodIdentifierPrefix, " for ")
	if !ok {
		return false
	}
	if strings.Contains(method, "OPTIONS") {
		logs.delete(id)
		optionMethodIDs[id] = true
		return true
	}
	log.method = method
	return true
}

func fillStringValue(line string, value *string, preText string) bool {
	got, ok := getStrPartWithPreText(line, preText)
	if ok {
		*value = got
	}
	return ok
}

func appendSqlQuery(log *requestLog, line string) bool {
	msecStr, ok := getStrPartWithPrePostText(line, " (", "ms)")
	if !ok {
		return false
	}
	msec, err := strconv.ParseFloat(msecStr, 64)
	if err != nil {
		return false
	}
	str, ok := getStrPartWithPreText(line, queryIdentifier)
	if !ok {
		return false
	}
	sql := queryIdentifier + str
	sqlQuery := &sqlQuery{msec, sql}
	log.sqlQueries = append(log.sqlQueries, sqlQuery)

	if !strings.Contains(sql, limitOne) {
		return true
	}

	lqs := log.limitOneQueries
	if lqs == nil {
		lqs = map[string]int{}
		log.limitOneQueries = lqs
	}
	lqs[replaceNumber(sql)]++
	return true
}

func fillResult(log *requestLog, line string) bool {
	res, ok := getStrPartWithPreText(line, resultIdentifier)
	if !ok {
		return false
	}
	log.result = res
	msecStr, ok := getStrPartWithPrePostText(res, " in ", "ms ")
	if !ok {
		return false
	}
	msec, err := strconv.Atoi(msecStr)
	if err != nil {
		fmt.Println(os.Stderr, "strconv.Atoi error [%v] line[%s]\n", err, line)
		return false
	}
	log.msec = msec
	return true
}

func fillEagerLoadLogs(s *bufio.Scanner, log *requestLog) bool {
	eagerLoadLog := ""
	for s.Scan() {
		line := s.Text()
		if strings.Contains(line, "Call stack") {
			break
		}
		if eagerLoadLog != "" {
			eagerLoadLog += "\n"
		}
		eagerLoadLog += line
	}
	log.eagerLoadLogs = append(log.eagerLoadLogs, eagerLoadLog)
	return true
}
