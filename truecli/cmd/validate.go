/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/unidoc/unitype"
)

const validateCmdDesc = `Validate font file.`

// validateCmd represents the font validation command.
var validateCmd = &cobra.Command{
	Use:   "validate <file.ttf>",
	Short: "Validate font file",
	Long:  validateCmdDesc,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("must provide an input font file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Validate.
		err := unitype.ValidateFile(args[0])
		if err != nil {
			panic(err)
		}

		// Parse and output info.
		fnt, err := unitype.ParseFile(args[0])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", fnt.String())
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
