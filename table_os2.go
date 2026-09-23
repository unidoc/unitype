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
	// length is the table record's declared byte length as parsed (or, for a
	// table built programmatically rather than parsed, the length writeOS2
	// would produce for the fields actually populated). It records how far
	// into the version's full field layout the source data actually went,
	// so a table that claims a version but was truncated before that
	// version's fields end can be told apart from a genuine short table of
	// an earlier version. Never read directly by callers - use
	// hasTypoWinMetrics()/hasV2Metrics()/hasV5Metrics().
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

// hasTypoWinMetrics reports whether sTypoAscender/sTypoDescender/
// sTypoLineGap/usWinAscent/usWinDescent were actually present in the source
// data (length >= os2LenV0Microsoft), as opposed to being zero because the
// table was a short Apple-style version 0 table.
func (t *os2Table) hasTypoWinMetrics() bool {
	return t.length >= os2LenV0Microsoft
}

// hasV2Metrics reports whether sxHeight/sCapHeight/usDefaultChar/
// usBreakChar/usMaxContext were actually present in the source data.
func (t *os2Table) hasV2Metrics() bool {
	return t.length >= os2LenV2to4
}

// hasV5Metrics reports whether usLowerOpticalPointSize/
// usUpperOpticalPointSize were actually present in the source data.
func (t *os2Table) hasV5Metrics() bool {
	return t.length >= os2LenV5
}

// os2MaxTableLen bounds the buffer parseOS2Table allocates for a table
// record's declared length: no defined OS/2 version needs more than
// os2LenV5 (100) bytes, so a much larger declared length can only be a
// corrupt or adversarial table record - reject it rather than allocate
// whatever size the file claims.
const os2MaxTableLen = 1024

// parseOS2Table parses the OS/2 table. Every read is bounded to the table
// record's declared length: `r` is read into a fixed buffer once and parsed
// from that buffer, rather than the shared file-wide byteReader, specifically
// so a truncated table (short v0, or a table that claims a version but was
// cut off before that version's fields end) cannot read past its own bytes
// into whatever table happens to follow OS/2 in the file. A short read
// leaves the fields for versions/blocks beyond that point at zero; callers
// distinguish "absent because truncated" from "present and legitimately
// zero" via hasTypoWinMetrics/hasV2Metrics/hasV5Metrics.
func (f *font) parseOS2Table(r *byteReader) (*os2Table, error) {
	tr, has, err := f.seekToTable(r, "OS/2")
	if err != nil {
		return nil, err
	}
	if !has {
		logrus.Debug("OS/2 table not present")
		return nil, nil
	}
	if tr.length < os2LenV0Apple {
		logrus.Debug("OS/2 table shorter than the minimum defined length")
		return nil, errRangeCheck
	}
	if tr.length > os2MaxTableLen {
		logrus.Debug("OS/2 table length range error")
		return nil, errRangeCheck
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

	if tr.length < os2LenV1 {
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

// writeOS2 mirrors parseOS2Table's truncation: it stops writing at the same
// boundary the source table was actually parsed to (hasTypoWinMetrics/
// hasV2Metrics/hasV5Metrics), not just t.version. Without this, parsing a
// short (68-byte) v0 table - which leaves sTypoAscender..usWinDescent at
// zero because they were never in the source - and then writing it back out
// would silently fabricate a 78-byte table with usWinAscent/usWinDescent = 0,
// which is a real (not absent) value renderers use to clip glyphs.
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

	// A zero-value os2Table (length never set, e.g. built programmatically
	// rather than parsed) has length 0, which is < os2LenV0Microsoft - write
	// it as a short v0 table rather than silently promoting it to full v0,
	// matching parseOS2Table's own "absent, not zero" semantics.
	if !t.hasTypoWinMetrics() {
		return nil
	}
	err = w.write(t.sTypoAscender, t.sTypoDescender, t.sTypoLineGap, t.usWinAscent, t.usWinDescent)
	if err != nil {
		return err
	}

	if t.length < os2LenV1 {
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
