/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"encoding/binary"
	"os"
	"reflect"
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

// cmap4Bytes encodes a format 4 subtable body (after the format field) from
// parallel segment slices, ending with the required 0xFFFF segment.
func cmap4Bytes(start, end, delta, rangeOffset, glyphIDs []uint16) []byte {
	segCount := len(start) + 1
	var buf bytes.Buffer
	w := func(v interface{}) { _ = binary.Write(&buf, binary.BigEndian, v) }
	w(uint16(2*8 + 2*4*segCount + 2*len(glyphIDs))) // length
	w([]uint16{0, uint16(2 * segCount), 0, 0, 0})   // language, segCountX2, searchRange, entrySelector, rangeShift
	w(append(append([]uint16(nil), end...), 0xFFFF))
	w(uint16(0)) // reservedPad
	w(append(append([]uint16(nil), start...), 0xFFFF))
	w(append(append([]uint16(nil), delta...), 1))
	w(append(append([]uint16(nil), rangeOffset...), 0))
	w(glyphIDs)
	return buf.Bytes()
}

// TestParseCmapFormat4_SegmentOrder asserts segments listed out of start-code
// order map correctly, each indexing glyphIDArray from its own position.
func TestParseCmapFormat4_SegmentOrder(t *testing.T) {
	// Segment 0 maps 'b' and segment 1 maps 'a', both through glyphIDArray:
	// index = idRangeOffset/2 + (c - start) + i - segCount, so with segCount 3
	// and idRangeOffset 6, segment 0 reads glyphIDArray[0] and segment 1 reads
	// glyphIDArray[1].
	data := cmap4Bytes([]uint16{'b', 'a'}, []uint16{'b', 'a'}, []uint16{0, 0}, []uint16{6, 6}, []uint16{5, 7})
	f := &font{maxp: &maxpTable{numGlyphs: 10}}
	st, err := f.parseCmapSubtableFormat4(newByteReader(bytes.NewReader(data)), 3, 1)
	require.NoError(t, err)
	assert.Equal(t, map[rune]GlyphIndex{'a': 7, 'b': 5}, st.cmap)
}

// TestParseCmapFormat4_OverlappingSegments asserts repeated overlapping
// segments are expanded once rather than once per segment.
func TestParseCmapFormat4_OverlappingSegments(t *testing.T) {
	const n = 1000
	start, end, zeros := make([]uint16, n), make([]uint16, n), make([]uint16, n)
	for i := range end {
		end[i] = 0xFFFE
	}
	f := &font{maxp: &maxpTable{numGlyphs: 0xFFFF}}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	st, err := f.parseCmapSubtableFormat4(newByteReader(bytes.NewReader(cmap4Bytes(start, end, zeros, zeros, nil))), 3, 1)
	runtime.ReadMemStats(&after)
	require.NoError(t, err)
	assert.Less(t, after.TotalAlloc-before.TotalAlloc, uint64(64<<20), "overlapping segments must not each be expanded")

	one, err := f.parseCmapSubtableFormat4(newByteReader(bytes.NewReader(cmap4Bytes(start[:1], end[:1], zeros[:1], zeros[:1], nil))), 3, 1)
	require.NoError(t, err)
	assert.True(t, reflect.DeepEqual(one.cmap, st.cmap), "must map the same as a single segment")
}

