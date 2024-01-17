package main

import (
	"flag"
	"os"

	"github.com/gadfly16/geronimo/cli"
	"github.com/gadfly16/geronimo/server"
	log "github.com/sirupsen/logrus"
)

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
	s            server.Settings
	logLevelName string
)

func init() {
	flag.StringVar(&logLevelName,
		"l", "debug", "Sets log level. (panic, fatal, error, warn, info, debug, trace)")
	flag.StringVar(&s.HTTPAddr,
		"A", "127.0.0.1:8088", "Listening address.")
	flag.StringVar(&s.WorkDir,
		"D", os.Getenv("HOME")+"/.config/Geronimo", "Working directory.")
}

func main() {
	flag.Parse()

	s.LogLevel = debugLevels[logLevelName]
	log.SetLevel(s.LogLevel)

	s.WSAddr = "ws://" + s.HTTPAddr + "/socket"

	if len(os.Args) < 1 {
		log.Fatal("No command given.")
	}

	command, exists := cli.CommandHandlers[flag.Arg(0)]
	if !exists {
		log.Fatalf("%s is not a valid command.", flag.Arg(0))
	}

	// geronimo.Setup()
	err := command(s)
	if err != nil {
		log.Errorln(err)
	}
}
