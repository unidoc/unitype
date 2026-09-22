/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

// OS2Metrics exposes the OS/2 table fields most relevant to line-layout.
// Ascender/Descender/LineGap are in font design units (FUnits). Divide by
// UnitsPerEm to convert to em; multiply by fontSize (px) for pixel values.
//
// UseTypoMetrics reports fsSelection bit 7 (USE_TYPO_METRICS). When set, the
// font author intends sTypoAscender/sTypoDescender/sTypoLineGap to drive
// line-height rather than the legacy usWinAscent/usWinDescent values returned
// by many TrueType rasterizers.
//
// https://docs.microsoft.com/en-us/typography/opentype/spec/os2
type OS2Metrics struct {
	TypoAscender   int16
	TypoDescender  int16
	TypoLineGap    int16
	WinAscent      uint16
	WinDescent     uint16
	XHeight        int16 // sxHeight (OS/2 v2+); 0 if unavailable
	CapHeight      int16 // sCapHeight (OS/2 v2+); 0 if unavailable
	UseTypoMetrics bool  // fsSelection bit 7
	Present        bool  // false if the font has no OS/2 table
}

// HheaMetrics exposes the hhea table fields most relevant to line-layout.
// Ascender/Descender/LineGap are in font design units (FUnits).
//
// https://docs.microsoft.com/en-us/typography/opentype/spec/hhea
type HheaMetrics struct {
	Ascender int16
	Descender int16
	LineGap   int16
	Present   bool
}

// OS2Metrics returns the OS/2 table metrics. Present is false if the font
// lacks an OS/2 table (some older TrueType fonts do).
func (f *Font) OS2Metrics() OS2Metrics {
	if f.font == nil || f.font.os2 == nil {
		return OS2Metrics{}
	}
	o := f.font.os2
	m := OS2Metrics{
		TypoAscender:   o.sTypoAscender,
		TypoDescender:  o.sTypoDescender,
		TypoLineGap:    o.sTypoLineGap,
		WinAscent:      o.usWinAscent,
		WinDescent:     o.usWinDescent,
		UseTypoMetrics: (o.fsSelection & 0x0080) != 0, // bit 7
		Present:        true,
	}
	if o.version >= 2 {
		m.XHeight = o.sxHeight
		m.CapHeight = o.sCapHeight
	}
	return m
}

// HheaMetrics returns the hhea table metrics. Present is false if the font
// lacks an hhea table.
func (f *Font) HheaMetrics() HheaMetrics {
	if f.font == nil || f.font.hhea == nil {
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
	if f.font == nil || f.font.head == nil {
		return 0
	}
	return f.font.head.unitsPerEm
}

// NumGlyphs returns the total number of glyphs in the font.
func (f *Font) NumGlyphs() int {
	if f.font == nil || f.font.maxp == nil {
		return 0
	}
	return int(f.font.maxp.numGlyphs)
}

// GlyphAdvance returns the horizontal advance width of glyph gid in FUnits.
// Returns 0 for out-of-range GIDs. hmtx stores advance widths only for the
// first numberOfHMetrics glyphs; trailing glyphs inherit the last advance.
func (f *Font) GlyphAdvance(gid GlyphIndex) uint16 {
	if f.font == nil || f.font.hmtx == nil {
		return 0
	}
	if int(gid) >= f.NumGlyphs() {
		return 0
	}
	h := f.font.hmtx
	if len(h.hMetrics) == 0 {
		return 0
	}
	if int(gid) < len(h.hMetrics) {
		return h.hMetrics[gid].advanceWidth
	}
	return h.hMetrics[len(h.hMetrics)-1].advanceWidth
}
