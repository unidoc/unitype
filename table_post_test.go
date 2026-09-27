/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParsePost_LongGlyphName asserts a version 2.0 glyph name longer than
// 127 bytes parses, since its Pascal-string length byte is a uint8.
func TestParsePost_LongGlyphName(t *testing.T) {
	data := make([]byte, 32)                    // version .. maxMemType1
	data[1] = 0x02                              // version 2.0 (0x00020000)
	data = append(data, 0x00, 0x01, 0x01, 0x02) // numGlyphs = 1, glyphNameIndex[0] = 258
	data = append(data, 200)
	data = append(data, bytes.Repeat([]byte{'a'}, 200)...)

	f := &font{
		maxp: &maxpTable{numGlyphs: 1},
		trec: &tableRecords{trMap: map[string]*tableRecord{"post": {offset: 0, length: uint32(len(data))}}},
	}
	post, err := f.parsePost(newByteReader(bytes.NewReader(data)))
	require.NoError(t, err)
	require.Len(t, post.glyphNames, 1)
	assert.Len(t, string(post.glyphNames[0]), 200)
}
