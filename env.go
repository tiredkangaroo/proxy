package main

import (
	"log/slog"
	"os"

	"github.com/tiredkangaroo/loadenv"
)

// LoadEnvironment represents an environment to be loaded from a
// .env file.
type LoadEnvironment struct {
	CACERT string
	CAKEY  string
	CERT   string
	KEY    string
	DEBUG  bool
}

// Environment represents an environment with important reusable
// slices and pointers.
type Environment struct {
	LoadEnvironment
	Logger             *slog.Logger
	CertificateService *CertificateService
	ResponseHandler    ResponseHandler
}

// env provides environment in a way that can be accessed throughout
// the entire codebase.
var env = Environment{
	CertificateService: NewCertificateService(),
}

// loadenv reads the .env file and loads the environment along
// with setting the crucial variables not immediately provided
// in the configuration file.
func load() {
	slog.SetLogLoggerLevel(slog.LevelError)
	env.Logger = slog.Default()

	// load and initialize
	loadedenv := new(LoadEnvironment)
	err := loadenv.Unmarshal(loadedenv)
	if err != nil {
		env.Logger.Error(err.Error())
		os.Exit(1)
	}

	if env.DEBUG == true {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	env.LoadEnvironment = *loadedenv
}
