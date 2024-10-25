package main

import (
	"database/sql"
	"log/slog"
	"net/url"
	"os"
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
	ActiveDB           *sql.DB
	Client             *redis.Client
	LogInfo            *LogInfo
	Logger             *slog.Logger
	BlockedSites       []*regexp.Regexp
	BlockerURLs        []*url.URL
	CertificateService *CertificateService
}

// env provides Environment in a way to be accessed throughout
// the entire codebase.
var env = Environment{
	CertificateService: NewCertificateService(),
}

// load reads the .env file and loads the environment along
// with setting the crucial variables not immediately provided
// in the configuration file.
func load() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	env.Logger = slog.Default()

	// load and initialize
	loadedenv := new(LoadEnvironment)
	err := loadenv.Unmarshal(loadedenv)
	if err != nil {
		env.Logger.Error(err.Error())
		os.Exit(1)
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
		os.Exit(1)
	}

	err = fetchBlockedSites()
	if err != nil {
		env.Logger.Error(err.Error())
		os.Exit(1)
	}
}
