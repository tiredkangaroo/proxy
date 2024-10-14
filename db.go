package main

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	_ "github.com/lib/pq"
)

// connectDB established a connection to the postgres uri provided
// in env.POSTGRESURI. It provides the connection in env and may
// return an error.
func connectDB() (err error) {
	env.ActiveDB, err = sql.Open("postgres", env.POSTGRESURI)
	return
}

// createProxyRequestTable creates the ProxyRequest table
// in the env.ActiveDB. It may return an error.
func createProxyRequestTable() (err error) {
	_, err = env.ActiveDB.Exec(`
		CREATE TABLE IF NOT EXISTS ProxyRequest (
			id char(30) PRIMARY KEY,
			clientIP text,
			proxyAuthorization text,
			rawHTTPRequest bytea,
			rawHTTPResponse bytea,
			method text,
			url text,
			error text,
			time bigint,
			upstreamResponseTime bigint,
			processingTime bigint
		)
	`)
	return
}

// initalizeDB initialized the database by connecting
// to it and creating all necessary tables if they do
// not yet exist.
func initalizeDB() (err error) {
	err = connectDB()
	if err != nil {
		return
	}
	err = createProxyRequestTable()
	return
}

// saveProxyRequest inserts a proxy request into the ProxyRequest
// table in the ActiveDB.
func saveProxyRequest(values map[string]any) (err error) {
	statement := `INSERT INTO ProxyRequest (`
	statement2 := `(`
	i := 0
	vs := []any{}
	for k, v := range values {
		statement += k
		statement2 += fmt.Sprintf("$%d", i+1)
		if i+1 != len(values) {
			statement += ","
			statement2 += ","
		}
		vs = append(vs, v)
		i += 1
	}
	statement += ") VALUES "
	statement2 += ")"
	statement += statement2
	statement += ";"
	_, err = env.ActiveDB.Exec(statement, vs...)
	return
}

// scanProxyRequest scans a proxy request into a map[string]interface{}.
func scanProxyRequest(rows *sql.Rows) (map[string]interface{}, error) {
	var id, clientIP, proxyAuthorization *string
	var rawHTTPRequest, rawHTTPResponse *[]byte
	var method, proxyURL, proxyError *string
	var proxyTime, upstreamResponseTime, processingTime *int64
	err := rows.Scan(&id, &clientIP, &proxyAuthorization, &rawHTTPRequest, &rawHTTPResponse, &method, &proxyURL, &proxyError, &proxyTime, &upstreamResponseTime, &processingTime)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return map[string]interface{}{
		"id":                   id,
		"clientIP":             clientIP,
		"proxyAuthorization":   proxyAuthorization,
		"rawHTTPRequest":       rawHTTPRequest,
		"rawHTTPResponse":      rawHTTPResponse,
		"method":               method,
		"url":                  proxyURL,
		"error":                proxyError,
		"time":                 proxyTime,
		"processing_time":      processingTime,
		"upstreamResponseTime": upstreamResponseTime,
	}, nil
}

// getProxyRequests retrieves all proxy requests stored on the ProxyRequest
// table in the ActiveDB.
func getProxyRequests() ([]map[string]interface{}, error) {
	rows, err := env.ActiveDB.Query(`SELECT * FROM ProxyRequest;`)
	if err != nil {
		return []map[string]interface{}{}, err
	}
	requests := []map[string]interface{}{}
	for rows.Next() {
		pr, err := scanProxyRequest(rows)
		if err != nil {
			return []map[string]interface{}{}, err
		}
		requests = append(requests, pr)
	}
	return requests, nil
}

// getProxyRequestByID gets a proxy request by its ID.
func getProxyRequestByID(id string) (map[string]interface{}, error) {
	rows, err := env.ActiveDB.Query(`SELECT * FROM ProxyRequest WHERE id=$1;`, id)
	if err != nil {
		return map[string]interface{}{}, err
	}
	for rows.Next() {
		pr, err := scanProxyRequest(rows)
		if err != nil {
			return map[string]interface{}{}, err
		}
		return pr, nil
	}
	return map[string]interface{}{}, fmt.Errorf("no proxy request by id")
}

// deleteProxyRequest deletes a proxy request by its ID.
func deleteProxyRequest(id string) (err error) {
	_, err = env.ActiveDB.Exec("DELETE FROM ProxyRequest WHERE id=$1;", id)
	return
}

// deleteAllProxyRequests deletes all proxy requests in the
// ProxyRequest table in the ActiveDB.
func deleteAllProxyRequests() (err error) {
	_, err = env.ActiveDB.Exec("DELETE FROM ProxyRequest;")
	return
}

// saveBlockedSite saves the blocked site into the Redis cache with key
// proxyblockedsites.
func saveBlockedSite(site *regexp.Regexp) (err error) {
	cmd := env.Client.SAdd(context.TODO(), "proxyblockedsites", site.String())
	return cmd.Err()
}

// deleteBlockedSite deletes the blocked site from the Redis cache with key
// proxyblockedsites.
func deleteBlockedSite(site *regexp.Regexp) (err error) {
	cmd := env.Client.SRem(context.TODO(), "proxyblockedsites", site.String())
	return cmd.Err()
}

// getBlockedSites gets all blocked sites with key "proxyblockedsites".
func getBlockedSites() ([]*regexp.Regexp, error) {
	cmd := env.Client.SMembers(context.TODO(), "proxyblockedsites")
	if cmd.Err() != nil {
		return []*regexp.Regexp{}, nil
	}
	nonRegexSites := cmd.Val()
	sites := []*regexp.Regexp{}
	for _, nonRegexSite := range nonRegexSites {
		site, err := regexp.Compile(nonRegexSite)
		if err != nil {
			return []*regexp.Regexp{}, fmt.Errorf(FailedToRetrieveBlockedSites, err.Error())
		}
		sites = append(sites, site)
	}
	return sites, nil
}
