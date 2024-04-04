package main

import (
	"os"

	"github.com/gadfly16/geronimo/server"
	"github.com/jessevdk/go-flags"
)

// var debugLevels = map[string]log.Level{
// 	"panic": log.PanicLevel,
// 	"fatal": log.FatalLevel,
// 	"error": log.ErrorLevel,
// 	"warn":  log.WarnLevel,
// 	"info":  log.InfoLevel,
// 	"debug": log.DebugLevel,
// 	"trace": log.TraceLevel,
// }

var s server.Settings

var parser = flags.NewParser(&s, flags.Default)

// func init() {
// 	flag.StringVar(&logLevelName,
// 		"l", "debug", "Sets log level. (panic, fatal, error, warn, info, debug, trace)")
// 	flag.StringVar(&s.HTTPAddr,
// 		"A", "127.0.0.1:8088", "Listening address.")
// 	flag.StringVar(&s.WorkDir,
// 		"D", os.Getenv("HOME")+"/.config/Geronimo", "Working directory.")
// 	flag.StringVar(&s.UserEmail,
// 		"u", "", "User email.")
// 	flag.StringVar(&s.UserPassword,
// 		"p", "", "User password.")
// }

func main() {
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	// s.WSAddr = "ws://" + s.HTTPAddr + "/socket"

	// // Init file paths
	// s.DBPath = s.WorkDir + "/" + server.NameStateDB
	// s.DBKeyPath = s.WorkDir + "/" + server.NameDBKey
	// s.JWTKeyPath = s.WorkDir + "/" + server.NameJWTKey
	// s.CLICookiePath = s.WorkDir + "/" + server.NameCLICookie

	// if len(os.Args) < 1 {
	// 	log.Fatal("No command given.")
	// }

	// command, exists := cli.CommandHandlers[flag.Arg(0)]
	// if !exists {
	// 	log.Fatalf("%s is not a valid command.", flag.Arg(0))
	// }

	// // geronimo.Setup()
	// err := command(s)
	// if err != nil {
	// 	log.Errorln(err)
	// }
}
