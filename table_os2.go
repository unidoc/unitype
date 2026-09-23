/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"

	"github.com/sirupsen/logrus"
)

// OS/2 table lengths in bytes, by the fields present at each version, per
// https://learn.microsoft.com/en-us/typography/opentype/spec/os2 . Version 0
// has two historical lengths: Apple's original TrueType table ends at
// usLastCharIndex (68 bytes); the Microsoft/OpenType version adds
// sTypoAscender..usWinDescent (78 bytes). Versions 1, 2-4 and 5 each add a
// further fixed block on top of the previous version's fields.
const (
	os2LenV0Apple     = 68
	os2LenV0Microsoft = 78
	os2LenV1          = 86
	os2LenV2to4       = 96
	os2LenV5          = 100
)

// os2Table represents the OS/2 metrics table. It consists of metrics and other data that are required.
type os2Table struct {
	// length is the table record's declared byte length as parsed, or 0 for
	// a table built programmatically. Never read directly - use
	// effectiveLength() or the hasXxxMetrics() presence checks.
	length uint32

	// Version 0+
	version             uint16
	xAvgCharWidth       int16
	usWeightClass       uint16
	usWidthClass        uint16
	fsType              uint16
	ySubscriptXSize     int16
	ySubscriptYSize     int16
	ySubscriptXOffset   int16
	ySubscriptYOffset   int16
	ySuperscriptXSize   int16
	ySuperscriptYSize   int16
	ySuperscriptXOffset int16
	ySuperscriptYOffset int16
	yStrikeoutSize      int16
	yStrikeoutPosition  int16
	sFamilyClass        int16
	panose10            []uint8 // panose10 len = 10
	ulUnicodeRange1     uint32  // Bits 0-31.
	ulUnicodeRange2     uint32  // Bits 32-63.
	ulUnicodeRange3     uint32  // Bits 64-95.
	ulUnicodeRange4     uint32  // Bits 96-127.
	achVendID           tag
	fsSelection         uint16
	usFirstCharIndex    uint16
	usLastCharIndex     uint16
	sTypoAscender       int16
	sTypoDescender      int16
	sTypoLineGap        int16
	usWinAscent         uint16
	usWinDescent        uint16

	// Version 1-5.
	ulCodePageRange1 uint32 // Bits 0-31
	ulCodePageRange2 uint32 // Bits 32-63.

	// Version 2-5
	sxHeight      int16
	sCapHeight    int16
	usDefaultChar uint16
	usBreakChar   uint16
	usMaxContext  uint16

	// Version 5
	usLowerOpticalPointSize uint16
	usUpperOpticalPointSize uint16
}

// effectiveLength returns t.length if set, else the spec length for
// t.version, so a table built in code is written at its version's full size.
func (t *os2Table) effectiveLength() uint32 {
	if t.length != 0 {
		return t.length
	}
	switch {
	case t.version >= 5:
		return os2LenV5
	case t.version >= 2:
		return os2LenV2to4
	case t.version >= 1:
		return os2LenV1
	default:
		return os2LenV0Microsoft
	}
}

// hasTypoWinMetrics reports whether sTypoAscender/sTypoDescender/
// sTypoLineGap/usWinAscent/usWinDescent were actually present in the source
// data. These fields exist in every version from the 78-byte Microsoft v0
// table onward, so length alone (no version check) gates them.
func (t *os2Table) hasTypoWinMetrics() bool {
	return t.effectiveLength() >= os2LenV0Microsoft
}

// hasV1Metrics reports whether ulCodePageRange1/ulCodePageRange2 were
// actually present in the source data.
func (t *os2Table) hasV1Metrics() bool {
	return t.version >= 1 && t.effectiveLength() >= os2LenV1
}

// hasV2Metrics reports whether sxHeight/sCapHeight/usDefaultChar/
// usBreakChar/usMaxContext were actually present in the source data.
func (t *os2Table) hasV2Metrics() bool {
	return t.version >= 2 && t.effectiveLength() >= os2LenV2to4
}

// hasV5Metrics reports whether usLowerOpticalPointSize/
// usUpperOpticalPointSize were actually present in the source data.
func (t *os2Table) hasV5Metrics() bool {
	return t.version >= 5 && t.effectiveLength() >= os2LenV5
}

// os2MaxTableLen bounds the buffer parseOS2Table allocates for a table
// record's declared length: no defined OS/2 version needs more than
// os2LenV5 (100) bytes, so a much larger declared length can only be a
// corrupt or adversarial table record - treat it as absent rather than
// allocate whatever size the file claims.
const os2MaxTableLen = 1024

