/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmapTableReadWrite(t *testing.T) {
	type expectedCmap struct {
		format         int
		platformID     int
		encodingID     int
		numRuneEntries int
		numMapEntries  int                 // number of entries in the map
		checks         map[rune]GlyphIndex // a few spot checks.
	}
	testcases := []struct {
		fontPath      string
		expectedCmaps map[string]expectedCmap
	}{
		{
			"./testdata/FreeSans.ttf",
			map[string]expectedCmap{
				"4,0,3": {
					4,
					0,
					3,
					3726,
					10378 - 7574 + 1,
					map[rune]GlyphIndex{
						'a':          70,
						' ':          5,
						'!':          6,
						'@':          37,
						'Æ':          138,
						'π':          711,
						rune(0x1FFE): 2193,
					},
				},
				"6,1,0": {
					6,
					1,
					0,
					256,
					10636 - 10381 + 1,
					map[rune]GlyphIndex{
						'a': 70,
						' ': 5,
						'!': 6,
						'@': 37,
						'Æ': 138,
						'π': 711,
					},
				},
				"4,3,1": {
					4,
					3,
					1,
					3726,
					13443 - 10639 + 1,
					map[rune]GlyphIndex{
						'a':          70,
						' ':          5,
						'!':          6,
						'@':          37,
						'Æ':          138,
						'π':          711,
						rune(0x1FFE): 2193,
					},
				},
			},
		},
		{
			"./testdata/wts11.ttf",
			map[string]expectedCmap{
				"4,0,3": {
					4,
					0,
					3,
					14148,
					42387 - 28418 + 1,
					map[rune]GlyphIndex{},
				},
				"0,1,0": {
					0,
					1,
					0,
					256,
					105, //42645 - 42390 + 1, not counting notdefs
					map[rune]GlyphIndex{},
				},
				"4,3,1": {
					4,
					3,
					1,
					14148,
					56617 - 42648 + 1,
					map[rune]GlyphIndex{},
				},
			},
		},
		{
			"./testdata/roboto/Roboto-BoldItalic.ttf",
			map[string]expectedCmap{
				"4,0,3": {
					4,
					0,
					3,
					1294,
					896,
					map[rune]GlyphIndex{},
				},
				"4,3,1": {
					4,
					3,
					1,
					1294,
					896,
					map[rune]GlyphIndex{},
				},
				"12,3,10": {
					12,
					3,
					10,
					1294,
					896,
					map[rune]GlyphIndex{},
				},
			},
		},
	}

	for _, tcase := range testcases {
		t.Run(tcase.fontPath, func(t *testing.T) {
			t.Logf("%s", tcase.fontPath)
			f, err := os.Open(tcase.fontPath)
			assert.Equal(t, nil, err)
			defer f.Close()

			br := newByteReader(f)
			fnt, err := parseFont(br)
			assert.Equal(t, nil, err)
			require.NoError(t, err)

			require.NotNil(t, fnt)
			require.NotNil(t, fnt.cmap)
			require.NotNil(t, fnt.cmap.subtables)

			require.Equal(t, len(tcase.expectedCmaps), len(fnt.cmap.subtables))

			for _, key := range fnt.cmap.subtableKeys {
				subtable := fnt.cmap.subtables[key]
				t.Logf("subtable %d %d/%d '%s'", subtable.format, subtable.platformID, subtable.encodingID, key)
				exp := tcase.expectedCmaps[key]
				require.Equal(t, exp.format, subtable.format)
				require.Equal(t, exp.platformID, subtable.platformID)
				require.Equal(t, exp.encodingID, subtable.encodingID)
				require.Equal(t, exp.numRuneEntries, len(subtable.runes))
				require.Equal(t, exp.numMapEntries, len(subtable.cmap))

				t.Logf("- cmap len: %d", len(subtable.cmap))
				// spot checks.
				for r, gid := range exp.checks {
					t.Logf("%c 0x%X", r, r)
					require.Equal(t, gid, subtable.cmap[r])
				}
			}

			// Write, read back and repeat checks.
			var buf bytes.Buffer
			bw := newByteWriter(&buf)
			err = fnt.write(bw)
			require.NoError(t, err)
			err = bw.flush()
			require.NoError(t, err)
			br = newByteReader(bytes.NewReader(buf.Bytes()))
			fnt, err = parseFont(br)
			assert.Equal(t, nil, err)
			require.NoError(t, err)
			require.NotNil(t, fnt)
			require.NotNil(t, fnt.cmap)
			require.NotNil(t, fnt.cmap.subtables)
			require.Equal(t, len(tcase.expectedCmaps), len(fnt.cmap.subtables))
			for _, key := range fnt.cmap.subtableKeys {
				subtable := fnt.cmap.subtables[key]
				t.Logf("2 subtable %d %d/%d '%s'", subtable.format, subtable.platformID, subtable.encodingID, key)
				exp := tcase.expectedCmaps[key]
				require.Equal(t, exp.format, subtable.format)
				require.Equal(t, exp.platformID, subtable.platformID)
				require.Equal(t, exp.encodingID, subtable.encodingID)
				require.Equal(t, exp.numRuneEntries, len(subtable.runes))
				require.Equal(t, exp.numMapEntries, len(subtable.cmap))

				t.Logf("2 - cmap len: %d", len(subtable.cmap))

				// spot checks.
				for r, gid := range exp.checks {
					t.Logf("%c 0x%X", r, r)
					require.Equal(t, gid, subtable.cmap[r])
				}
			}
		})
	}
}
