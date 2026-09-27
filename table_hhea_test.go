/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"testing"
)

// TestParseHhea_RejectsShortTable asserts an hhea record shorter than 36
// bytes fails rather than reading the bytes after it.
func TestParseHhea_RejectsShortTable(t *testing.T) {
	data := append(make([]byte, 10), bytes.Repeat([]byte{0xFF}, 20)...)
	f := &font{
		trec: &tableRecords{
			trMap: map[string]*tableRecord{
				"hhea": {offset: 0, length: 10},
			},
		},
	}
	_, err := f.parseHhea(newByteReader(bytes.NewReader(data)))
	assertTruncatedRead(t, err)
}
