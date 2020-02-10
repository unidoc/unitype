/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package truetype

type glyphClass int

const (
	baseGlyph glyphClass = 1 + iota
	ligatureGlyph
	markGlyph
	componentGlyph
)

type gdefTable struct {
	offset int64

	// Version 1.0+
	majorVersion uint16
	minorVersion uint16

	// Offset of subtables.  Offset specified with respect to
	// beginning of GDEF table header.
	glyphClassDefOffset      offset16
	attachListOffset         offset16
	ligCaretListOffset       offset16
	markAttachClassDefOffset offset16

	// Additional for 1.2 (minorVersion>=2).
	markGlyphSetsDefOffset offset16

	// Additional for 1.3 (minorVersion>=3).
	itemVarStoreOffset offset32
}

func (gd *gdefTable) Unmarshal(br *byteReader) error {
	gd.offset = br.Offset()

	err := br.read(&gd.majorVersion, &gd.minorVersion)
	if err != nil {
		return err
	}

	err = br.read(&gd.glyphClassDefOffset, &gd.attachListOffset, &gd.ligCaretListOffset, &gd.markAttachClassDefOffset)
	if err != nil {
		return err
	}

	if gd.minorVersion < 2 {
		return nil
	}

	// Version 1.2 and above:
	err = br.read(&gd.markGlyphSetsDefOffset)
	if err != nil {
		return err
	}

	if gd.minorVersion < 3 {
		return nil
	}

	// Version 1.3 and above:
	return br.read(&gd.itemVarStoreOffset)

	/*
		var err error
		gd.majorVersion, err = br.readUint16()
		if err != nil {
			return err
		}

		gd.minorVersion, err = br.readUint16()
		if err != nil {
			return err
		}

		gd.glyphClassDefOffset, err = br.readOffset16()
		if err != nil {
			return err
		}

		gd.attachListOffset, err = br.readOffset16()
		if err != nil {
			return err
		}

		gd.ligCaretListOffset, err = br.readOffset16()
		if err != nil {
			return err
		}

		gd.markAttachClassDefOffset, err = br.readOffset16()
		if err != nil {
			return err
		}

		if gd.minorVersion < 2 {
			return nil
		}
		// For 1.2+

		gd.markGlyphSetsDefOffset, err = br.readOffset16()
		if err != nil {
			return err
		}

		if gd.minorVersion < 3 {
			return nil
		}
		// For 1.3+

		gd.itemVarStoreOffset, err = br.readOffset32()
		if err != nil {
			return err
		}

		return nil
	*/
}

// GDEF GlyphClassDef subtable.
type gdefGlyphClassDef struct {
}

// GDEF AttachmentList subtable.
type gdefAttachmentList struct {
}

// GDEF LigatureCaretList subtable.
type gdefLigatureCaretList struct {
}

// GDEF MarkAttachClassDef subtable.
type gdefMarkAttachClassDef struct {
}

// GDEF MarkGlyphSetsTable subtable.
type gdefMarkGlyphSetsTable struct {
}

// GDEF ItemVariationStore subtable.
type gdefItemVariationStore struct {
}
