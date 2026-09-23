/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

// OS2Metrics exposes the OS/2 table fields most relevant to line-layout.
// Ascender/Descender/LineGap are in font design units (FUnits). Divide by
// UnitsPerEm to convert to em; multiply by fontSize (px) for pixel values.
//
// Version distinguishes "field not defined at this OS/2 version" from "field
// defined and legitimately zero" for every version-gated field below -
// callers that need to tell the two apart should check Version themselves;
// a bare zero value cannot express the difference.
//
// TypoAscender/TypoDescender/TypoLineGap/WinAscent/WinDescent read zero if
// the table is a short (68-byte) Apple-style version 0 OS/2 table, which ends
// before these fields; Version alone does not distinguish this case from a
// full-length version 0 table, since both report Version == 0.
//
// UseTypoMetrics reports fsSelection bit 7 (USE_TYPO_METRICS). When set, the
// font author intends sTypoAscender/sTypoDescender/sTypoLineGap to drive
// line-height rather than the legacy usWinAscent/usWinDescent values returned
// by many TrueType rasterizers. Bit 7 is only defined from OS/2 version 4
// onward; in versions 0-3 it is reserved and may hold unrelated data, so this
// field is always false for those versions regardless of the bit's value.
//
// https://docs.microsoft.com/en-us/typography/opentype/spec/os2
type OS2Metrics struct {
	Version        uint16
	TypoAscender   int16
	TypoDescender  int16
	TypoLineGap    int16
	WinAscent      uint16
	WinDescent     uint16
	XHeight        int16 // sxHeight; only defined for Version >= 2, else 0
	CapHeight      int16 // sCapHeight; only defined for Version >= 2, else 0
	UseTypoMetrics bool  // fsSelection bit 7
	Present        bool  // false if the font has no OS/2 table
}

// HheaMetrics exposes the hhea table fields most relevant to line-layout.
// Ascender/Descender/LineGap are in font design units (FUnits).
//
// https://docs.microsoft.com/en-us/typography/opentype/spec/hhea
type HheaMetrics struct {
	Ascender  int16
	Descender int16
	LineGap   int16
	// Present is always true for a Font returned by Parse/ParseFile: hhea is
	// a required table and parsing already fails before construction if it
	// is missing. It is false only for a Font hand-constructed without going
	// through Parse (as some white-box tests in this package do).
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
		Version:        o.version,
		TypoAscender:   o.sTypoAscender,
		TypoDescender:  o.sTypoDescender,
		TypoLineGap:    o.sTypoLineGap,
		WinAscent:      o.usWinAscent,
		WinDescent:     o.usWinDescent,
		UseTypoMetrics: o.version >= 4 && (o.fsSelection&0x0080) != 0, // bit 7, defined from v4
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

// NumGlyphs returns the total number of glyphs in the font. Returns 0 if the
// maxp table is missing; for a Font from Parse/ParseFile this cannot happen
// (maxp is required and parsing fails first) - the nil check only guards a
// Font hand-constructed without going through Parse.
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
