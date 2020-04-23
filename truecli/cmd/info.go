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
	fmt.Sprintf("%s info --trec font.ttf", appName),
}

// infoCmd represents the font info command.
var infoCmd = &cobra.Command{
	Use:     "info <file.ttf>",
	Short:   "Get font file info",
	Long:    infoCmdDesc,
	Example: strings.Join(infoCmdExamples, "\n"),
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

		showTrec, _ := cmd.Flags().GetBool("trec")
		showHead, _ := cmd.Flags().GetBool("head")
		showOS2, _ := cmd.Flags().GetBool("os2")
		showHhea, _ := cmd.Flags().GetBool("hhea")
		showHmtx, _ := cmd.Flags().GetBool("hmtx")
		showCmap, _ := cmd.Flags().GetBool("cmap")
		showLoca, _ := cmd.Flags().GetBool("loca")
		showGlyf, _ := cmd.Flags().GetBool("glyf")
		showPost, _ := cmd.Flags().GetBool("post")
		showName, _ := cmd.Flags().GetBool("name")
		showCmappings, _ := cmd.Flags().GetBool("cmappings")

		if showTrec {
			fmt.Print(fnt.TableInfo("trec"))
		}
		if showHead {
			fmt.Print(fnt.TableInfo("head"))
		}
		if showOS2 {
			fmt.Print(fnt.TableInfo("os2"))
		}
		if showHhea {
			fmt.Print(fnt.TableInfo("hhea"))
		}
		if showHmtx {
			fmt.Print(fnt.TableInfo("hmtx"))
		}
		if showCmap {
			fmt.Print(fnt.TableInfo("cmap"))
		}
		if showLoca {
			fmt.Print(fnt.TableInfo("loca"))
		}
		if showGlyf {
			fmt.Print(fnt.TableInfo("glyf"))
		}
		if showPost {
			fmt.Print(fnt.TableInfo("post"))
		}
		if showName {
			fmt.Print(fnt.TableInfo("name"))
		}

		if showCmappings {
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
					//if cnt > 100 {
					//	break
					//	}
					fmt.Printf("%d/%s: %d - %c\n", i, mapNames[i], gid, gidMap[gid])
					cnt++
				}
			}
		}
	},
}

func init() {
	infoCmd.Flags().Bool("trec", false, "Show info for trec table")
	infoCmd.Flags().Bool("head", false, "Show info for head table")
	infoCmd.Flags().Bool("os2", false, "Show info for os2 table")
	infoCmd.Flags().Bool("hhea", false, "Show info for hhea table")
	infoCmd.Flags().Bool("hmtx", false, "Show info for hmtx table")
	infoCmd.Flags().Bool("cmap", false, "Show info for cmap table")
	infoCmd.Flags().Bool("loca", false, "Show info for loca table")
	infoCmd.Flags().Bool("glyf", false, "Show info for glyf table")
	infoCmd.Flags().Bool("post", false, "Show info for post table")
	infoCmd.Flags().Bool("name", false, "Show info for name table")
	infoCmd.Flags().Bool("cmappings", false, "List cmap mapping entries")
	rootCmd.AddCommand(infoCmd)
}
