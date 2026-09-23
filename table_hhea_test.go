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

// TestParseHhea_RejectsShortTable: an hhea table record shorter than its
// fixed required length (36 bytes) must be rejected rather than read past
// its own declared length into whatever bytes follow it in the file - a
// garbage numberOfHMetrics from doing so would propagate straight into
// parseHmtx's own length check. White-box (package unitype) since no
// bundled font has a short hhea table.
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
	assert.ErrorIs(t, err, errRangeCheck)
}
