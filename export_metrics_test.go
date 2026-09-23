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
		// Roboto design is 2048 UPEM, typical sTypo values are around
		// +1900 ascender / -500 descender. We don't want to hard-code exact
		// numbers (font may be updated), just assert ballpark positivity/sign.
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
		_, ok = f.GlyphAdvance(invalid + 1000)
		assert.False(t, ok, "far out-of-range GID must also report not-ok")
	})
}

// TestOS2Metrics_UseTypoMetricsVersionGate exercises the fsSelection bit 7
// mask together with the OS/2 version gate: bit 7 (USE_TYPO_METRICS) is only
// defined from version 4 onward, so a v0-v3 font with the bit set (a
// reserved field, potentially garbage) must NOT report UseTypoMetrics=true.
// None of the bundled test fonts have the bit set (Roboto/FreeSans/wts11 all
// report false), so this needs a synthetic os2Table to cover the mask and
// the gate at all - white-box (package unitype) for that reason.
func TestOS2Metrics_UseTypoMetricsVersionGate(t *testing.T) {
	newFont := func(version, fsSelection uint16) *Font {
		return &Font{font: &font{os2: &os2Table{version: version, fsSelection: fsSelection}}}
	}

	v3BitSet := newFont(3, 0x0080)
	assert.False(t, v3BitSet.OS2Metrics().UseTypoMetrics,
		"bit 7 is reserved before v4 and must not be read as USE_TYPO_METRICS")

	v4BitSet := newFont(4, 0x0080)
	assert.True(t, v4BitSet.OS2Metrics().UseTypoMetrics, "bit 7 is defined from v4 onward")

	v4BitClear := newFont(4, 0x0000)
	assert.False(t, v4BitClear.OS2Metrics().UseTypoMetrics)
}

// TestParseOS2Table_ShortAppleV0Table covers a genuine Apple-style 68-byte
// OS/2 version 0 table, which ends before sTypoAscender/sTypoDescender/
// sTypoLineGap/usWinAscent/usWinDescent (the 78-byte Microsoft/OpenType v0
// table adds those five fields). Without a table-length check, parseOS2Table
// would read those fields from whatever bytes follow OS/2 in the file. No
// bundled test font has a short v0 table, so this constructs one directly:
// white-box (package unitype) to call parseOS2Table with a synthetic
// tableRecords + byteReader instead of a real font file.
func TestParseOS2Table_ShortAppleV0Table(t *testing.T) {
	// 68 bytes: version=0 then zeros through usLastCharIndex (offset 68),
	// where a genuine Apple v0 table ends.
	payload := make([]byte, 68)
	// Sentinel bytes immediately after the table, standing in for whatever
	// unrelated table happens to follow OS/2 in a real font file - if
	// parseOS2Table reads past the table's declared length, it would pick
	// these up as sTypoAscender/usWinDescent instead of leaving them zero.
	trailing := []byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}
	data := append(payload, trailing...)

	f := &font{
		trec: &tableRecords{
			trMap: map[string]*tableRecord{
				"OS/2": {offset: 0, length: 68},
			},
		},
	}
	r := newByteReader(bytes.NewReader(data))

	table, err := f.parseOS2Table(r)
	require.NoError(t, err)
	require.NotNil(t, table)

	assert.Equal(t, uint16(0), table.version)
	assert.Zero(t, table.sTypoAscender, "short v0 table must not read past its declared length")
	assert.Zero(t, table.sTypoDescender)
	assert.Zero(t, table.sTypoLineGap)
	assert.Zero(t, table.usWinAscent)
	assert.Zero(t, table.usWinDescent)

	f.os2 = table
	m := (&Font{font: f}).OS2Metrics()
	assert.True(t, m.Present)
	assert.Equal(t, uint16(0), m.Version)
	assert.Zero(t, m.TypoAscender, "OS2Metrics must not surface bytes read past the table boundary")
	assert.Zero(t, m.WinDescent)
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
