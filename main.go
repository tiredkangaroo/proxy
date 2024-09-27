package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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
	Logger   *slog.Logger
}

var env = new(Environment)

type CustomHandler struct{}

func (_ CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request, err := parseRequest(r)
	if err != nil {
		env.Logger.Error("malformed request.", "error", err.Error())
		http.Error(w, fmt.Sprintf("Malformed request: %s.", err.Error()), http.StatusBadRequest)
		return
	}

	if r.Method == "CONNECT" {
		log(request, connectHTTPS(w, request))
	} else {
		log(request, connectHTTP(w, request))
	}
}

func main() {
	env.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	loadedenv := new(LoadEnvironment)
	err := loadenv.Unmarshal(loadedenv)
	if err != nil {
		env.Logger.Error(err.Error())
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
