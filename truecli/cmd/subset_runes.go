/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gunnsth/unitype"
)

const subsetRunesCmdDesc = `Subset a font file to a set of runes.

Works by removing data for any runes corresponding to GIDs outside the subset.
Maintains GID order and other information.

Advantage is that the GIDs are maintained, less risky, and gives
good size reduction as the glyf table is usually biggest by far.
`

var subsetRunesCmdExamples = []string{
	fmt.Sprintf("%s subset runes font.ttf abcefgh", appName),
}

// subsetRunesCmd represents the font subsetting by runes command.
var subsetRunesCmd = &cobra.Command{
	Use:     "runes <file.ttf> <runes>",
	Short:   "Subset font file to specific rune subset",
	Long:    subsetRunesCmdDesc,
	Example: strings.Join(subsetRunesCmdExamples, "\n"),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("must provide an input font file")
		}
		if len(args) < 2 {
			return errors.New("must provide runes to subset")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		tfnt, err := unitype.ParseFile(args[0])
		if err != nil {
			fatalf("Error: %+v\n", err)
		}

		outpath, _ := cmd.Flags().GetString("outfile")

		fmt.Printf("Original %s: %s\n", args[0], tfnt.String())

		runes := []rune(args[1])

		var buf bytes.Buffer
		err = tfnt.Write(&buf)
		if err != nil {
			fmt.Printf("Error writing: %+v\n", err)
			return
		}
		origSize := buf.Len()

		err = unitype.ValidateBytes(buf.Bytes())
		if err != nil {
			fmt.Printf("Invalid font: %+v\n", err)
			panic(err)
		} else {
			fmt.Printf("Font is valid\n")
		}

		// Try subsetting font.
		subfnt, err := tfnt.SubsetKeepRunes(runes)
		if err != nil {
			panic(err)
		}

		buf.Reset()
		err = subfnt.Write(&buf)
		if err != nil {
			fmt.Printf("Failed writing: %+v\n", err)
			panic(err)
		}
		subsetSize := buf.Len()

		fmt.Printf("Original size: %d\n", origSize)
		fmt.Printf("Subset size: %d (%.2f X)\n", buf.Len(), float64(origSize)/float64(subsetSize))
		err = unitype.ValidateBytes(buf.Bytes())
		if err != nil {
			fmt.Printf("Invalid subfnt: %+v\n", err)
			panic(err)
		} else {
			fmt.Printf("subset font is valid\n")
		}

		err = subfnt.WriteFile(outpath)
		if err != nil {
			fatalf("ERROR: %v\n", err)
		}
		fmt.Printf("Output written: %s\n", outpath)
	},
}

func init() {
	subsetRunesCmd.Flags().StringP("outfile", "o", "subset_runes.ttf", "Output file name")
	subsetCmd.AddCommand(subsetRunesCmd)
}
