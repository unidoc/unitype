/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/unidoc/unitype"
)

const subsetGIDsCmdDesc = `Subset a font file to a set of GIDs.

Works by removing data for any GIDs outside the subset. Does not
change GID order.

Advantage is that the GIDs are maintained, less risky, and gives
good size reduction as the glyf table is usually biggest by far.
`

var subsetGIDsCmdExamples = []string{
	fmt.Sprintf("%s subset gids font.ttf 10 20 30", appName),
}

// subsetGIDsCmd represents the font subsetting by GIDs command.
var subsetGIDsCmd = &cobra.Command{
	Use:     "gids <file.ttf> <gid1> <gid2> ...",
	Short:   "Subset font file to specific GID subset",
	Long:    subsetGIDsCmdDesc,
	Example: strings.Join(subsetGIDsCmdExamples, "\n"),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("must provide an input font file")
		}
		if len(args) < 2 {
			return errors.New("must provide at least one GID")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var gids []unitype.GlyphIndex
		for i := 1; i < len(args); i++ {
			gid, err := strconv.ParseUint(args[i], 10, 32)
			if err != nil {
				fatalf("Invalid gid: %v\n", err)
			}
			gids = append(gids, unitype.GlyphIndex(gid))
		}

		tfnt, err := unitype.ParseFile(args[0])
		if err != nil {
			fatalf("Error: %+v\n", err)
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
		subfnt, err := tfnt.SubsetKeepIndices(gids)
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

		err = subfnt.WriteFile("subset_gids.ttf")
		if err != nil {
			fatalf("ERROR: %v\n", err)
		}
	},
}

func init() {
	subsetCmd.AddCommand(subsetGIDsCmd)
}
