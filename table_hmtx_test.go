package unitype

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptimizeHmtxTable(t *testing.T) {
	testcases := []struct {
		fnt            *font
		expNumGlyphs   int
		expNumHMetrics int
		expLSB         []int16
		exphMetrics    []longHorMetric
	}{
		{
			fnt: &font{
				maxp: &maxpTable{
					numGlyphs: 5,
				},
				hhea: &hheaTable{
					numberOfHMetrics: 3,
				},
				hmtx: &hmtxTable{
					hMetrics: []longHorMetric{
						{advanceWidth: 10, lsb: 1},
						{advanceWidth: 20, lsb: 2},
						{advanceWidth: 30, lsb: 3},
					},
					leftSideBearings: []int16{
						4,
						5,
					},
				},
			},
			expNumGlyphs:   5,
			expNumHMetrics: 3,
			exphMetrics: []longHorMetric{
				{advanceWidth: 10, lsb: 1},
				{advanceWidth: 20, lsb: 2},
				{advanceWidth: 30, lsb: 3},
			},
			expLSB: []int16{4, 5},
		},
		{
			fnt: &font{
				maxp: &maxpTable{
					numGlyphs: 7,
				},
				hhea: &hheaTable{
					numberOfHMetrics: 7,
				},
				hmtx: &hmtxTable{
					hMetrics: []longHorMetric{
						{advanceWidth: 10, lsb: 1},
						{advanceWidth: 20, lsb: 2},
						{advanceWidth: 30, lsb: 3},
						{advanceWidth: 40, lsb: 4},
						{advanceWidth: 50, lsb: 5},
						{advanceWidth: 60, lsb: 6}, // should include this once optimized.
						{advanceWidth: 60, lsb: 7},
					},
					leftSideBearings: []int16{},
				},
			},
			expNumGlyphs:   7,
			expNumHMetrics: 6,
			exphMetrics: []longHorMetric{
				{advanceWidth: 10, lsb: 1},
				{advanceWidth: 20, lsb: 2},
				{advanceWidth: 30, lsb: 3},
				{advanceWidth: 40, lsb: 4},
				{advanceWidth: 50, lsb: 5},
				{advanceWidth: 60, lsb: 6},
			},
			expLSB: []int16{7},
		},
		{
			fnt: &font{
				maxp: &maxpTable{
					numGlyphs: 13,
				},
				hhea: &hheaTable{
					numberOfHMetrics: 10,
				},
				hmtx: &hmtxTable{
					hMetrics: []longHorMetric{
						{advanceWidth: 10, lsb: 1},
						{advanceWidth: 20, lsb: 2},
						{advanceWidth: 30, lsb: 3},
						{advanceWidth: 40, lsb: 4},
						{advanceWidth: 50, lsb: 5},
						{advanceWidth: 60, lsb: 6}, // should include this once optimized.
						{advanceWidth: 60, lsb: 7},
						{advanceWidth: 60, lsb: 8},
						{advanceWidth: 60, lsb: 9},
						{advanceWidth: 60, lsb: 10},
					},
					leftSideBearings: []int16{
						11,
						12,
						13,
					},
				},
			},
			expNumGlyphs:   13,
			expNumHMetrics: 6,
			exphMetrics: []longHorMetric{
				{advanceWidth: 10, lsb: 1},
				{advanceWidth: 20, lsb: 2},
				{advanceWidth: 30, lsb: 3},
				{advanceWidth: 40, lsb: 4},
				{advanceWidth: 50, lsb: 5},
				{advanceWidth: 60, lsb: 6},
			},
			expLSB: []int16{7, 8, 9, 10, 11, 12, 13},
		},
	}

	for _, tcase := range testcases {
		tcase.fnt.optimizeHmtx()
		assert.EqualValues(t, tcase.expNumGlyphs, tcase.fnt.maxp.numGlyphs)
		assert.EqualValues(t, tcase.expNumHMetrics, tcase.fnt.hhea.numberOfHMetrics)
		assert.Len(t, tcase.fnt.hmtx.hMetrics, tcase.expNumHMetrics)
		assert.Len(t, tcase.fnt.hmtx.leftSideBearings, tcase.expNumGlyphs-tcase.expNumHMetrics)
		assert.Equal(t, tcase.expLSB, tcase.fnt.hmtx.leftSideBearings)
		assert.Equal(t, tcase.exphMetrics, tcase.fnt.hmtx.hMetrics)
	}
}

