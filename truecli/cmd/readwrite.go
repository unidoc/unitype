package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gunnsth/unitype"
)

const readwriteCmdDesc = `Reads and write font file back out.

Loads the font file and writes back out. Great for testing the capability
for loading a font file and serializing back.

The input file is loaded from the output argument and the output is
written to "readwrite.ttf".
`

var readwriteCmdExamples = []string{
	fmt.Sprintf("%s readwrite font.ttf", appName),
}

// readwriteCmd represents the font readwrite command.
var readwriteCmd = &cobra.Command{
	Use:     "readwrite <file.ttf>",
	Short:   "Read and write font file",
	Long:    readwriteCmdDesc,
	Example: strings.Join(readwriteCmdExamples, "\n"),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("must provide an input font file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		tfnt, err := unitype.ParseFile(args[0])
		if err != nil {
			log.Fatalf("Error: %+v", err)
		}

		fmt.Printf("tfnt----\n")
		fmt.Printf("%s\n", tfnt.String())

		var buf bytes.Buffer
		err = tfnt.Write(&buf)
		if err != nil {
			fmt.Printf("Error writing: %+v\n", err)
			return
		}

		err = unitype.ValidateBytes(buf.Bytes())
		if err != nil {
			fmt.Printf("Invalid font: %+v\n", err)
			panic(err)
		} else {
			fmt.Printf("Font is valid\n")
		}

		err = tfnt.WriteFile("readwrite.ttf")
		if err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(readwriteCmd)
}
