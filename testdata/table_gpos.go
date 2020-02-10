/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package truetype

// gposTable represents the Glyph Positioning Table (GPOS).
// The GPOS table provides precise control over glyph placement for sophisticated text layout and
// rendering in each script and language that a font supports.
// (https://docs.microsoft.com/en-us/typography/opentype/spec/gpos).
type gposTable struct {
	majorVersion uint16
	minorVersion uint16

	scriptListOffset  offset16
	featureListOffset offset16
	lookupListOffset  offset16

	// For 1.1 and above:
	featureVariationsOffset offset32
}

func (t *gposTable) Unmarshal(r *byteReader) error {
	err := r.read(&t.minorVersion, t.minorVersion)
	if err != nil {
		return err
	}

	err = r.read(&t.scriptListOffset, &t.featureListOffset, &t.lookupListOffset)
	if err != nil {
		return err
	}

	if t.minorVersion < 1 {
		return nil
	}

	// For 1.1 and above:
	return r.read(&t.featureVariationsOffset)
}

// TODO(gunnsth): Need to construct the actual table containing subtables to know the offsets in write mode?
//    Might be best to load subtables at the time of loading a table.
//    Need to recalculate offsets at the time of writing.

func (t *gposTable) Marshal(w *byteWriter) error {
	err := w.write(t.majorVersion, t.minorVersion, t.scriptListOffset, t.featureListOffset, t.lookupListOffset)
	if err != nil {
		return err
	}

	if t.minorVersion < 1 {
		return nil
	}
	// For 1.1 and above:
	return w.write(t.featureVariationsOffset)
}

type gposScriptListSubtable struct {
}
