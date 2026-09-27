/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

// OS2Metrics exposes the OS/2 table fields most relevant to line-layout, in
// font design units (FUnits); divide by UnitsPerEm to convert to em.
//
// HasTypoWinMetrics is false for a short (68-byte) version 0 table, which
// ends before the Typo/Win fields; fall back to HheaMetrics in that case.
//
// https://docs.microsoft.com/en-us/typography/opentype/spec/os2
type OS2Metrics struct {
	Version           uint16
	TypoAscender      int16
	TypoDescender     int16
	TypoLineGap       int16
	WinAscent         uint16
	WinDescent        uint16
	HasTypoWinMetrics bool  // true if TypoAscender..WinDescent are from the source table, not absent-and-zero
	XHeight           int16 // sxHeight; only defined for Version >= 2, else 0
	CapHeight         int16 // sCapHeight; only defined for Version >= 2, else 0
	UseTypoMetrics    bool  // fsSelection bit 7; false unless Version >= 4 and HasTypoWinMetrics
	Present           bool  // false if the font has no OS/2 table
}

// HheaMetrics exposes the hhea table fields most relevant to line-layout.
// Ascender/Descender/LineGap are in font design units (FUnits).
//
// https://docs.microsoft.com/en-us/typography/opentype/spec/hhea
type HheaMetrics struct {
	Ascender  int16
	Descender int16
	LineGap   int16
	// Present is always true for a Font from Parse/ParseFile (hhea is
	// required); false only for a Font not built by Parse.
	Present bool
}

// OS2Metrics returns the OS/2 table metrics. Present is false if the font
// lacks an OS/2 table (some older TrueType fonts do).
func (f *Font) OS2Metrics() OS2Metrics {
	if f.font.os2 == nil {
		return OS2Metrics{}
	}
	o := f.font.os2
	m := OS2Metrics{
		Version:           o.version,
		HasTypoWinMetrics: o.hasTypoWinMetrics(),
		Present:           true,
	}
	if m.HasTypoWinMetrics {
		m.TypoAscender = o.sTypoAscender
		m.TypoDescender = o.sTypoDescender
		m.TypoLineGap = o.sTypoLineGap
		m.WinAscent = o.usWinAscent
		m.WinDescent = o.usWinDescent
		m.UseTypoMetrics = o.version >= 4 && (o.fsSelection&0x0080) != 0
	}
	if o.hasV2Metrics() {
		m.XHeight = o.sxHeight
		m.CapHeight = o.sCapHeight
	}
	return m
}

// HheaMetrics returns the hhea table metrics; see the Present field comment.
func (f *Font) HheaMetrics() HheaMetrics {
	if f.font.hhea == nil {
		return HheaMetrics{}
	}
	h := f.font.hhea
	return HheaMetrics{
		Ascender:  int16(h.ascender),
		Descender: int16(h.descender),
		LineGap:   int16(h.lineGap),
		Present:   true,
	}
}

// UnitsPerEm returns the font's units-per-em from the head table. Returns 0
// if the head table is missing (should not happen for a valid font).
func (f *Font) UnitsPerEm() uint16 {
	if f.font.head == nil {
		return 0
	}
	return f.font.head.unitsPerEm
}

// NumGlyphs returns the total number of glyphs in the font, or 0 if the maxp
// table is missing (never the case for a Font from Parse/ParseFile).
func (f *Font) NumGlyphs() int {
	if f.font.maxp == nil {
		return 0
	}
	return int(f.font.maxp.numGlyphs)
}

// GlyphAdvance returns the horizontal advance width of glyph gid in FUnits.
// ok is false for an out-of-range gid or a font with no hmtx table. hmtx
// stores advance widths only for the first numberOfHMetrics glyphs; trailing
// glyphs inherit the last advance.
func (f *Font) GlyphAdvance(gid GlyphIndex) (uint16, bool) {
	if f.font.hmtx == nil {
		return 0, false
	}
	if int(gid) >= f.NumGlyphs() {
		return 0, false
	}
	h := f.font.hmtx
	if len(h.hMetrics) == 0 {
		return 0, false
	}
	if int(gid) < len(h.hMetrics) {
		return h.hMetrics[gid].advanceWidth, true
	}
	return h.hMetrics[len(h.hMetrics)-1].advanceWidth, true
}
