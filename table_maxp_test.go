/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
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
