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
	POSTGRESURI   string
	DEBUG         bool
	RAWLOGINFO    string `required:"false"`
}

type Environment struct {
	LoadEnvironment
	ActiveDB *sql.DB
	Client   *redis.Client
	LogInfo  *LogInfo
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
		log(request, connectHTTP(w, r, request))
	}
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	env.Logger = slog.Default()

	loadedenv := new(LoadEnvironment)
	err := loadenv.Unmarshal(loadedenv)
	if err != nil {
		env.Logger.Error(err.Error())
		return
	}
	env.LoadEnvironment = *loadedenv
	env.LogInfo = parseLogInfo(env.RAWLOGINFO)

	db, _ := strconv.Atoi(env.REDISDB)
	env.Client = redis.NewClient(&redis.Options{
		Network:  env.REDISNETWORK,
		Addr:     env.REDISADDR,
		Username: env.REDISUSERNAME,
		Password: env.REDISPASSWORD,
		DB:       db,
	})

	err = initalizeDB()
	if err != nil {
		env.Logger.Error(err.Error())
		return
	}

	go startAPI()
	handler := new(CustomHandler)
	http.ListenAndServe(":8000", handler)
}
