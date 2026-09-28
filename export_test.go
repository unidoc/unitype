/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// readTestFont returns the bytes of the font at `path`.
func readTestFont(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	return data
}

// tableRecordPos returns the position in `data` of the table record for
// `table`, failing the test if the font has no such record.
func tableRecordPos(t *testing.T, data []byte, table string) int {
	t.Helper()
	numTables := int(binary.BigEndian.Uint16(data[4:6]))
	for i := 0; i < numTables; i++ {
		pos := 12 + 16*i
		if string(data[pos:pos+4]) == table {
			return pos
		}
	}
	t.Fatalf("no %q table record", table)
	return 0
}

// renameTable renames the table record `from` to `to` in `data`, so parsers
// looking for `from` find it absent.
func renameTable(t *testing.T, data []byte, from, to string) {
	t.Helper()
	pos := tableRecordPos(t, data, from)
	copy(data[pos:pos+4], to)
}

// TestSubsetKeepIndices_NoGlyf asserts a font with a loca table but no glyf
// table subsets without dereferencing the missing glyf table.
func TestSubsetKeepIndices_NoGlyf(t *testing.T) {
	data := readTestFont(t, "testdata/FreeSans.ttf")
	renameTable(t, data, "glyf", "xxxx")
	fnt, err := Parse(bytes.NewReader(data))
	require.NoError(t, err)
	require.Nil(t, fnt.glyf)

	sub, err := fnt.SubsetKeepIndices([]GlyphIndex{0, 5})
	require.NoError(t, err)
	assert.Equal(t, 6, int(sub.maxp.numGlyphs))

	sub, err = fnt.SubsetKeepRunes([]rune("abc"))
	require.NoError(t, err)
	assert.NotNil(t, sub)
}
