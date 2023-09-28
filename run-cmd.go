package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

func init() {
	commands["run"] = runCommand
}

func runCommand() {
	log.Debug("Running 'run' command.")

	var (
		simulation bool
	)

	runFlags := flag.NewFlagSet("run", flag.ExitOnError)
	runFlags.BoolVar(&simulation, "y", false, "Simulation mode.")
	runFlags.Parse(flag.Args()[1:])

	if simulation {
		log.Info("Running in simulation mode.")
	}

	db := openDB()
	defer db.Close()

	activeAccs := getActiveAccounts(db)

	// Start accounts
	for _, acc := range activeAccs {
		acc.decryptKeys()
		go acc.run()
	}

	// Wait for stop signals and quit gently
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	for range signals {
		log.Warn("Stopping...")
		return
	}
}
