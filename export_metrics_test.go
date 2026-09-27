/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRobotoMetrics covers the metric accessors against Roboto-Regular.ttf.
func TestRobotoMetrics(t *testing.T) {
	f, err := ParseFile("./testdata/roboto/Roboto-Regular.ttf")
	require.NoError(t, err)

	t.Run("OS2Metrics", func(t *testing.T) {
		m := f.OS2Metrics()
		assert.True(t, m.Present, "Roboto should have OS/2 table")
		assert.Greater(t, m.TypoAscender, int16(0))
		assert.Less(t, m.TypoDescender, int16(0))
		assert.GreaterOrEqual(t, m.TypoLineGap, int16(0))
		assert.Greater(t, m.WinAscent, uint16(0))
		assert.Greater(t, m.WinDescent, uint16(0))
		// Exact values from Roboto-Regular.ttf.
		assert.Equal(t, int16(1082), m.XHeight, "Roboto-Regular sxHeight")
		assert.Equal(t, int16(1456), m.CapHeight, "Roboto-Regular sCapHeight")
		assert.False(t, m.UseTypoMetrics, "Roboto-Regular fsSelection bit 7 (USE_TYPO_METRICS) is unset")
	})

	t.Run("HheaMetrics", func(t *testing.T) {
		m := f.HheaMetrics()
		assert.True(t, m.Present, "Roboto should have hhea table")
		assert.Greater(t, m.Ascender, int16(0))
		assert.Less(t, m.Descender, int16(0))
	})

	t.Run("UnitsPerEm", func(t *testing.T) {
		assert.Equal(t, uint16(2048), f.UnitsPerEm(), "Roboto uses 2048 UPEM")
	})

	t.Run("NumGlyphs", func(t *testing.T) {
		assert.Greater(t, f.NumGlyphs(), 100)
	})

	t.Run("GlyphAdvance", func(t *testing.T) {
		// Look up glyph for 'A' and verify it has a non-zero advance.
		gids := f.LookupRunes([]rune{'A'})
		require.Len(t, gids, 1)
		require.NotEqual(t, GlyphIndex(0), gids[0], "'A' should resolve to a glyph")

		adv, ok := f.GlyphAdvance(gids[0])
		assert.True(t, ok)
		assert.Greater(t, adv, uint16(0), "'A' should have a positive advance")
	})

	t.Run("GlyphAdvance_OutOfRange", func(t *testing.T) {
		// A GID at or beyond the font's real glyph count is out of range: ok
		// must be false, distinguishing it from a real zero-width glyph.
		invalid := GlyphIndex(f.NumGlyphs())
		_, ok := f.GlyphAdvance(invalid)
		assert.False(t, ok, "GID at NumGlyphs() is out of range")

		// GID 0xFFFF is out of range.
		require.Less(t, f.NumGlyphs(), 0xFFFF, "test assumes NumGlyphs() leaves room below the GlyphIndex max")
		_, ok = f.GlyphAdvance(GlyphIndex(0xFFFF))
		assert.False(t, ok, "far out-of-range GID must also report not-ok")
	})
}

// TestOS2Metrics_UseTypoMetricsVersionGate asserts fsSelection bit 7 is
// honored only for OS/2 version >= 4.
func TestOS2Metrics_UseTypoMetricsVersionGate(t *testing.T) {
	// length: os2LenV2to4, a genuine non-truncated v4 table, so
	// hasTypoWinMetrics() is true and doesn't mask the bit-7 gate under test.
	newFont := func(version, fsSelection uint16) *Font {
		return &Font{font: &font{os2: &os2Table{version: version, fsSelection: fsSelection, length: os2LenV2to4}}}
	}

	v3BitSet := newFont(3, 0x0080)
	assert.False(t, v3BitSet.OS2Metrics().UseTypoMetrics,
		"bit 7 is reserved before v4 and must not be read as USE_TYPO_METRICS")

	v4BitSet := newFont(4, 0x0080)
	assert.True(t, v4BitSet.OS2Metrics().UseTypoMetrics, "bit 7 is defined from v4 onward")

	v4BitClear := newFont(4, 0x0000)
	assert.False(t, v4BitClear.OS2Metrics().UseTypoMetrics)
}

// truncatedOS2Bytes returns payloadLen bytes of an OS/2 table claiming
// declaredVersion, followed by sentinel bytes standing in for the next table.
func truncatedOS2Bytes(declaredVersion uint16, payloadLen int) []byte {
	payload := bytes.Repeat([]byte{0xAB}, payloadLen)
	payload[0] = byte(declaredVersion >> 8) // big-endian, per byteReader.readUint16
	payload[1] = byte(declaredVersion)
	return append(payload, bytes.Repeat([]byte{0xFF}, 20)...)
}

