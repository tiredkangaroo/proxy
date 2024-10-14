package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/tiredkangaroo/loadenv"
)

// LoadEnvironment represents an environment to be loaded from a
// .env file.
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

// Environment represents an environment with important reusable
// slices and pointers.
type Environment struct {
	LoadEnvironment
	ActiveDB     *sql.DB
	Client       *redis.Client
	LogInfo      *LogInfo
	Logger       *slog.Logger
	BlockedSites []*regexp.Regexp
}

// env provides Environment in a way to be accessed throughout
// the entire codebase.
var env = new(Environment)

// CustomHandler provides an http.Handler in which to accept ALL request
// methods.
type CustomHandler struct{}

// ServeHTTP serves the HTTP server for the proxy.
func (_ CustomHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request, err := newProxyHTTPRequest(r)
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

	// load and initialize
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

	err = fetchBlockedSites()
	if err != nil {
		env.Logger.Error(err.Error())
		return
	}

	// start api and proxy servers
	go startAPI()
	handler := new(CustomHandler)
	http.ListenAndServe(":8000", handler)
}
