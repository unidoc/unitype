/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

// assertTruncatedRead asserts err is the EOF a parser gets from reading past
// the end of its bounded table bytes.
func assertTruncatedRead(t *testing.T, err error) {
	t.Helper()
	assert.True(t, errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF), "want EOF-class error, got %v", err)
}

// TestParseHead_RejectsShortTable asserts a head record shorter than 54 bytes
// fails, even when the bytes after it hold the magic number.
func TestParseHead_RejectsShortTable(t *testing.T) {
	// magicNumber at its real offset (12), just past the declared length.
	data := make([]byte, 12)
	data = append(data, 0x5F, 0x0F, 0x3C, 0xF5)
	data = append(data, bytes.Repeat([]byte{0xFF}, 64)...) // enough for a full head past the magic
	f := &font{
		trec: &tableRecords{
			trMap: map[string]*tableRecord{
				"head": {offset: 0, length: 12},
			},
		},
	}
	_, err := f.parseHead(newByteReader(bytes.NewReader(data)))
	assertTruncatedRead(t, err)
}
