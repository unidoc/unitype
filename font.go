/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/unidoc/unipdf/v3/common"
)

// TODO: Export only what unipdf needs:
// Encoding: rune <-> GID map.
// font flags:
//		IsFixedPitch, Serif, etc (Table 123 PDF32000_2008 - font flags)
//		FixedPitch() bool
//		Serif() bool
//		Symbolic() bool
//		Script() bool
//		Nonsymbolic() bool
//		Italic() bool
//		AllCap() bool
//		SmallCap() bool
//		ForceBold() bool
//      Need to be able to derive the font flags from the font to build a font descriptor
//
// Required table according to PDF32000_2008 (9.9 Embedded font programs - p. 299):
// “head”, “hhea”, “loca”, “maxp”, “cvt”, “prep”, “glyf”, “hmtx”, and “fpgm”. If used with a simple
// font dictionary, the font program shall additionally contain a cmap table defining one or more
// encodings, as discussed in 9.6.6.4, "Encodings for TrueType Fonts". If used with a CIDFont
// dictionary, the cmap table is not needed and shall not be present, since the mapping from
// character codes to glyph descriptions is provided separately.
//

// font is a data model for truetype fonts with basic access methods.
type font struct {
	strict            bool
	incompatibilities []string

	ot   *offsetTable
	trec *tableRecords // table records (references other tables).
	head *headTable
	maxp *maxpTable
	hhea *hheaTable
	hmtx *hmtxTable
	loca *locaTable
	glyf *glyfTable
	name *nameTable
	os2  *os2Table
	post *postTable
	cmap *cmapTable
}

// Returns an error in strict mode, otherwise adds the incompatibility to a list of noted incompatibilities.
func (f *font) recordIncompatibilityf(fmtstr string, a ...interface{}) error {
	str := fmt.Sprintf(fmtstr, a...)
	if f.strict {
		return fmt.Errorf("incompatibility: %s", str)
	}
	f.incompatibilities = append(f.incompatibilities, str)
	return nil
}

func (f font) numTables() int {
	return int(f.ot.numTables)
}

