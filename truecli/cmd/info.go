/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package cmd

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gunnsth/unitype"
)

const infoCmdDesc = `Information from font file.`

var infoCmdExamples = []string{
	fmt.Sprintf("%s info font.ttf", appName),
}

// infoCmd represents the font info command.
var infoCmd = &cobra.Command{
	Use:     "info <file.ttf>",
	Short:   "Get font file info",
	Long:    infoCmdDesc,
	Example: strings.Join(subsetCmdExamples, "\n"),
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("must provide an input font file")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		fnt, err := unitype.ParseFile(args[0])
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s\n", fnt.String())

		var maps []map[rune]unitype.GlyphIndex
		var mapNames []string
		maps = append(maps, fnt.GetCmap(0, 3))
		mapNames = append(mapNames, "0,3")
		maps = append(maps, fnt.GetCmap(1, 0))
		mapNames = append(mapNames, "1,0")
		maps = append(maps, fnt.GetCmap(3, 1))
		mapNames = append(mapNames, "3,1")

		for i := range maps {
			var gids []unitype.GlyphIndex
			gidMap := map[unitype.GlyphIndex]rune{}
			for rune, gid := range maps[i] {
				gidMap[gid] = rune
				gids = append(gids, gid)
			}
			sort.Slice(gids, func(i, j int) bool {
				return gids[i] < gids[j]
			})
			cnt := 0
			for _, gid := range gids {
				if cnt > 100 {
					break
				}
				fmt.Printf("%d/%s: %d - %c\n", i, mapNames[i], gid, gidMap[gid])
				cnt++
			}
			/*
				for rune, igid := range maps[i] {
					for j := range maps {
						if i == j {
							continue
						}
						jgid, has := maps[j][rune]
						if has && jgid != igid {
							fmt.Printf("Disagreement map %d and %d: %c: %d/%d\n", i, j, rune, igid, jgid)
						}
					}
				}
			*/
		}

		/*
			for rune, GID := range map1 {
				fmt.Printf("0,3: %c - %d\n", rune, GID)
			}
			for rune, GID := range map2 {
				fmt.Printf("1,0: %c - %d\n", rune, GID)
			}
			for rune, GID := range map3 {
				fmt.Printf("3,1: %c - %d\n", rune, GID)
			}
		*/
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
