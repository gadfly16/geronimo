package main

import (
	"flag"
	"os"

	log "github.com/sirupsen/logrus"
)

var commands = make(map[string]func())

var debugLevels = map[string]log.Level{
	"panic": log.PanicLevel,
	"fatal": log.FatalLevel,
	"error": log.ErrorLevel,
	"warn":  log.WarnLevel,
	"info":  log.InfoLevel,
	"debug": log.DebugLevel,
	"trace": log.TraceLevel,
}

var (
	debugFlag    string
	databaseFlag string
)

func init() {
	flag.StringVar(&debugFlag, "d", "debug", "Sets log level. (panic, fatal, error, warn, info, debug, trace)")
	flag.StringVar(&databaseFlag, "D", "./settings.db", "Path to the settings database.")
}

func main() {
	log.Debug("Parsing global flags.")
	flag.Parse()
	log.SetLevel(debugLevels[debugFlag])

	log.Debug("Parsing main command flags.")

	if len(os.Args) < 1 {
		log.Fatal("No command given.")
	}

	command, exists := commands[flag.Arg(0)]
	if !exists {
		log.Fatalf("%s is not a valid command.", os.Args[0])
	}

	command()
}