// parseOS2Table parses the OS/2 table. `r` is read into a fixed buffer sized
// to the table record's declared length and parsed from that buffer, so a
// truncated table cannot read past its own bytes into whatever table follows
// OS/2 in the file; callers distinguish "absent because truncated" from
// "present and legitimately zero" via hasTypoWinMetrics/hasV1Metrics/
// hasV2Metrics/hasV5Metrics.
func (f *font) parseOS2Table(r *byteReader) (*os2Table, error) {
	tr, has, err := f.seekToTable(r, "OS/2")
	if err != nil {
		return nil, err
	}
	if !has {
		logrus.Debug("OS/2 table not present")
		return nil, nil
	}
	// OS/2 is optional (the !has branch above already treats its absence as
	// fine), so a table whose declared length can't possibly hold any
	// defined version degrades the same way: log and report it as absent
	// rather than failing the whole font over one malformed optional table.
	if tr.length < os2LenV0Apple || tr.length > os2MaxTableLen {
		logrus.Debug("OS/2 table length outside any defined version's range, treating as absent")
		return nil, nil
	}

	var buf []byte
	if err := r.readBytes(&buf, int(tr.length)); err != nil {
		return nil, err
	}
	br := newByteReader(bytes.NewReader(buf))

	t := &os2Table{length: tr.length}
	err = br.read(&t.version, &t.xAvgCharWidth, &t.usWeightClass, &t.usWidthClass, &t.fsType)
	if err != nil {
		return nil, err
	}

	if t.version > 10 {
		logrus.Debug("OS/2 table version range error")
		return nil, errRangeCheck
	}

	err = br.read(&t.ySubscriptXSize, &t.ySubscriptYSize, &t.ySubscriptXOffset, &t.ySubscriptYOffset)
	if err != nil {
		return nil, err
	}

	err = br.read(&t.ySuperscriptXSize, &t.ySuperscriptYSize, &t.ySuperscriptXOffset, &t.ySuperscriptYOffset)
	if err != nil {
		return nil, err
	}

	err = br.read(&t.yStrikeoutSize, &t.yStrikeoutPosition, &t.sFamilyClass)
	if err != nil {
		return nil, err
	}

	err = br.readSlice(&t.panose10, 10)
	if err != nil {
		return nil, err
	}

	err = br.read(&t.ulUnicodeRange1, &t.ulUnicodeRange2, &t.ulUnicodeRange3, &t.ulUnicodeRange4)
	if err != nil {
		return nil, err
	}
	err = br.read(&t.achVendID, &t.fsSelection, &t.usFirstCharIndex, &t.usLastCharIndex)
	if err != nil {
		return nil, err
	}

	if !t.hasTypoWinMetrics() {
		return t, nil
	}
	err = br.read(&t.sTypoAscender, &t.sTypoDescender, &t.sTypoLineGap, &t.usWinAscent, &t.usWinDescent)
	if err != nil {
		return nil, err
	}

	if !t.hasV1Metrics() {
		return t, nil
	}
	err = br.read(&t.ulCodePageRange1, &t.ulCodePageRange2)
	if err != nil {
		return nil, err
	}

	if !t.hasV2Metrics() {
		return t, nil
	}
	err = br.read(&t.sxHeight, &t.sCapHeight, &t.usDefaultChar, &t.usBreakChar, &t.usMaxContext)
	if err != nil {
		return nil, err
	}

	if !t.hasV5Metrics() {
		return t, nil
	}
	err = br.read(&t.usLowerOpticalPointSize, &t.usUpperOpticalPointSize)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// writeOS2 mirrors parseOS2Table's field-presence boundaries (hasTypoWinMetrics/
// hasV1Metrics/hasV2Metrics/hasV5Metrics), not just t.version, so a round trip
// never fabricates fields the source table never had.
func (f *font) writeOS2(w *byteWriter) error {
	if f.os2 == nil {
		return nil
	}
	t := f.os2

	err := w.write(t.version, t.xAvgCharWidth, t.usWeightClass, t.usWidthClass, t.fsType)
	if err != nil {
		return err
	}

	err = w.write(t.ySubscriptXSize, t.ySubscriptYSize, t.ySubscriptXOffset, t.ySubscriptYOffset)
	if err != nil {
		return err
	}

	err = w.write(t.ySuperscriptXSize, t.ySuperscriptYSize, t.ySuperscriptXOffset, t.ySuperscriptYOffset)
	if err != nil {
		return err
	}

	err = w.write(t.yStrikeoutSize, t.yStrikeoutPosition, t.sFamilyClass)
	if err != nil {
		return err
	}

	err = w.writeSlice(t.panose10)
	if err != nil {
		return err
	}

	err = w.write(t.ulUnicodeRange1, t.ulUnicodeRange2, t.ulUnicodeRange3, t.ulUnicodeRange4)
	if err != nil {
		return err
	}
	err = w.write(t.achVendID, t.fsSelection, t.usFirstCharIndex, t.usLastCharIndex)
	if err != nil {
		return err
	}

	if !t.hasTypoWinMetrics() {
		return nil
	}
	err = w.write(t.sTypoAscender, t.sTypoDescender, t.sTypoLineGap, t.usWinAscent, t.usWinDescent)
	if err != nil {
		return err
	}

	if !t.hasV1Metrics() {
		return nil
	}
	err = w.write(t.ulCodePageRange1, t.ulCodePageRange2)
	if err != nil {
		return err
	}

	if !t.hasV2Metrics() {
		return nil
	}
	err = w.write(t.sxHeight, t.sCapHeight, t.usDefaultChar, t.usBreakChar, t.usMaxContext)
	if err != nil {
		return err
	}

	if !t.hasV5Metrics() {
		return nil
	}
	return w.write(t.usLowerOpticalPointSize, t.usUpperOpticalPointSize)
}