// TestParseOS2Table_Truncated asserts, at every OS/2 length boundary, that
// fields beyond the declared length or version stay zero.
func TestParseOS2Table_Truncated(t *testing.T) {
	tests := []struct {
		name            string
		declaredVersion uint16
		payloadLen      int
		wantTypoWin     bool
		wantV2Metrics   bool
		wantV5Metrics   bool
	}{
		{"v0 Apple (68)", 0, os2LenV0Apple, false, false, false},
		{"v0 Microsoft (78)", 0, os2LenV0Microsoft, true, false, false},
		{"v1 (86)", 1, os2LenV1, true, false, false},
		{"v4 truncated at v1 length (86)", 4, os2LenV1, true, false, false},
		{"v1 padded to v2-4 length (96)", 1, os2LenV2to4, true, false, false},
		{"v2-4 (96)", 4, os2LenV2to4, true, true, false},
		{"v2-4 padded to v5 length (100)", 4, os2LenV5, true, true, false},
		{"v5 (100)", 5, os2LenV5, true, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := truncatedOS2Bytes(tt.declaredVersion, tt.payloadLen)
			f := &font{
				trec: &tableRecords{
					trMap: map[string]*tableRecord{
						"OS/2": {offset: 0, length: uint32(tt.payloadLen)},
					},
				},
			}

			table, err := f.parseOS2Table(newByteReader(bytes.NewReader(data)))
			require.NoError(t, err)
			require.NotNil(t, table)

			assert.Equal(t, tt.wantTypoWin, table.hasTypoWinMetrics())
			assert.Equal(t, tt.wantV2Metrics, table.hasV2Metrics())
			assert.Equal(t, tt.wantV5Metrics, table.hasV5Metrics())
			if !tt.wantTypoWin {
				assert.Zero(t, table.sTypoAscender, "must not read past declared length")
				assert.Zero(t, table.usWinDescent, "must not read past declared length")
			}
			if !tt.wantV2Metrics {
				assert.Zero(t, table.sxHeight, "must not read past declared length")
				assert.Zero(t, table.sCapHeight, "must not read past declared length")
			}
			if !tt.wantV5Metrics {
				assert.Zero(t, table.usLowerOpticalPointSize, "must not read past declared length")
			}

			f.os2 = table
			m := (&Font{font: f}).OS2Metrics()
			assert.True(t, m.Present)
			assert.Equal(t, tt.wantTypoWin, m.HasTypoWinMetrics)
			if !tt.wantTypoWin {
				assert.Zero(t, m.TypoAscender, "OS2Metrics must not surface bytes read past the table boundary")
				assert.Zero(t, m.WinDescent)
			}
			if !tt.wantV2Metrics {
				assert.Zero(t, m.XHeight, "OS2Metrics must not surface a v2+ field the table's own version doesn't define")
				assert.Zero(t, m.CapHeight)
			}
		})
	}
}

// TestOS2Metrics_UseTypoMetrics_RequiresPresence asserts UseTypoMetrics is
// false for a v4 table with bit 7 set but no Typo/Win fields.
func TestOS2Metrics_UseTypoMetrics_RequiresPresence(t *testing.T) {
	f := &font{os2: &os2Table{length: os2LenV0Apple, version: 4, fsSelection: 0x0080}}
	m := (&Font{font: f}).OS2Metrics()
	assert.False(t, m.HasTypoWinMetrics)
	assert.False(t, m.UseTypoMetrics, "bit 7 must not be honored when the typo fields it refers to are absent")
}

// TestOS2_ShortV0RoundTrip asserts a short (68-byte) v0 table is still 68
// bytes after writeOS2 and re-parse.
func TestOS2_ShortV0RoundTrip(t *testing.T) {
	payload := make([]byte, os2LenV0Apple)
	data := append(payload, bytes.Repeat([]byte{0xFF}, 20)...)
	f := &font{trec: &tableRecords{trMap: map[string]*tableRecord{"OS/2": {offset: 0, length: os2LenV0Apple}}}}
	table, err := f.parseOS2Table(newByteReader(bytes.NewReader(data)))
	require.NoError(t, err)
	require.False(t, table.hasTypoWinMetrics())

	var buf bytes.Buffer
	w := newByteWriter(&buf)
	f.os2 = table
	require.NoError(t, f.writeOS2(w))
	require.NoError(t, w.flush())
	assert.Equal(t, os2LenV0Apple, buf.Len(), "writeOS2 must not write more than the source table had")

	reparsed := &font{trec: &tableRecords{trMap: map[string]*tableRecord{"OS/2": {offset: 0, length: uint32(buf.Len())}}}}
	roundTripped, err := reparsed.parseOS2Table(newByteReader(bytes.NewReader(buf.Bytes())))
	require.NoError(t, err)
	assert.False(t, roundTripped.hasTypoWinMetrics(), "round-tripped table must still be short, not promoted to full v0 with zeroed fields")
}

