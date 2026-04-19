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
