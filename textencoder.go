package unitype

import (
	"bytes"
	"fmt"

	"github.com/sirupsen/logrus"
)

// TODO(gunnsth): Make another type: FontEncoder ?  Or simply make *Font implement the interface.

//var _ textencoding.TextEncoder = &font{}

func (f *font) String() string {
	var b bytes.Buffer

	if f.trec != nil {
		b.WriteString(fmt.Sprintf("trec: present with %d table records\n", len(f.trec.list)))
		for _, tr := range f.trec.list {
			b.WriteString(fmt.Sprintf("%s: %.2f kB\n", tr.tableTag.String(), float64(tr.length)/1024))
			b.WriteString(fmt.Sprintf("%s: %d B\n", tr.tableTag.String(), tr.length))
		}
	}
	b.WriteString("--\n")
	if f.hhea != nil {
		b.WriteString(fmt.Sprintf("hhea table: numHMetrics: %d\n", f.hhea.numberOfHMetrics))
	} else {
		b.WriteString("hhea: missing\n")
	}

	if f.hmtx != nil {
		b.WriteString(fmt.Sprintf("hmtxtable:  hmetrics: %d, leftSideBearings: %d\n",
			len(f.hmtx.hMetrics), len(f.hmtx.leftSideBearings)))
	} else {
		b.WriteString("hmtx: missing\n")
	}

	if f.cmap != nil {
		b.WriteString(fmt.Sprintf("cmap version: %d\n",
			f.cmap.version))
		b.WriteString(fmt.Sprintf("cmap: encoding records: %d subtables: %d\n",
			len(f.cmap.encodingRecords), len(f.cmap.subtables)))
		b.WriteString(fmt.Sprintf("cmap: subtables: %+v\n", f.cmap.subtableKeys))
		for _, k := range f.cmap.subtableKeys {
			subt := f.cmap.subtables[k]
			b.WriteString(fmt.Sprintf("cmap subtable: %s: runes: %d\n", k, len(subt.runes)))
		}
	}
	if f.loca != nil {
		b.WriteString(fmt.Sprintf("Loca table\n"))
		b.WriteString(fmt.Sprintf("- Short offsets: %d\n", len(f.loca.offsetsShort)))
		b.WriteString(fmt.Sprintf("- Long offsets: %d\n", len(f.loca.offsetsLong)))
	}

	if f.glyf != nil {
		rawTotal := 0.0
		for _, desc := range f.glyf.descs {
			rawTotal += float64(len(desc.raw))
		}
		b.WriteString(fmt.Sprintf("glyf table present: %d descriptions (%.2f kB)\n", len(f.glyf.descs), rawTotal/1024))
	} else {
		b.WriteString("glyf: missing\n")
	}

	if f.post != nil {
		b.WriteString(fmt.Sprintf("post table present: %d numGlyphs\n", f.post.numGlyphs))
		b.WriteString(fmt.Sprintf("- post glyphNameIndex: %d\n", len(f.post.glyphNameIndex)))
		b.WriteString(fmt.Sprintf("- post glyphNames: %d\n", len(f.post.glyphNames)))
		for i, gn := range f.post.glyphNames {
			if i > 10 {
				break
			}
			b.WriteString(fmt.Sprintf("- post: %d: %s\n", i+1, gn))
		}
	} else {
		b.WriteString("post: missing\n")
	}

	return b.String()
}

// Encode encodes `str` into a byte array.
// TODO(gunnsth): rune -> charcode (need to have a representation of the char code).
//      Charcodes are what is stored in the PDF content stream and also in the TTF directly.
func (f *font) Encode(str string) []byte {
	if f.cmap == nil {
		logrus.Debug("ERROR: No cmap loaded - returning back")
		return []byte(str)
	}

	var buf bytes.Buffer

	// TODO: Need rune -> charcode and charcode -> bytes
	/*
		for _, r := range str {
			for _, subt := range f.cmap.subtables {
				bb, has := subt.cmap[r].codeBytes[r]
				if has {
					buf.Write(bb)
				}
			}
		}
	*/

	return buf.Bytes()
}

// Decode decodes a `raw` byte array into a string.
//func (f *font) Decode(raw []byte) string {
//	return ""
//}

// RuneToCharcode returns the charcode corrsponding to rune `r`.
//func (f *font) RuneToCharcode(r rune) (CharCode, bool) {
//	return 0, false
//}

// ToPdfObject returns a PDF representation of the truetype encoder.
//func (f *font) ToPdfObject() core.PdfObject {
//	return nil
//}
