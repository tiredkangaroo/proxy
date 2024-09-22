package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/tiredkangaroo/loadenv"
)

type LoadEnvironment struct {
	CACERT        string
	CAKEY         string
	CERT          string
	KEY           string
	REDISNETWORK  string
	REDISADDR     string
	REDISUSERNAME string
	REDISPASSWORD string
	REDISDB       string
}

type Environment struct {
	LoadEnvironment
	Client *redis.Client
}

var env = new(Environment)

type CustomHandler struct{}

func (_ CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request, err := parseRequest(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Malformed request."), http.StatusBadRequest)
		return
	}

	err = allowRequest(request)
	if err != nil {
		http.Error(w, fmt.Sprintf("Request blocked by proxy: %s.", err.Error()), http.StatusForbidden)
		return
	}

	if r.Method == "CONNECT" {
		if err := connectHTTPS(w, request); err != nil {
			logerror(request, err)
		}
	} else {
		r.Header.Del("Proxy-Authorization")
		r.Header.Del("Proxy-Connection")
		r.RequestURI = ""

		request.Method = r.Method
		request.URL = r.URL

		resp, err := http.DefaultClient.Do(r)
		w.Header().Add("X-ProxyRequest-ID", request.ID)
		if err != nil {
			logerror(request, err)
			http.Error(w, "an error occured with the upstream client", 502)
			return
		}
		resp.Write(w)
	}
	log(request)
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	loadedenv := new(LoadEnvironment)
	err := loadenv.Unmarshal(loadedenv)
	if err != nil {
		slog.Error(err.Error())
		return
	}
	env.LoadEnvironment = *loadedenv

	db, _ := strconv.Atoi(env.REDISDB)
	env.Client = redis.NewClient(&redis.Options{
		Network:  env.REDISNETWORK,
		Addr:     env.REDISADDR,
		Username: env.REDISUSERNAME,
		Password: env.REDISPASSWORD,
		DB:       db,
	})

	handler := new(CustomHandler)
	http.ListenAndServe(":8000", handler)
}
