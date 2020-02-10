/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package truetype

import (
	"bytes"
	"testing"
)

/*
TODO: There seems to be little point in trying to make each table independent, there are dependencies between
      the tables and makes more sense to load together with full access to the font.
              font.parseGlyf(r *byteReader) (*glyfTable, error)
              font.writeGlyf(w *byteWriter) error

*/

func TestGlyfUnmarshalBasic(t *testing.T) {
	/*
		Mock case.
		Number of glyphs: 3.
		Number of contours: 1, 1, 3
		Points per contour: 2
	*/
	// TODO (gunnsth): Unclear what is defined and what is for context.
	//      Seems easier if unmarshaller can have access to everything or scope off context.
	//      In global context, typically to font.parseGlyf() where font has access to everything.
	//      Could also make the font the context parameter...
	//      - Embed *font into tables to provide the context while Read/Writing.

	// Generate mock font for testing glyf table.
	makeFont := func() *font {
		head := headTable{
			indexToLocFormat: 1,
		}
		maxp := maxpTable{
			numGlyphs: 1,
		}
		loca := locaTable{
			offsetsShort: []offset16{0},
		}
		glyf := glyfTable{
			descs: []*glyphDescription{
				{
					header: glyfGlyphHeader{
						numberOfContours: 1,
					},
					simple: &simpleGlyphDescription{
						endPtsOfContours: []uint16{1},
						flags:            []uint8{uint8(xShortVector | repeatFlag), uint8(xShortVector | repeatFlag)}, // unpacked.
						xCoordinates:     []uint16{10, 20},
						yCoordinates:     []uint16{30, 50},
					},
				},
			},
		}

		return &font{
			head: &head,
			maxp: &maxp,
			loca: &loca,
			glyf: &glyf,
		}
	}

	f := makeFont()
	var buf bytes.Buffer
	bw := newByteWriter(&buf)
	bw.flush()
	t.Logf("@0 checksum: %d", bw.checksum())
	t.Logf("@0 Cur bufLen: %d", bw.bufferedLen())
	f.writeGlyf(bw)
	/*(
	checksum := bw.checksum()

	fmt.Printf("checksum: %d\n", checksum)
	fmt.Printf("Cur bufLen: %d\n", bw.bufferedLen())
	fmt.Printf("@0 %d\n", buf.Len())
	bw.flush()
	fmt.Printf("@1 %d\n", buf.Len())
	*/

	_ = f

	/*
		// starts with glyph data
		//expected := []byte{}

		var buf bytes.Buffer
		bw := newByteWriter(&buf)

		err := f.writeGlyf(bw)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		b := buf.Bytes()
		fmt.Printf("b(%d): % X\n", len(b), b)

		// pad with a single byte.
		b = append(b, 0, 0, 0, 0)

		// TODO: Check marshalled output and see if it makes sense.
		br := newByteReader(bytes.NewReader(b))
		//font2, err := parseFont(br)
		glyf2, err := f.ParseGlyf(br)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		fmt.Printf("%+v\n", glyf2)
	*/
}
