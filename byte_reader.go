/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bufio"
	"io"

	"github.com/sirupsen/logrus"
)

// byteReader encapsulates io.ReadSeeker with buffering and provides methods to read binary data as
// needed for truetype fonts.  The buffered reader is used to enhance the performance when reading
// binary data types one at a time.
type byteReader struct {
	rs     io.ReadSeeker
	reader *bufio.Reader
}

func newByteReader(rs io.ReadSeeker) *byteReader {
	return &byteReader{
		rs:     rs,
		reader: bufio.NewReader(rs),
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

// Skip skips over `n` bytes.
func (r *byteReader) Skip(n int) error {
	_, err := r.reader.Discard(n)
	return err
}

// readBytes reads bytes straight from `r`.
func (r *byteReader) readBytes(bp *[]byte, length int) error {
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

// readBE reads an n-byte big-endian unsigned integer (n <= 8) from the read
// buffer without allocating. Like binary.Read, it returns io.EOF if no bytes
// remain and io.ErrUnexpectedEOF if fewer than n do.
func (r byteReader) readBE(n int) (uint64, error) {
	b, err := r.reader.Peek(n)
	if len(b) < n {
		if err == io.EOF && len(b) > 0 {
			err = io.ErrUnexpectedEOF
		}
		_, _ = r.reader.Discard(len(b))
		return 0, err
	}
	var v uint64
	for _, c := range b {
		v = v<<8 | uint64(c)
	}
	_, err = r.reader.Discard(n)
	return v, err
}

func (r byteReader) readF2dot14() (f2dot14, error) {
	v, err := r.readBE(2)
	return f2dot14(v), err
}

func (r byteReader) readFixed() (fixed, error) {
	v, err := r.readBE(4)
	return fixed(v), err
}

func (r byteReader) readFword() (fword, error) {
	v, err := r.readBE(2)
	return fword(v), err
}

func (r byteReader) readUint8() (uint8, error) {
	v, err := r.readBE(1)
	return uint8(v), err
}

func (r byteReader) readUint16() (uint16, error) {
	v, err := r.readBE(2)
	return uint16(v), err
}

func (r byteReader) readInt8() (int8, error) {
	v, err := r.readBE(1)
	return int8(v), err
}

func (r byteReader) readInt16() (int16, error) {
	v, err := r.readBE(2)
	return int16(v), err
}

func (r byteReader) readInt32() (int32, error) {
	v, err := r.readBE(4)
	return int32(v), err
}

func (r byteReader) readUint32() (uint32, error) {
	v, err := r.readBE(4)
	return uint32(v), err
}

func (r byteReader) readTag() (tag, error) {
	v, err := r.readBE(4)
	return tag{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}, err
}

func (r byteReader) readUfword() (ufword, error) {
	v, err := r.readBE(2)
	return ufword(v), err
}

func (r byteReader) readLongdatetime() (longdatetime, error) {
	v, err := r.readBE(8)
	return longdatetime(v), err
}

func (r byteReader) readOffset16() (offset16, error) {
	v, err := r.readBE(2)
	return offset16(v), err
}

func (r byteReader) readOffset32() (offset32, error) {
	v, err := r.readBE(4)
	return offset32(v), err
}
