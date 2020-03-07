package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const appName = "truecli"
const appVersion = "0.0.1"

const rootCmdDesc = `TrueCLI`

var rootCmd = &cobra.Command{
	Use:  appName,
	Long: appName + " - " + rootCmdDesc,
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
	if err := rootCmd.Execute(); err != nil {
		fatalf("%s\n", err)
	}
}