// TestParseHmtx_RejectsLengthMismatch: hmtx has no length field of its own -
// its size is implied by numberOfHMetrics (hhea) and numGlyphs (maxp). A
// font where those imply more bytes than the hmtx table record actually
// declares must be rejected, not silently read past hmtx into whatever
// bytes follow it in the file. White-box (package unitype): no bundled font
// has this inconsistency.
func TestParseHmtx_RejectsLengthMismatch(t *testing.T) {
	// hhea says 3 hMetrics (3*4=12 bytes) and maxp says 5 glyphs, so 2
	// trailing lsb entries (2*2=4 bytes) are implied: 16 bytes total. The
	// table record only declares 10, so parseHmtx must reject rather than
	// read 6 bytes belonging to whatever follows hmtx in the file.
	f := &font{
		maxp: &maxpTable{numGlyphs: 5},
		hhea: &hheaTable{numberOfHMetrics: 3},
		trec: &tableRecords{
			trMap: map[string]*tableRecord{
				"hmtx": {offset: 0, length: 10},
			},
		},
	}
	data := bytes.Repeat([]byte{0xFF}, 32)
	_, err := f.parseHmtx(newByteReader(bytes.NewReader(data)))
	assert.ErrorIs(t, err, errRangeCheck)
}

// TestParseHmtx_AcceptsExactLength is the positive-path complement to
// TestParseHmtx_RejectsLengthMismatch: a table record whose declared length
// exactly matches what numberOfHMetrics/numGlyphs imply must parse.
func TestParseHmtx_AcceptsExactLength(t *testing.T) {
	f := &font{
		maxp: &maxpTable{numGlyphs: 5},
		hhea: &hheaTable{numberOfHMetrics: 3},
		trec: &tableRecords{
			trMap: map[string]*tableRecord{
				"hmtx": {offset: 0, length: 16}, // 3*4 + 2*2
			},
		},
	}
	data := bytes.Repeat([]byte{0x00}, 16)
	table, err := f.parseHmtx(newByteReader(bytes.NewReader(data)))
	require.NoError(t, err)
	assert.Len(t, table.hMetrics, 3)
	assert.Len(t, table.leftSideBearings, 2)
}

// TestParseHmtx_ClampsShortLeftSideBearings: unlike a short hMetrics block
// (which has no defined fallback and is rejected outright), a table record
// that has every hMetric entry but falls short on the trailing
// leftSideBearings array must clamp rather than reject - those glyphs simply
// have no explicit left-side-bearing, which is no worse than any font that
// omits hmtx entirely.
func TestParseHmtx_ClampsShortLeftSideBearings(t *testing.T) {
	f := &font{
		maxp: &maxpTable{numGlyphs: 5},
		hhea: &hheaTable{numberOfHMetrics: 3},
		trec: &tableRecords{
			trMap: map[string]*tableRecord{
				// 3*4 = 12 bytes for hMetrics, only 2 bytes left over for
				// the 2 implied trailing lsb entries (2*2 = 4 bytes wanted).
				"hmtx": {offset: 0, length: 14},
			},
		},
	}
	data := bytes.Repeat([]byte{0x00}, 14)
	table, err := f.parseHmtx(newByteReader(bytes.NewReader(data)))
	require.NoError(t, err)
	assert.Len(t, table.hMetrics, 3, "hMetrics must be complete even when lsb is clamped")
	assert.Len(t, table.leftSideBearings, 1, "leftSideBearings must clamp to what the declared length actually fits")
}
