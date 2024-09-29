package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func connectDB() (err error) {
	env.ActiveDB, err = sql.Open("postgres", env.POSTGRESURI)
	return
}

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

func initalizeDB() (err error) {
	err = connectDB()
	if err != nil {
		return
	}
	err = createProxyRequestTable()
	return
}

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

func getProxyRequests() ([]map[string]interface{}, error) {
	rows, err := env.ActiveDB.Query(`SELECT * FROM ProxyRequest;`)
	if err != nil {
		return []map[string]interface{}{}, err
	}
	requests := []map[string]interface{}{}
	for rows.Next() {
		var id, clientIP, proxyAuthorization *string
		var rawHTTPRequest, rawHTTPResponse *[]byte
		var method, proxyURL, proxyError *string
		var proxyTime, upstreamResponseTime, processingTime *int64
		err := rows.Scan(&id, &clientIP, &proxyAuthorization, &rawHTTPRequest, &rawHTTPResponse, &method, &proxyURL, &proxyError, &proxyTime, &upstreamResponseTime, &processingTime)
		if err != nil {
			return []map[string]interface{}{}, err
		}
		requests = append(requests, map[string]any{
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
		})
	}
	return requests, nil
}

func deleteProxyRequest(id string) (err error) {
	_, err = env.ActiveDB.Exec("DELETE FROM ProxyRequest WHERE id=$1;", id)
	return
}

func deleteAllProxyRequests() (err error) {
	_, err = env.ActiveDB.Exec("DELETE FROM ProxyRequest;")
	return
}
