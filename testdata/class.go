/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package truetype

/*

type classDefFormat1Table struct {
	classFormat     uint16 // = 1.
	startGlyphID    uint16
	glyphCount      uint16
	classValueArray []uint16 // numGlyphCount entries
}

type classDefFormat2Table struct {
	classFormat       uint16 // = 2
	classRangeCount   uint16
	classRangeRecords []classRangeRecord // size: classRangeCount
}

type classRangeRecord struct {
	startGlyphID uint16
	endGlyphID   uint16
	class        uint16
}

func (cf1 *classDefFormat1Table) Unmarshal(br *byteReader) error {
	var err error
	cf1.classFormat, err = br.readUint16()
	if err != nil {
		return err
	}

	cf1.startGlyphID, err = br.readUint16()
	if err != nil {
		return err
	}

	cf1.glyphCount, err = br.readUint16()
	if err != nil {
		return err
	}

	for i := 0; i < int(cf1.glyphCount); i++ {
		val, err := br.readUint16()
		if err != nil {
			return err
		}
		cf1.classValueArray = append(cf1.classValueArray, val)
	}

	return nil
}

func (cf1 *classDefFormat1Table) Marshal(bw *byteWriter) error {
	err := bw.writeUint16(cf1.classFormat)
	if err != nil {
		return err
	}
	err = bw.writeUint16(cf1.startGlyphID)
	if err != nil {
		return err
	}
	err = bw.writeUint16(cf1.glyphCount)
	if err != nil {
		return err
	}
	for _, cv := range cf1.classValueArray {
		err := bw.writeUint16(cv)
		if err != nil {
			return err
		}
	}

	return nil
}

func (cf2 *classDefFormat2Table) Unmarshal(br *byteReader) error {
	var err error
	cf2.classFormat, err = br.readUint16()
	if err != nil {
		return err
	}

	cf2.classRangeCount, err = br.readUint16()
	if err != nil {
		return err
	}

	for i := 0; i < int(cf2.classRangeCount); i++ {
		var crRec classRangeRecord
		err := crRec.Unmarshal(br)
		if err != nil {
			return err
		}
		cf2.classRangeRecords = append(cf2.classRangeRecords, crRec)
	}

	return nil
}

func (cf2 *classDefFormat2Table) Marshal(bw *byteWriter) error {
	err := bw.writeUint16(cf2.classFormat)
	if err != nil {
		return err
	}
	err = bw.writeUint16(cf2.classRangeCount)
	if err != nil {
		return err
	}
	for _, crRec := range cf2.classRangeRecords {
		err := crRec.Marshal(bw)
		if err != nil {
			return err
		}
	}

	return nil
}

func (crr *classRangeRecord) Unmarshal(br *byteReader) error {
	var err error
	crr.startGlyphID, err = br.readUint16()
	if err != nil {
		return err
	}
	crr.endGlyphID, err = br.readUint16()
	if err != nil {
		return err
	}
	crr.class, err = br.readUint16()
	return err
}

func (crr *classRangeRecord) Marshal(bw *byteWriter) error {
	err := bw.writeUint16(crr.startGlyphID)
	if err != nil {
		return err
	}
	err = bw.writeUint16(crr.endGlyphID)
	if err != nil {
		return err
	}
	return bw.writeUint16(crr.class)
}
*/
