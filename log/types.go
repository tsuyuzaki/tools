package main

import (
	"fmt"
	"sort"
	"strings"
)

type sqlQuery struct {
	msec float64 // SQL クエリーの実行時間。
	sql  string  // SQL クエリー。
}

type requestLog struct {
	reqID           string         // request ID。
	msec            int            // request にかかったトータル時間。
	method          string         // OPITONS を除いたメソッド情報。 ex) POST "/deals"
	action          string         // controller のアクション。
	params          string         // リクエストで指定されたパラメータ。
	sqlQueries      []*sqlQuery    // リクエストで処理された SQL クエリー情報。
	limitOneQueries map[string]int // LIMIT 1 指定された SQL クエリーと、その回数。クエリーの整数値部分は "[NUMBER]" という文字に置き換えています。
	jbuilder        string         // jbuilder 情報。
	result          string         // status code と View, ActiveRecord 処理にかかった時間の情報。
	eagerLoadLogs   []string       // eager loading のログ。
}

type requestLogs map[string]*requestLog

type sortableRequestLogs []*requestLog // requestLog の msec でソートするための配列。

/*** requestLogs's methods ***/
func (logs requestLogs) get(id string) *requestLog {
	log := logs[id]
	if log == nil {
		log = &requestLog{reqID: id}
		logs[id] = log
	}
	return log
}

func (logs requestLogs) delete(id string) {
	delete(logs, id)
}

func (logs requestLogs) print(msecBoundary int) {
	for _, log := range logs.toSortedArray() {
		if log.msec <= msecBoundary {
			return
		}
		log.print()
	}
}

func (logs requestLogs) toSortedArray() sortableRequestLogs {
	sortableLogs := sortableRequestLogs{}
	for _, log := range logs {
		sortableLogs = append(sortableLogs, log)
	}
	sort.Sort(sortableLogs)
	return sortableLogs
}

/*** requestLog's methods ***/
func (log *requestLog) print() {
	fmt.Printf("\n===TOTAL TIME[%d msec] req-id[%s]\n", log.msec, log.reqID)
	line := log.method
	if log.params != "" {
		line += "," + log.params
	}
	fmt.Println(line)

	fmt.Println("  " + log.action)
	if log.jbuilder != "" {
		fmt.Println("  " + log.jbuilder)
	}
	if log.result != "" {
		fmt.Println("  " + log.result)
	}
	log.printQueries()
	if len(log.eagerLoadLogs) != 0 {
		fmt.Println("### EAGER LOAD ###")
	}
	for _, l := range log.eagerLoadLogs {
		fmt.Println(l)
	}
}

func (log *requestLog) printQueries() {
	alreadyLogged := map[string]bool{}

	for _, q := range log.sqlQueries {
		cnt := 1
		sql := q.sql
		if strings.Contains(sql, limitOne) {
			sql = replaceNumber(sql)
			if alreadyLogged[sql] {
				continue
			}
			alreadyLogged[sql] = true
			cnt = log.limitOneQueries[sql]
		}
		line := fmt.Sprintf("  (%.1f msec),", q.msec)
		if cnt == 1 {
			line += q.sql
		} else {
			line += fmt.Sprintf("%s [CALLED %d TIMES]", sql, cnt)
		}
		fmt.Println(line)
	}
}

/*** sortableRequestLogs's methods ***/
func (logs sortableRequestLogs) Len() int {
	return len(logs)
}

func (logs sortableRequestLogs) Swap(i, j int) {
	logs[i], logs[j] = logs[j], logs[i]
}

func (logs sortableRequestLogs) Less(i, j int) bool {
	return logs[i].msec > logs[j].msec
}
