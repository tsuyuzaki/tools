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

type pair struct { // request の msec でソートするための構造体。
	key   string
	value *requestLog
}

type pairs []*pair // request の msec でソートするための配列。

/*** requestLogs's methods ***/
func (logs requestLogs) get(id string) *requestLog {
	log := logs[id]
	if log == nil {
		log = &requestLog{}
		logs[id] = log
	}
	return log
}

func (logs requestLogs) delete(id string) {
	delete(logs, id)
}

func (logs requestLogs) print(msecBoundary int) {
	for _, p := range logs.toSortedArray() {
		if p.value.msec <= msecBoundary {
			return
		}
		p.value.print(p.key)
	}
}

func (logs requestLogs) toSortedArray() pairs {
	ps := pairs{}
	for key, value := range logs {
		ps = append(ps, &pair{key, value})
	}
	sort.Sort(ps)
	return ps
}

/*** requestLog's methods ***/
func (log *requestLog) print(reqID string) {
	fmt.Printf("\n===TOTAL TIME[%d msec] req-id[%s]\n", log.msec, reqID)
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
			cnt = log.limitOneQueries[sql]
			if alreadyLogged[sql] {
				continue
			} else {
				alreadyLogged[sql] = true
			}
		}
		fmt.Printf("  (%.1f msec),%s", q.msec, sql)
		if cnt > 1 {
			fmt.Printf(" [CALLED %d TIMES]", cnt)
		}
		fmt.Printf("\n")
	}
}

/*** pairs's methods ***/
func (ps pairs) Len() int {
	return len(ps)
}

func (ps pairs) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

func (ps pairs) Less(i, j int) bool {
	return ps[i].value.msec > ps[j].value.msec
}
