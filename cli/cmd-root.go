package cli

import (
	"fmt"
	"log/slog"
	"os"
	"runtime/pprof"

	"github.com/gadfly16/geronimo/node"
	"github.com/spf13/cobra"
)

var (
	rp           node.RootParms
	logLevelName string
	sdb          string
	userEmail    string
	userPassword string
	prof_cpu     bool
	prof_mem     bool

	runtimeErr bool

	cpuProf *os.File
	// memProf *os.File
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&logLevelName, "log-level", "L", "debug", "logging level")
	rootCmd.PersistentFlags().StringVarP(&sdb, "state-db", "S", os.Getenv("HOME")+"/.config/Nerd/state.db", "state database path")
	rootCmd.PersistentFlags().StringVarP(&rp.HTTPAddr, "http-address", "A", "localhost:8088", "server HTTP address")
	rootCmd.PersistentFlags().StringVarP(&userEmail, "user-email", "U", "", "login email address")
	rootCmd.PersistentFlags().StringVarP(&userPassword, "user-password", "P", "", "login password")
	rootCmd.PersistentFlags().BoolVarP(&prof_cpu, "prof-cpu", "", false, "CPU pprof file")
	rootCmd.PersistentFlags().BoolVarP(&prof_mem, "prof-mem", "", false, "memory pprof file")
}

var rootCmd = &cobra.Command{
	Use:   "geronimo",
	Short: "Geronimo is an crypto investment platform",
	Long:  `Geronimo is a web application to track, manage and automate crypto investments.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		slog.Info("Inside root command's persistent pre run.")
		l := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: node.LogLevel})
		slog.SetDefault(slog.New(l))
		if prof_cpu {
			var err error
			cpuProf, err = os.Create("cpu.prof")
			if err != nil {
				slog.Error("Could not create CPU profile. Exiting!", "error", err)
				return
			}
			if err := pprof.StartCPUProfile(cpuProf); err != nil {
				slog.Error("Could not start CPU profile. Exiting!", "error", err)
			}
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if prof_cpu {
			pprof.StopCPUProfile()
			cpuProf.Close()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Running root command.")
	},
}

func Execute() {
	cobra.EnableTraverseRunHooks = true
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if runtimeErr {
		os.Exit(1)
	}
}
