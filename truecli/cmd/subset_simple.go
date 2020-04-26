/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"

	"github.com/unidoc/unitype"
)

const subsetSimpleCmdDesc = `Subset a font file to a simple subset (256 glyphs).

Outputs a new font file "subset.ttf" that contains only
the first 256 glyphs from the input font file.

TODO: In the future add options to select what glyphs are
picked, like a set of GID ranges or lists of runes.
`

var subsetSimpleCmdExamples = []string{
	fmt.Sprintf("%s subset simple font.ttf", appName),
}

// subsetSimpleCmd represents the font subsetting command.
var subsetSimpleCmd = &cobra.Command{
	Use:     "simple <file.ttf>",
	Short:   "Subset font file to simple subset",
	Long:    subsetSimpleCmdDesc,
	Example: strings.Join(subsetSimpleCmdExamples, "\n"),
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

		// Try subsetting font.
		subfnt, err := tfnt.SubsetSimple(256)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Subset font: %s\n", subfnt.String())

		buf.Reset()
		err = subfnt.Write(&buf)
		if err != nil {
			fmt.Printf("Failed writing: %+v\n", err)
			panic(err)
		}
		fmt.Printf("Subset font length: %d\n", buf.Len())
		err = unitype.ValidateBytes(buf.Bytes())
		if err != nil {
			fmt.Printf("Invalid subfnt: %+v\n", err)
			panic(err)
		} else {
			fmt.Printf("subset font is valid\n")
		}

		err = subfnt.WriteFile("subset.ttf")
		if err != nil {
			fatalf("ERROR: %v\n", err)
		}
	},
}

func init() {
	subsetCmd.AddCommand(subsetSimpleCmd)
}
