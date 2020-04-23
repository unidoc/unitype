package cmd

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

const appName = "truecli"
const appVersion = "0.0.1"

const rootCmdDesc = `TrueCLI`

var rootCmd = &cobra.Command{
	Use:  appName,
	Long: appName + " - " + rootCmdDesc,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		ll, _ := cmd.Flags().GetString("loglevel")
		switch ll {
		case "debug":
			logrus.SetFormatter(&logrus.TextFormatter{
				ForceColors:            true,
				FullTimestamp:          true,
				DisableLevelTruncation: true,
			})
			logrus.SetLevel(logrus.DebugLevel)
			logrus.SetOutput(os.Stdout)
		}
	},
}

func fatalf(format string, a ...interface{}) {
	fmt.Printf(format, a...)
	os.Exit(1)
}

func printUsageErr(cmd *cobra.Command, format string, a ...interface{}) {
	fmt.Printf("Error: "+format+"\n", a...)
	cmd.Help()
	os.Exit(1)
}

// Execute represents the entry point of the application.
// The method parses the command line arguments and executes the appropriate
// action.
func Execute() {
	rootCmd.PersistentFlags().String("loglevel", "", "Log level 'debug' and 'trace' give debug information")
	if err := rootCmd.Execute(); err != nil {
		fatalf("Error: %v\n", err)
	}
}
