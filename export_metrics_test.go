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
	// XHeight/CapHeight/UseTypoMetrics were previously unasserted, so a wrong
	// version gate (o.version >= 2) or a wrong fsSelection bit mask would pass
	// silently. Unlike the ascent/descent fields above, these three come
	// straight from this specific bundled font file rather than "any
	// reasonable font," so exact values are the right check here (verified
	// directly against testdata/roboto/Roboto-Regular.ttf, not assumed).
	assert.Equal(t, int16(1082), m.XHeight, "Roboto-Regular sxHeight")
	assert.Equal(t, int16(1456), m.CapHeight, "Roboto-Regular sCapHeight")
	assert.False(t, m.UseTypoMetrics, "Roboto-Regular fsSelection bit 7 (USE_TYPO_METRICS) is unset")
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
