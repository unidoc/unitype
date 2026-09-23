/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOS2Metrics_Roboto(t *testing.T) {
	f, err := ParseFile("./testdata/roboto/Roboto-Regular.ttf")
	require.NoError(t, err)

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
	// XHeight, CapHeight and UseTypoMetrics come from this specific bundled
	// font file, so assert exact values rather than sign/range.
	assert.Equal(t, int16(1082), m.XHeight, "Roboto-Regular sxHeight")
	assert.Equal(t, int16(1456), m.CapHeight, "Roboto-Regular sCapHeight")
	assert.False(t, m.UseTypoMetrics, "Roboto-Regular fsSelection bit 7 (USE_TYPO_METRICS) is unset")
}

// TestOS2Metrics_UseTypoMetricsVersionGate exercises the fsSelection bit 7
// mask together with the OS/2 version gate: bit 7 (USE_TYPO_METRICS) is only
// defined from version 4 onward, so a v0-v3 font with the bit set (a
// reserved field, potentially garbage) must NOT report UseTypoMetrics=true.
// None of the bundled test fonts have the bit set (Roboto/FreeSans/wts11 all
// report false), so this needs a synthetic os2Table to cover the mask and
// the gate at all — white-box (package unitype) for that reason.
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

func TestHheaMetrics_Roboto(t *testing.T) {
	f, err := ParseFile("./testdata/roboto/Roboto-Regular.ttf")
	require.NoError(t, err)

	m := f.HheaMetrics()
	assert.True(t, m.Present, "Roboto should have hhea table")
	assert.Greater(t, m.Ascender, int16(0))
	assert.Less(t, m.Descender, int16(0))
}

func TestUnitsPerEm_Roboto(t *testing.T) {
	f, err := ParseFile("./testdata/roboto/Roboto-Regular.ttf")
	require.NoError(t, err)
	assert.Equal(t, uint16(2048), f.UnitsPerEm(), "Roboto uses 2048 UPEM")
}

func TestNumGlyphs_Roboto(t *testing.T) {
	f, err := ParseFile("./testdata/roboto/Roboto-Regular.ttf")
	require.NoError(t, err)
	assert.Greater(t, f.NumGlyphs(), 100)
}

func TestGlyphAdvance_Roboto(t *testing.T) {
	f, err := ParseFile("./testdata/roboto/Roboto-Regular.ttf")
	require.NoError(t, err)

	// Look up glyph for 'A' and verify it has a non-zero advance.
	gids := f.LookupRunes([]rune{'A'})
	require.Len(t, gids, 1)
	require.NotEqual(t, GlyphIndex(0), gids[0], "'A' should resolve to a glyph")

	adv, ok := f.GlyphAdvance(gids[0])
	assert.True(t, ok)
	assert.Greater(t, adv, uint16(0), "'A' should have a positive advance")
}

func TestGlyphAdvance_OutOfRange(t *testing.T) {
	f, err := ParseFile("./testdata/roboto/Roboto-Regular.ttf")
	require.NoError(t, err)

	// A GID at or beyond the font's real glyph count is out of range: ok must
	// be false, distinguishing it from a real zero-width glyph.
	invalid := GlyphIndex(f.NumGlyphs())
	_, ok := f.GlyphAdvance(invalid)
	assert.False(t, ok, "GID at NumGlyphs() is out of range")
	_, ok = f.GlyphAdvance(invalid + 1000)
	assert.False(t, ok, "far out-of-range GID must also report not-ok")
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