// TestParseCmapFormat4_PartialOverlap asserts a segment partly overlapping an
// earlier one maps only the codes it adds, still indexing glyphIDArray from
// its own start code.
func TestParseCmapFormat4_PartialOverlap(t *testing.T) {
	// Segment 0 maps a-c to glyphs 1-3 by delta. Segment 1 covers b-d through
	// glyphIDArray: index = idRangeOffset/2 + (c - 'b') + 1 - 3, so 'd' reads
	// glyphIDArray[1] and 'b' (index -1) would be out of bounds if evaluated.
	data := cmap4Bytes([]uint16{'a', 'b'}, []uint16{'c', 'd'}, []uint16{0xFFA0, 0}, []uint16{0, 2}, []uint16{5, 9}) // idDelta 0xFFA0 is 1-'a' mod 65536
	f := &font{maxp: &maxpTable{numGlyphs: 10}}
	st, err := f.parseCmapSubtableFormat4(newByteReader(bytes.NewReader(data)), 3, 1)
	require.NoError(t, err)
	assert.Equal(t, map[rune]GlyphIndex{'a': 1, 'b': 2, 'c': 3, 'd': 9}, st.cmap)
}

// cmapTableBytes encodes a cmap table whose encoding records, given as
// {platformID, encodingID, subtable index}, point at the format 12 subtables.
func cmapTableBytes(records [][3]uint16, subtables ...[]byte) []byte {
	var buf bytes.Buffer
	w := func(v interface{}) { _ = binary.Write(&buf, binary.BigEndian, v) }
	w([]uint16{0, uint16(len(records))})
	offsets := make([]uint32, len(subtables))
	next := uint32(4 + 8*len(records))
	for i, st := range subtables {
		offsets[i] = next
		next += 2 + uint32(len(st))
	}
	for _, rec := range records {
		w([]uint16{rec[0], rec[1]})
		w(offsets[rec[2]])
	}
	for _, st := range subtables {
		w(uint16(12))
		w(st)
	}
	return buf.Bytes()
}

// parseCmapBytes parses data as a font's cmap table.
func parseCmapBytes(t *testing.T, data []byte, numGlyphs uint16) (*cmapTable, uint64) {
	t.Helper()
	f := &font{
		maxp: &maxpTable{numGlyphs: numGlyphs},
		trec: &tableRecords{trMap: map[string]*tableRecord{"cmap": {offset: 0, length: uint32(len(data))}}},
	}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	cmap, err := f.parseCmap(newByteReader(bytes.NewReader(data)))
	runtime.ReadMemStats(&after)
	require.NoError(t, err)
	return cmap, after.TotalAlloc - before.TotalAlloc
}

// wideCmap12 is a format 12 subtable body mapping the whole Unicode range.
func wideCmap12() []byte {
	var groups [][3]uint32
	for start := uint32(0); start <= maxUnicodeCodePoint; start += 0x10000 {
		groups = append(groups, [3]uint32{start, start + 0xFFFE, 1})
	}
	return cmap12Bytes(groups...)
}

// TestParseCmap_SharedSubtable asserts encoding records pointing at the same
// subtable with the same rune decoding parse it once, and all stay reachable.
func TestParseCmap_SharedSubtable(t *testing.T) {
	var records [][3]uint16
	for enc := uint16(0); enc < 20; enc++ {
		records = append(records, [3]uint16{0, enc, 0}) // platform 0: all decode as UCS-2
	}
	cmap, alloc := parseCmapBytes(t, cmapTableBytes(records, wideCmap12()), 0xFFFF)
	assert.Len(t, cmap.subtableKeys, 20)
	assert.Equal(t, 19, cmap.subtables["12,0,19"].encodingID)
	assert.Less(t, alloc, uint64(256<<20), "a shared subtable must be parsed once")
}

// TestParseCmap_TotalMappingsCapped asserts parsing stops once distinct
// subtables have mapped more than maxCmapMappings codes.
func TestParseCmap_TotalMappingsCapped(t *testing.T) {
	var records [][3]uint16
	var subtables [][]byte
	for i := uint16(0); i < 5; i++ {
		records = append(records, [3]uint16{3, 10, i})
		subtables = append(subtables, wideCmap12())
	}
	cmap, _ := parseCmapBytes(t, cmapTableBytes(records, subtables...), 0xFFFF)
	assert.Len(t, cmap.subtableKeys, 2, "the first subtable fits the cap; the second passes it; the rest are skipped")
}
