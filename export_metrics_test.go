/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRobotoMetrics covers OS2Metrics/HheaMetrics/UnitsPerEm/NumGlyphs/
// GlyphAdvance in one ParseFile of the bundled Roboto-Regular.ttf, since each
// only reads a few header fields off the same parsed Font.
func TestRobotoMetrics(t *testing.T) {
	f, err := ParseFile("./testdata/roboto/Roboto-Regular.ttf")
	require.NoError(t, err)

	t.Run("OS2Metrics", func(t *testing.T) {
		m := f.OS2Metrics()
		assert.True(t, m.Present, "Roboto should have OS/2 table")
		// Asserted by sign/range; XHeight/CapHeight/UseTypoMetrics below are
		// pinned to Roboto-Regular's exact values instead.
		assert.Greater(t, m.TypoAscender, int16(0))
		assert.Less(t, m.TypoDescender, int16(0))
		assert.GreaterOrEqual(t, m.TypoLineGap, int16(0))
		assert.Greater(t, m.WinAscent, uint16(0))
		assert.Greater(t, m.WinDescent, uint16(0))
		// XHeight, CapHeight and UseTypoMetrics come from this specific
		// bundled font file, so assert exact values rather than sign/range.
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

		// Uses the max representable GlyphIndex rather than an offset.
		require.Less(t, f.NumGlyphs(), 0xFFFF, "test assumes NumGlyphs() leaves room below the GlyphIndex max")
		_, ok = f.GlyphAdvance(GlyphIndex(0xFFFF))
		assert.False(t, ok, "far out-of-range GID must also report not-ok")
	})
}

// TestOS2Metrics_UseTypoMetricsVersionGate: UseTypoMetrics honors fsSelection
// bit 7 only for OS/2 version >= 4; the bit is reserved and ignored below
// that. White-box (package unitype) to construct a synthetic os2Table, since
// none of the bundled fonts have the bit set.
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

// truncatedOS2Bytes builds OS/2 table bytes claiming declaredVersion but
// only payloadLen bytes long, followed by sentinel 0xFF bytes standing in
// for whatever unrelated table happens to follow OS/2 in a real font file -
// if parseOS2Table ever reads past its own declared length, it picks these
// up instead of leaving the corresponding fields zero.
func truncatedOS2Bytes(declaredVersion uint16, payloadLen int) []byte {
	payload := make([]byte, payloadLen)
	payload[0] = byte(declaredVersion >> 8) // big-endian, per byteReader.readUint16
	payload[1] = byte(declaredVersion)
	return append(payload, bytes.Repeat([]byte{0xFF}, 20)...)
}

// TestParseOS2Table_Truncated covers every OS/2 version boundary
// (os2LenV0Apple/V0Microsoft/V1/V2to4/V5): a table truncated before a given
// block's fields must leave them zero rather than read past its own declared
// length into whatever bytes follow OS/2 in the file. White-box (package
// unitype) since no bundled font has a truncated OS/2 table.
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
		{"v2-4 (96)", 4, os2LenV2to4, true, true, false},
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
		})
	}
}

// TestOS2Metrics_UseTypoMetrics_RequiresPresence: a table that claims
// version 4+ with fsSelection bit 7 set, but is truncated before
// sTypoAscender..usWinDescent (so those fields are absent, not just zero),
// must report UseTypoMetrics=false - otherwise a caller honoring
// USE_TYPO_METRICS would drive line-height off TypoAscender=0 instead of
// falling back to HheaMetrics.
func TestOS2Metrics_UseTypoMetrics_RequiresPresence(t *testing.T) {
	f := &font{os2: &os2Table{length: os2LenV0Apple, version: 4, fsSelection: 0x0080}}
	m := (&Font{font: f}).OS2Metrics()
	assert.False(t, m.HasTypoWinMetrics)
	assert.False(t, m.UseTypoMetrics, "bit 7 must not be honored when the typo fields it refers to are absent")
}

// TestOS2_ShortV0RoundTrip: a font parsed from a short (68-byte) Apple v0
// OS/2 table must still be a short v0 table after writeOS2 - if writeOS2
// wrote the zero-valued TypoAscender/WinDescent as if they were real values,
// re-parsing would see a full 78-byte table with WinAscent/WinDescent = 0,
// which a renderer could use to clip all glyphs to zero height.
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

// TestParseOS2Table_LengthRangeCheck: a table record claiming a length far
// beyond any defined OS/2 version (os2LenV5 = 100) must be rejected rather
// than trigger an allocation sized off an untrusted file-controlled value.
func TestParseOS2Table_LengthRangeCheck(t *testing.T) {
	f := &font{
		trec: &tableRecords{
			trMap: map[string]*tableRecord{
				"OS/2": {offset: 0, length: 0xFFFFFFFF},
			},
		},
	}
	_, err := f.parseOS2Table(newByteReader(bytes.NewReader(nil)))
	assert.ErrorIs(t, err, errRangeCheck)
}

// TestGlyphAdvance_TrailingInheritance covers a VALID gid beyond
// numberOfHMetrics (hmtx stores explicit widths only for the first
// numberOfHMetrics glyphs; every later glyph inherits the last one's
// advance, per the OpenType hmtx spec). FreeSans.ttf has numberOfHMetrics
// (3722) < numGlyphs (3726), so gids 3722-3725 all reach this branch with ok
// still true.
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
