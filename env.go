package main

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/tiredkangaroo/loadenv"
)

// LoadEnvironment represents an environment to be loaded from a
// .env file.
type LoadEnvironment struct {
	CACERT     string
	CAKEY      string
	CERT       string
	KEY        string
	DEBUG      bool
	RAWLOGINFO string `required:"false"`
}

// Environment represents an environment with important reusable
// slices and pointers.
type Environment struct {
	LoadEnvironment
	ActiveDB           *sql.DB
	Logger             *slog.Logger
	CertificateService *CertificateService
}

// env provides Environment in a way to be accessed throughout
// the entire codebase.
var env = Environment{
	CertificateService: NewCertificateService(),
}

// loadenv reads the .env file and loads the environment along
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
	// env.LogInfo = parseLogInfo(env.RAWLOGINFO)
}