func parseFont(r *byteReader) (*font, error) {
	f := &font{}

	var err error

	// Load table offsets and records.
	f.ot, err = f.parseOffsetTable(r)
	if err != nil {
		return nil, err
	}

	f.trec, err = f.parseTableRecords(r)
	if err != nil {
		return nil, err
	}

	// TODO: Avoid parsing tables unless needed?  Like have a f.GetHead that returns the head table if it is
	//   not already loaded. Guarantees that we only lo
	// Or at least avoid the biggest tables that have (optional) information - not used by most frequent use cases.

	f.head, err = f.parseHead(r)
	if err != nil {
		return nil, err
	}

	f.maxp, err = f.parseMaxp(r)
	if err != nil {
		return nil, err
	}

	f.hhea, err = f.parseHhea(r)
	if err != nil {
		return nil, err
	}

	f.hmtx, err = f.parseHmtx(r)
	if err != nil {
		return nil, err
	}

	f.loca, err = f.parseLoca(r)
	if err != nil {
		return nil, err
	}

	f.glyf, err = f.parseGlyf(r)
	if err != nil {
		return nil, err
	}

	f.name, err = f.parseNameTable(r)
	if err != nil {
		return nil, err
	}

	f.os2, err = f.parseOS2Table(r)
	if err != nil {
		return nil, err
	}

	f.post, err = f.parsePost(r)
	if err != nil {
		return nil, err
	}

	f.cmap, err = f.parseCmap(r)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// numTablesToWrite returns the number of tables in `f`.
// Calculates based on the number of tables will be written out.
// NOTE that not all tables that are loaded are written out.
func (f *font) numTablesToWrite() int {
	var num int

	if f.head != nil {
		num++
	}
	if f.maxp != nil {
		num++
	}
	if f.hhea != nil {
		num++
	}
	if f.hmtx != nil {
		num++
	}
	if f.loca != nil {
		num++
	}
	if f.glyf != nil {
		num++
	}
	if f.name != nil {
		num++
	}
	if f.os2 != nil {
		num++
	}
	if f.post != nil {
		num++
	}
	if f.cmap != nil {
		num++
	}
	return num
}

func (f *font) write(w *byteWriter) error {
	// TODO(gunnsth): Can be memory intensive on large fonts, any way to improve?
	//     Another option to write to temp files and combine.
	//     Or a combined fixed size buffer with partial file system use.
	//     Best if such implementation is hidden within a well tested package.

	common.Log.Debug("Write 1")
	numTables := f.numTablesToWrite()
	otTable := &offsetTable{
		sfntVersion:   f.ot.sfntVersion,
		numTables:     uint16(numTables),
		searchRange:   f.ot.searchRange,
		entrySelector: f.ot.entrySelector,
		rangeShift:    f.ot.rangeShift,
	}
	trec := &tableRecords{}

	f.ot.numTables = uint16(numTables)

	// Starting offset after offset table and table records.
	startOffset := int64(12 + numTables*16)

	fmt.Printf("==== write\nnumTables: %d\nstartOffset: %d\n", numTables, startOffset)

	common.Log.Debug("Write 2")
	// Writing is two phases and is done in a few steps:
	// 1. Write the content tables: head, hhea, etc in the expected order and keep track of the length, checksum for each.
	// 2. Generate the table records based on the information.
	// 3. Write out in final order: offset table, table records, head, ...
	// 4. Set checkAdjustment of head table based on checksumof entire file
	// 5. Write the final output

	// Write to buffer to get offsets.
	var buf bytes.Buffer
	var headChecksum uint32
	{
		bufw := newByteWriter(&buf)

		// head.
		f.head.checksumAdjustment = 0
		offset := startOffset
		err := f.writeHead(bufw)
		if err != nil {
			return err
		}
		headChecksum = bufw.checksum()
		trec.Set("head", offset, bufw.bufferedLen(), headChecksum)
		err = bufw.flush()
		if err != nil {
			return err
		}

		// maxp.
		offset = startOffset + bufw.flushedLen
		err = f.writeMaxp(bufw)
		if err != nil {
			return err
		}
		trec.Set("maxp", offset, bufw.bufferedLen(), bufw.checksum())
		err = bufw.flush()
		if err != nil {
			return err
		}

		// hhea.
		if f.hhea != nil {
			offset = startOffset + bufw.flushedLen
			err = f.writeHhea(bufw)
			if err != nil {
				return err
			}
			trec.Set("hhea", offset, bufw.bufferedLen(), bufw.checksum())
			err = bufw.flush()
			if err != nil {
				return err
			}
		}

		// hmtx.
		if f.hmtx != nil {
			offset = startOffset + bufw.flushedLen
			err = f.writeHmtx(bufw)
			if err != nil {
				return err
			}
			trec.Set("hmtx", offset, bufw.bufferedLen(), bufw.checksum())
			err = bufw.flush()
			if err != nil {
				return err
			}
		}

		// loca.
		if f.loca != nil {
			offset = startOffset + bufw.flushedLen
			err = f.writeLoca(bufw)
			if err != nil {
				return err
			}
			trec.Set("loca", offset, bufw.bufferedLen(), bufw.checksum())
			err = bufw.flush()
			if err != nil {
				return err
			}
		}

		// glyf.
		if f.glyf != nil {
			offset = startOffset + bufw.flushedLen
			err = f.writeGlyf(bufw)
			if err != nil {
				return err
			}
			trec.Set("glyf", offset, bufw.bufferedLen(), bufw.checksum())
			err = bufw.flush()
			if err != nil {
				return err
			}
		}

		// name.
		if f.name != nil {
			offset = startOffset + bufw.flushedLen
			err = f.writeNameTable(bufw)
			if err != nil {
				return err
			}
			trec.Set("name", offset, bufw.bufferedLen(), bufw.checksum())
			err = bufw.flush()
			if err != nil {
				return err
			}
		}

		// os2.
		if f.os2 != nil {
			offset = startOffset + bufw.flushedLen
			err = f.writeOS2(bufw)
			if err != nil {
				return err
			}
			trec.Set("OS/2", offset, bufw.bufferedLen(), bufw.checksum())
			err = bufw.flush()
			if err != nil {
				return err
			}
		}

		// post
		if f.post != nil {
			offset = startOffset + bufw.flushedLen
			err = f.writePost(bufw)
			if err != nil {
				return err
			}
			trec.Set("post", offset, bufw.bufferedLen(), bufw.checksum())
			err = bufw.flush()
			if err != nil {
				return err
			}
		}

		// cmap
		if f.cmap != nil {
			offset = startOffset + bufw.flushedLen
			err = f.writeCmap(bufw)
			if err != nil {
				return err
			}
			trec.Set("cmap", offset, bufw.bufferedLen(), bufw.checksum())
			err = bufw.flush()
			if err != nil {
				return err
			}
		}
	}
	common.Log.Debug("Write 3")

	// Write the offset and table records to another mock buffer.
	var bufh bytes.Buffer
	{
		bufw := newByteWriter(&bufh)
		// Create a mock font for writing without modifying the original entries of `f`.
		mockf := &font{
			ot:   otTable,
			trec: trec,
		}

		err := mockf.writeOffsetTable(bufw)
		if err != nil {
			return err
		}

		err = mockf.writeTableRecords(bufw)
		if err != nil {
			return err
		}
		err = bufw.flush()
		if err != nil {
			return err
		}
	}

	// Write everything to bufh.
	_, err := buf.WriteTo(&bufh)
	if err != nil {
		return err
	}

	// Calculate total checksum for the entire font.
	checksummer := byteWriter{
		buffer: bufh,
	}
	fontChecksum := checksummer.checksum()
	checksumAdjustment := 0xB1B0AFBA - fontChecksum

	// Set the checksumAdjustment of the head table.
	data := bufh.Bytes()
	hoff := startOffset
	binary.BigEndian.PutUint32(data[hoff+8:hoff+12], checksumAdjustment)

	buffer := bytes.NewBuffer(data)
	_, err = io.Copy(&w.buffer, buffer)
	return err
}
