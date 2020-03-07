package unitype

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
