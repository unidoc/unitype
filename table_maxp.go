/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"github.com/sirupsen/logrus"
)

// maxpTable represents the Maximum Profile (maxp) table.
// This table establishes the memory requirements for the font.
type maxpTable struct {
	// Version 0.5 and above:
	version   fixed
	numGlyphs uint16

	// Version 1.0 and above:
	maxPoints             uint16
	maxContours           uint16
	maxCompositePoints    uint16
	maxCompositeContours  uint16
	maxZones              uint16
	maxTwilightPoints     uint16
	maxStorage            uint16
	maxFunctionDefs       uint16
	maxInstructionDefs    uint16
	maxStackElements      uint16
	maxSizeOfInstructions uint16
	maxComponentElements  uint16
	maxComponentDepth     uint16
}

// maxpTableV1Len is maxp's byte size at version 1.0, per
// https://learn.microsoft.com/en-us/typography/opentype/spec/maxp . Version
// 0.5 (6 bytes: version+numGlyphs only) is rejected below, so every table
// this parser accepts commits to the full v1.0 layout.
const maxpTableV1Len = 32

func (f *font) parseMaxp(r *byteReader) (*maxpTable, error) {
	tr, has, err := f.seekToTable(r, "maxp")
	if err != nil {
		return nil, err
	}
	if !has {
		logrus.Debug("maxp table not present")
		return nil, nil
	}

	t := &maxpTable{}

	err = r.read(&t.version, &t.numGlyphs)
	if err != nil {
		return nil, err
	}

	if t.version < 0x00010000 {
		logrus.Debug("Range check error")
		return nil, errRangeCheck
	}
	if tr.length < maxpTableV1Len {
		logrus.Debug("maxp table shorter than the version 1.0 required length")
		return nil, errRangeCheck
	}

	err = r.read(&t.maxPoints, &t.maxContours, &t.maxCompositePoints, &t.maxCompositeContours)
	if err != nil {
		return nil, err
	}

	err = r.read(&t.maxZones, &t.maxTwilightPoints, &t.maxStorage, &t.maxFunctionDefs, &t.maxInstructionDefs)
	if err != nil {
		return nil, err
	}

	return t, r.read(&t.maxStackElements, &t.maxSizeOfInstructions, &t.maxComponentElements, &t.maxComponentDepth)
}

func (f *font) writeMaxp(w *byteWriter) error {
	if f.maxp == nil {
		return errRequiredField
	}
	t := f.maxp
	err := w.write(t.version, t.numGlyphs)
	if err != nil {
		return err
	}

	if t.version < 0x00010000 {
		logrus.Debug("Range check error")
		return errRangeCheck
	}

	err = w.write(t.maxPoints, t.maxContours, t.maxCompositePoints, t.maxCompositeContours)
	if err != nil {
		return err
	}

	err = w.write(t.maxZones, t.maxTwilightPoints, t.maxStorage, t.maxFunctionDefs, t.maxInstructionDefs)
	if err != nil {
		return err
	}

	return w.write(t.maxStackElements, t.maxSizeOfInstructions, t.maxComponentElements, t.maxComponentDepth)
}
