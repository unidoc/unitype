/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/sirupsen/logrus"
)

// byteReader encapsulates io.ReadSeeker with buffering and provides methods to read binary data as
// needed for truetype fonts.  The buffered reader is used to enhance the performance when reading
// binary data types one at a time.
type byteReader struct {
	rs     io.ReadSeeker
	reader *bufio.Reader

	// size is the stream's size, recorded by remaining the first time it's
	// needed (sizeKnown).
	size      int64
	sizeKnown bool
}

func newByteReader(rs io.ReadSeeker) *byteReader {
	return &byteReader{
		rs:     rs,
		reader: bufio.NewReader(rs),
	}
}

// minReadBufferSize is the smallest read buffer bufio allows.
const minReadBufferSize = 16

// newBytesReader returns a byteReader over the in-memory table b, with the
// smallest read buffer rather than bufio's 4 KB default or a second copy of b.
func newBytesReader(b []byte) *byteReader {
	rs := bytes.NewReader(b)
	return &byteReader{
		rs:     rs,
		reader: bufio.NewReaderSize(rs, minReadBufferSize),
	}
}

// Offset returns current offset position of `r`.
func (r byteReader) Offset() int64 {
	offset, _ := r.rs.Seek(0, io.SeekCurrent)
	offset -= int64(r.reader.Buffered())
	return offset
}

// SeekTo seeks to offset, discarding any buffered data but reusing the
// read buffer.
func (r *byteReader) SeekTo(offset int64) error {
	_, err := r.rs.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}
	r.reader.Reset(r.rs)
	return nil
}

// remaining returns the number of bytes left in the stream after the current
// offset of `r`.
func (r *byteReader) remaining() (int64, error) {
	pos, err := r.rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	if !r.sizeKnown {
		end, err := r.rs.Seek(0, io.SeekEnd)
		if err != nil {
			return 0, err
		}
		if _, err := r.rs.Seek(pos, io.SeekStart); err != nil {
			return 0, err
		}
		r.size, r.sizeKnown = end, true
	}
	return r.size - pos + int64(r.reader.Buffered()), nil
}

// Skip skips over `n` bytes.
func (r *byteReader) Skip(n int) error {
	_, err := r.reader.Discard(n)
	return err
}

// maxUncheckedReadLen is the largest length readBytes allocates without
// first checking that the stream holds that many more bytes.
const maxUncheckedReadLen = 64 * 1024

// readBytes reads bytes straight from `r`. A negative length is an error, and
// a length over maxUncheckedReadLen is checked against the bytes left in the
// stream before allocating.
func (r *byteReader) readBytes(bp *[]byte, length int) error {
	if length < 0 {
		return errRangeCheck
	}
	if length > maxUncheckedReadLen {
		remaining, err := r.remaining()
		if err != nil {
			return err
		}
		if int64(length) > remaining {
			return io.ErrUnexpectedEOF
		}
	}
	*bp = make([]byte, length)
	_, err := io.ReadFull(r.reader, *bp)
	if err != nil {
		return err
	}

	return nil
}

// readSlice reads a series of values into `slice` from `r` (big endian).
func (r *byteReader) readSlice(slice interface{}, length int) error {
	switch t := slice.(type) {
	case *[]uint8:
		for i := 0; i < length; i++ {
			val, err := r.readUint8()
			if err != nil {
				return err
			}
			*t = append(*t, val)
		}
	case *[]uint16:
		for i := 0; i < length; i++ {
			val, err := r.readUint16()
			if err != nil {
				return err
			}
			*t = append(*t, val)
		}
	case *[]int16:
		for i := 0; i < length; i++ {
			val, err := r.readInt16()
			if err != nil {
				return err
			}
			*t = append(*t, val)
		}
	case *[]offset16:
		for i := 0; i < length; i++ {
			val, err := r.readOffset16()
			if err != nil {
				return err
			}
			*t = append(*t, val)
		}
	case *[]offset32:
		for i := 0; i < length; i++ {
			val, err := r.readOffset32()
			if err != nil {
				return err
			}
			*t = append(*t, val)
		}

	default:
		logrus.Errorf("Unsupported type: %T (readSlice)", t)
		return errTypeCheck
	}
	return nil
}

// read reads a series of fields from `r`.
func (r byteReader) read(fields ...interface{}) error {
	for _, f := range fields {
		switch t := f.(type) {
		case **f2dot14:
			val, err := r.readF2dot14()
			if err != nil {
				return err
			}
			*t = &val
		case *f2dot14:
			val, err := r.readF2dot14()
			if err != nil {
				return err
			}
			*t = val
		case *fixed:
			val, err := r.readFixed()
			if err != nil {
				return err
			}
			*t = val
		case *fword:
			val, err := r.readFword()
			if err != nil {
				return err
			}
			*t = val
		case *int8:
			val, err := r.readInt8()
			if err != nil {
				return err
			}
			*t = val
		case *int16:
			val, err := r.readInt16()
			if err != nil {
				return err
			}
			*t = val
		case *int32:
			val, err := r.readInt32()
			if err != nil {
				return err
			}
			*t = val
		case *longdatetime:
			val, err := r.readLongdatetime()
			if err != nil {
				return err
			}
			*t = val
		case *offset16:
			val, err := r.readOffset16()
			if err != nil {
				return err
			}
			*t = val
		case *offset32:
			val, err := r.readOffset32()
			if err != nil {
				return err
			}
			*t = val
		case *ufword:
			val, err := r.readUfword()
			if err != nil {
				return err
			}
			*t = val
		case *uint8:
			val, err := r.readUint8()
			if err != nil {
				return err
			}
			*t = val
		case *uint16:
			val, err := r.readUint16()
			if err != nil {
				return err
			}
			*t = val
		case *tag:
			val, err := r.readTag()
			if err != nil {
				return err
			}
			*t = val
		case *uint32:
			val, err := r.readUint32()
			if err != nil {
				return err
			}
			*t = val

		default:
			logrus.Errorf("Unsupported type: %T (read)", t)
			return errTypeCheck
		}
	}
	return nil
}

func (r byteReader) readF2dot14() (f2dot14, error) {
	b := make([]byte, 2)
	_, err := io.ReadFull(r.reader, b)
	if err != nil {
		return 0, err
	}
	u16 := binary.BigEndian.Uint16(b)
	return f2dot14(u16), nil
}

func (r byteReader) readFixed() (fixed, error) {
	var val fixed
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readFword() (fword, error) {
	var val fword
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readUint8() (uint8, error) {
	var val uint8
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readUint16() (uint16, error) {
	var val uint16
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readInt8() (int8, error) {
	var val int8
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readInt16() (int16, error) {
	var val int16
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readInt32() (int32, error) {
	var val int32
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readUint32() (uint32, error) {
	var val uint32
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readTag() (tag, error) {
	var val tag
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readUfword() (ufword, error) {
	var val ufword
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readLongdatetime() (longdatetime, error) {
	var val longdatetime
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readOffset16() (offset16, error) {
	var val offset16
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}

func (r byteReader) readOffset32() (offset32, error) {
	var val offset32
	err := binary.Read(r.reader, binary.BigEndian, &val)
	return val, err
}
