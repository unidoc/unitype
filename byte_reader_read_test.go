/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"testing"
	"testing/iotest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestByteReaderTypedReads asserts each typed read decodes the same value, and
// returns the same error on short input, as binary.Read, and that a short
// read consumes the bytes present.
func TestByteReaderTypedReads(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for n := 0; n < 200; n++ {
		data := make([]byte, 8)
		rng.Read(data)
		for size := 0; size <= 8; size++ {
			in := data[:size]
			check := func(name string, want interface{}, got func(*byteReader) (interface{}, error)) {
				t.Helper()
				wantErr := binary.Read(bytes.NewReader(in), binary.BigEndian, want)
				for _, rs := range []io.ReadSeeker{bytes.NewReader(in), oneByteReadSeeker(in)} {
					r := newByteReader(rs)
					val, err := got(r)
					assert.Equal(t, wantErr, err, "%s on %d bytes", name, size)
					if err != nil {
						assert.Equal(t, int64(size), r.Offset(), "%s: a short read consumes the bytes present", name)
					}
					if wantErr == nil {
						assert.Equal(t, derefAny(want), val, "%s on % x", name, in)
					}
				}
			}
			check("uint8", new(uint8), func(r *byteReader) (interface{}, error) { return r.readUint8() })
			check("int8", new(int8), func(r *byteReader) (interface{}, error) { return r.readInt8() })
			check("uint16", new(uint16), func(r *byteReader) (interface{}, error) { return r.readUint16() })
			check("int16", new(int16), func(r *byteReader) (interface{}, error) { return r.readInt16() })
			check("uint32", new(uint32), func(r *byteReader) (interface{}, error) { return r.readUint32() })
			check("int32", new(int32), func(r *byteReader) (interface{}, error) { return r.readInt32() })
			check("fixed", new(fixed), func(r *byteReader) (interface{}, error) { return r.readFixed() })
			check("fword", new(fword), func(r *byteReader) (interface{}, error) { return r.readFword() })
			check("ufword", new(ufword), func(r *byteReader) (interface{}, error) { return r.readUfword() })
			check("f2dot14", new(f2dot14), func(r *byteReader) (interface{}, error) { return r.readF2dot14() })
			check("offset16", new(offset16), func(r *byteReader) (interface{}, error) { return r.readOffset16() })
			check("offset32", new(offset32), func(r *byteReader) (interface{}, error) { return r.readOffset32() })
			check("longdatetime", new(longdatetime), func(r *byteReader) (interface{}, error) { return r.readLongdatetime() })
			check("tag", new(tag), func(r *byteReader) (interface{}, error) { return r.readTag() })
		}
	}
}

// oneByteReadSeeker returns a ReadSeeker over b whose Read returns at most one
// byte per call, so reads span several fills of the read buffer.
func oneByteReadSeeker(b []byte) io.ReadSeeker {
	br := bytes.NewReader(b)
	return struct {
		io.Reader
		io.Seeker
	}{iotest.OneByteReader(br), br}
}

// derefAny returns the value p points to.
func derefAny(p interface{}) interface{} {
	switch v := p.(type) {
	case *uint8:
		return *v
	case *int8:
		return *v
	case *uint16:
		return *v
	case *int16:
		return *v
	case *uint32:
		return *v
	case *int32:
		return *v
	case *fixed:
		return *v
	case *fword:
		return *v
	case *ufword:
		return *v
	case *f2dot14:
		return *v
	case *offset16:
		return *v
	case *offset32:
		return *v
	case *longdatetime:
		return *v
	case *tag:
		return *v
	}
	return nil
}

// TestByteReaderSeekToDiscardsBuffer asserts a read after SeekTo returns the
// bytes at the new offset, not data buffered before the seek.
func TestByteReaderSeekToDiscardsBuffer(t *testing.T) {
	data := make([]byte, 8192)
	for i := range data {
		data[i] = byte(i / 2)
	}
	r := newByteReader(bytes.NewReader(data))
	_, err := r.readUint16() // fills the read buffer from offset 0
	require.NoError(t, err)

	require.NoError(t, r.SeekTo(6000))
	v, err := r.readUint16()
	require.NoError(t, err)
	assert.Equal(t, binary.BigEndian.Uint16(data[6000:]), v)
	assert.Equal(t, int64(6002), r.Offset())
}

// failingReader returns data, then fails every Read with err.
type failingReader struct {
	*bytes.Reader
	err error
}

func (r failingReader) Read(p []byte) (int, error) {
	if r.Len() == 0 {
		return 0, r.err
	}
	return r.Reader.Read(p)
}

// TestByteReaderReadError asserts a typed read passes up a read error other
// than EOF, with or without bytes before it.
func TestByteReaderReadError(t *testing.T) {
	readErr := errors.New("read failed")
	for _, present := range [][]byte{nil, {0x12}} {
		r := newByteReader(failingReader{bytes.NewReader(present), readErr})
		_, err := r.readUint32()
		assert.ErrorIs(t, err, readErr, "with %d bytes present", len(present))
	}
}
