/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestParseHead_RejectsShortTable: a head table record shorter than its
// fixed required length (54 bytes) must be rejected before the magic-number
// check ever gets a chance to run (never mind further fields) - reading past
// its own declared length into whatever bytes follow it in the file could
// coincidentally happen to contain the magic number too. White-box (package
// unitype) since no bundled font has a short head table.
func TestParseHead_RejectsShortTable(t *testing.T) {
	// magicNumber (0x5F0F3CF5) placed at its real offset (12) in the
	// trailing "next table" bytes, so a version that skipped the length
	// check would pass the magic-number check too and mask this test.
	data := make([]byte, 12)
	data = append(data, 0x5F, 0x0F, 0x3C, 0xF5)
	data = append(data, bytes.Repeat([]byte{0xFF}, 20)...)
	f := &font{
		trec: &tableRecords{
			trMap: map[string]*tableRecord{
				"head": {offset: 0, length: 16},
			},
		},
	}
	_, err := f.parseHead(newByteReader(bytes.NewReader(data)))
	assert.ErrorIs(t, err, errRangeCheck)
}
