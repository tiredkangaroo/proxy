package main

import (
	"database/sql"
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
	ActiveDB *sql.DB
	Client   *redis.Client
}

var env = new(Environment)

type CustomHandler struct{}

func (_ CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request, err := parseRequest(r)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, fmt.Sprintf("Malformed request: %s.", err.Error()), http.StatusBadRequest)
		return
	}

	err = allowRequest(request)
	if err != nil {
		request.Error = fmt.Errorf("request blocked by proxy: %s", err.Error())
	}
	if r.Method == "CONNECT" {
		log(request, connectHTTPS(w, request))
	} else {
		log(request, connectHTTP(w, request))
	}
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
