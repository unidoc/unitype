package cmd

import "github.com/spf13/cobra"

const subsetCmdDesc = `Subset font.`

// subset represents font subsetting commands root.
var subsetCmd = &cobra.Command{
	Use:   "subset [FLAG]... COMMAND",
	Short: "Subset font",
	Long:  subsetCmdDesc,
}

func init() {
	rootCmd.AddCommand(subsetCmd)
}
