/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestReadBytes_LargeLength asserts readBytes accepts a length over
// maxUncheckedReadLen that fits in the stream, counting bytes already
// buffered, and rejects one byte more.
func TestReadBytes_LargeLength(t *testing.T) {
	const size = 3 * maxUncheckedReadLen
	newReader := func() *byteReader {
		r := newByteReader(bytes.NewReader(make([]byte, size)))
		var head uint16
		require.NoError(t, r.read(&head)) // leaves read-ahead in the buffer
		return r
	}

	var b []byte
	require.NoError(t, newReader().readBytes(&b, size-2))
	assert.Len(t, b, size-2)

	assert.ErrorIs(t, newReader().readBytes(&b, size-1), io.ErrUnexpectedEOF)
	assert.ErrorIs(t, newReader().readBytes(&b, -1), errRangeCheck)
}
