/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"encoding/binary"
	"os"
	"runtime"
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

// cmap12Bytes encodes a format 12 subtable body (after the format field) from
// {startCharCode, endCharCode, startGlyphID} groups.
func cmap12Bytes(groups ...[3]uint32) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, uint16(0))                 // reserved
	_ = binary.Write(&buf, binary.BigEndian, uint32(16+12*len(groups))) // length
	_ = binary.Write(&buf, binary.BigEndian, uint32(0))                 // language
	_ = binary.Write(&buf, binary.BigEndian, uint32(len(groups)))       // numGroups
	for _, g := range groups {
		_ = binary.Write(&buf, binary.BigEndian, g)
	}
	return buf.Bytes()
}

// TestParseCmapFormat12_Groups asserts groups map as declared in any order;
// a group overlapping an earlier one is trimmed to the codes it adds, one
// covered by earlier groups is skipped, as are codes beyond U+10FFFF.
func TestParseCmapFormat12_Groups(t *testing.T) {
	f := &font{maxp: &maxpTable{numGlyphs: 10}}
	data := cmap12Bytes(
		[3]uint32{0x61, 0x62, 8},         // a-b -> 8-9, listed out of order
		[3]uint32{0x41, 0x43, 1},         // A-C -> 1-3
		[3]uint32{0x42, 0x44, 5},         // overlaps A-C: trimmed to D -> 7
		[3]uint32{0x41, 0x42, 8},         // covered by A-C: skipped
		[3]uint32{0x110000, 0x110005, 7}, // beyond U+10FFFF: skipped
	)
	st, err := f.parseCmapSubtableFormat12(newByteReader(bytes.NewReader(data)), 3, 10)
	require.NoError(t, err)
	assert.Equal(t, map[rune]GlyphIndex{'A': 1, 'B': 2, 'C': 3, 'D': 7, 'a': 8, 'b': 9}, st.cmap)
}

// TestParseCmapFormat12_ManyGroups asserts a table with many wide groups
// expands at most the Unicode range, however many groups it declares.
func TestParseCmapFormat12_ManyGroups(t *testing.T) {
	var groups [][3]uint32
	for i := uint32(0); i < 400; i++ {
		groups = append(groups, [3]uint32{i * 0x10000, i*0x10000 + 0xFFFF, 0})
	}
	f := &font{maxp: &maxpTable{numGlyphs: 0xFFFF}}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	_, err := f.parseCmapSubtableFormat12(newByteReader(bytes.NewReader(cmap12Bytes(groups...))), 3, 10)
	runtime.ReadMemStats(&after)
	require.NoError(t, err)
	assert.Less(t, after.TotalAlloc-before.TotalAlloc, uint64(256<<20), "expansion must stay bounded")
}

// TestParseCmapFormat12_StartGlyphIDRange asserts a startGlyphID beyond
// numGlyphs is rejected rather than truncated to a valid glyph index.
func TestParseCmapFormat12_StartGlyphIDRange(t *testing.T) {
	f := &font{maxp: &maxpTable{numGlyphs: 10}}
	_, err := f.parseCmapSubtableFormat12(newByteReader(bytes.NewReader(cmap12Bytes([3]uint32{0x41, 0x41, 0x10001}))), 3, 10)
	assert.ErrorIs(t, err, errRangeCheck)
}

// TestParseCmapFormat12_TrimPastGlyphs asserts a group trimmed so that its
// first glyph falls past numGlyphs maps nothing, rather than erroring.
func TestParseCmapFormat12_TrimPastGlyphs(t *testing.T) {
	f := &font{maxp: &maxpTable{numGlyphs: 10}}
	data := cmap12Bytes(
		[3]uint32{0x41, 0x50, 1}, // A-P -> 1-9 (stops at numGlyphs)
		[3]uint32{0x41, 0x60, 9}, // trimmed to Q-`, first glyph 9+16 >= numGlyphs
	)
	st, err := f.parseCmapSubtableFormat12(newByteReader(bytes.NewReader(data)), 3, 10)
	require.NoError(t, err)
	_, hasQ := st.cmap['Q']
	assert.False(t, hasQ)
}