// TestOS2_ProgrammaticTableWritesFullVersion asserts a v4 os2Table built in
// code (length unset) writes its full v4 layout.
func TestOS2_ProgrammaticTableWritesFullVersion(t *testing.T) {
	table := &os2Table{version: 4, sxHeight: 500, sCapHeight: 700, panose10: make([]uint8, 10)}
	require.True(t, table.hasV1Metrics())
	require.True(t, table.hasV2Metrics())
	require.False(t, table.hasV5Metrics())

	var buf bytes.Buffer
	w := newByteWriter(&buf)
	f := &font{os2: table}
	require.NoError(t, f.writeOS2(w))
	require.NoError(t, w.flush())
	assert.Equal(t, int(os2LenV2to4), buf.Len(), "must write the full v4 table, not truncate on an unset length")
}

// TestParseOS2Table_LengthDegrades asserts an OS/2 record with fewer than
// os2LenV0Apple bytes present parses as absent (nil, nil), not an error.
func TestParseOS2Table_LengthDegrades(t *testing.T) {
	tests := []struct {
		name   string
		length uint32
	}{
		{"declared shorter than any defined OS/2 version", 67},
		{"declared far longer, only 67 bytes present", 0xFFFFFFFF},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &font{
				trec: &tableRecords{
					trMap: map[string]*tableRecord{
						"OS/2": {offset: 0, length: tt.length},
					},
				},
			}
			table, err := f.parseOS2Table(newByteReader(bytes.NewReader(make([]byte, 67))))
			require.NoError(t, err)
			assert.Nil(t, table)
		})
	}
}

// TestParseOS2Table_DeclaredLengthCapped asserts an OS/2 record declared far
// longer than maxBoundedTableLen parses its first maxBoundedTableLen bytes.
func TestParseOS2Table_DeclaredLengthCapped(t *testing.T) {
	f := &font{trec: &tableRecords{trMap: map[string]*tableRecord{"OS/2": {offset: 0, length: 0xFFFFFFFF}}}}
	payload := make([]byte, maxBoundedTableLen+100)
	payload[1] = 5 // version 5
	table, err := f.parseOS2Table(newByteReader(bytes.NewReader(payload)))
	require.NoError(t, err)
	require.NotNil(t, table)
	assert.Equal(t, uint32(maxBoundedTableLen), table.length)
	assert.True(t, table.hasV5Metrics())
}

// TestParseOS2Table_FileEndsEarly asserts an OS/2 record whose declared
// length runs past the end of the file parses the field groups present (and
// writes back only those), and is absent if fewer than os2LenV0Apple bytes
// are present.
func TestParseOS2Table_FileEndsEarly(t *testing.T) {
	declared := &tableRecords{trMap: map[string]*tableRecord{"OS/2": {offset: 0, length: os2LenV5}}}

	f := &font{trec: declared}
	payload := bytes.Repeat([]byte{0xAB}, 80) // file ends 20 bytes short
	payload[0], payload[1] = 0, 4             // version 4
	table, err := f.parseOS2Table(newByteReader(bytes.NewReader(payload)))
	require.NoError(t, err)
	require.NotNil(t, table)
	assert.True(t, table.hasTypoWinMetrics())
	assert.False(t, table.hasV1Metrics())

	var buf bytes.Buffer
	w := newByteWriter(&buf)
	f.os2 = table
	require.NoError(t, f.writeOS2(w))
	require.NoError(t, w.flush())
	assert.Equal(t, int(os2LenV0Microsoft), buf.Len())

	short := &font{trec: declared}
	table, err = short.parseOS2Table(newByteReader(bytes.NewReader(make([]byte, os2LenV0Apple-1))))
	require.NoError(t, err)
	assert.Nil(t, table)
}

// TestGlyphAdvance_TrailingInheritance asserts gids past numberOfHMetrics
// inherit the last advance (FreeSans.ttf: 3722 hMetrics, 3726 glyphs).
func TestGlyphAdvance_TrailingInheritance(t *testing.T) {
	f, err := ParseFile("./testdata/FreeSans.ttf")
	require.NoError(t, err)
	require.Equal(t, 3726, f.NumGlyphs())

	last, ok := f.GlyphAdvance(3721)
	require.True(t, ok, "last explicit hMetrics entry")

	for gid := GlyphIndex(3722); gid <= 3725; gid++ {
		adv, ok := f.GlyphAdvance(gid)
		assert.True(t, ok, "gid %d is valid, past hMetrics but within numGlyphs", gid)
		assert.Equal(t, last, adv, "gid %d inherits the last explicit advance", gid)
	}
}

// TestParseFile_BundledCorpus asserts every font under testdata/ parses and
// reports OS2Metrics().Present.
func TestParseFile_BundledCorpus(t *testing.T) {
	var paths []string
	err := filepath.WalkDir("./testdata", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		switch filepath.Ext(path) {
		case ".ttf", ".otf":
			paths = append(paths, path)
		}
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, paths, "expected at least one bundled font under testdata/")

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			f, err := ParseFile(path)
			require.NoError(t, err)
			assert.True(t, f.OS2Metrics().Present, "bundled fonts are expected to carry an OS/2 table")
		})
	}
}
