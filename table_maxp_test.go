/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxpTable(t *testing.T) {
	// Run only this function in debugmode.
	/*
		common.SetLogger(common.NewConsoleLogger(common.LogLevelDebug))
		defer func() {
			common.SetLogger(common.NewConsoleLogger(common.LogLevelInfo))
		}()
	*/

	testcases := []struct {
		fontPath  string
		numGlyphs int
	}{
		{
			"./testdata/FreeSans.ttf",
			3726,
		},
		{
			"./testdata/wts11.ttf",
			14148,
		},
		{
			"./testdata/roboto/Roboto-BoldItalic.ttf",
			1294,
		},
	}

	for _, tcase := range testcases {
		t.Run(tcase.fontPath, func(t *testing.T) {
			t.Logf("%s", tcase.fontPath)
			f, err := os.Open(tcase.fontPath)
			assert.Equal(t, nil, err)
			defer f.Close()

			br := newByteReader(f)
			fnt, err := parseFont(br)
			assert.Equal(t, nil, err)
			require.NoError(t, err)

			require.NotNil(t, fnt)
			require.NotNil(t, fnt.maxp)
			require.Equal(t, int(tcase.numGlyphs), int(fnt.maxp.numGlyphs))
		})
	}
}

// TestParseMaxp_RejectsShortTable asserts a maxp record shorter than its
// fields fails rather than reading the bytes after it, including one too
// short for numGlyphs.
func TestParseMaxp_RejectsShortTable(t *testing.T) {
	// version=1.0 (0x00010000, big-endian) + numGlyphs=5, then sentinel bytes.
	data := append([]byte{0x00, 0x01, 0x00, 0x00, 0x00, 0x05}, bytes.Repeat([]byte{0xFF}, 40)...)
	for _, length := range []uint32{4, 6} {
		f := &font{
			trec: &tableRecords{
				trMap: map[string]*tableRecord{
					"maxp": {offset: 0, length: length},
				},
			},
		}
		_, err := f.parseMaxp(newByteReader(bytes.NewReader(data)))
		assertTruncatedRead(t, err)
	}
}
