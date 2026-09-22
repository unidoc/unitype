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

	adv := f.GlyphAdvance(gids[0])
	assert.Greater(t, adv, uint16(0), "'A' should have a positive advance")
}

func TestGlyphAdvance_OutOfRange(t *testing.T) {
	f, err := ParseFile("./testdata/roboto/Roboto-Regular.ttf")
	require.NoError(t, err)

	// A GID at or beyond the font's real glyph count is out of range and must
	// return 0, per the doc comment — not hmtx's trailing-glyph inheritance
	// value, which only applies to real glyphs beyond numberOfHMetrics.
	invalid := GlyphIndex(f.NumGlyphs())
	assert.Equal(t, uint16(0), f.GlyphAdvance(invalid), "GID at NumGlyphs() is out of range")
	assert.Equal(t, uint16(0), f.GlyphAdvance(invalid+1000), "far out-of-range GID must also return 0")
}

// TestGlyphAdvance_TrailingInheritance covers the one GlyphAdvance branch no
// bundled test font exercises: a VALID gid beyond numberOfHMetrics (hmtx
// stores explicit widths only for the first numberOfHMetrics glyphs; every
// later glyph inherits the last one's advance, per the OpenType hmtx spec).
// Roboto-Regular has numberOfHMetrics == numGlyphs (verified: no two
// consecutive trailing glyphs share an advance width), so it never reaches
// this path — a synthetic Font is the only way to cover it. White-box
// (package unitype) so the private hmtx/maxp fields are reachable directly;
// only what GlyphAdvance itself reads needs to be populated.
func TestGlyphAdvance_TrailingInheritance(t *testing.T) {
	f := &Font{font: &font{
		maxp: &maxpTable{numGlyphs: 5},
		hmtx: &hmtxTable{hMetrics: []longHorMetric{
			{advanceWidth: 100},
			{advanceWidth: 200},
			{advanceWidth: 300}, // numberOfHMetrics == 3; gids 3 and 4 have no entry
		}},
	}}

	assert.Equal(t, uint16(100), f.GlyphAdvance(0), "gid within hMetrics uses its own entry")
	assert.Equal(t, uint16(300), f.GlyphAdvance(2), "last explicit hMetrics entry")
	assert.Equal(t, uint16(300), f.GlyphAdvance(3), "valid gid past hMetrics inherits the last advance")
	assert.Equal(t, uint16(300), f.GlyphAdvance(4), "valid gid past hMetrics inherits the last advance")
	assert.Equal(t, uint16(0), f.GlyphAdvance(5), "gid == numGlyphs is out of range, not trailing")
}
