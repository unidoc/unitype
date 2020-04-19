package unitype

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Try reading in a glyf table and write back out, read again and verify.
func TestGlyfReadWrite(t *testing.T) {
	testcases := []struct {
		fontPath string
	}{
		{
			"./testdata/FreeSans.ttf",
		},
		/*
			{
				"../../creator/testdata/wts11.ttf",
				14148,
			},
			{
				"../../creator/testdata/roboto/Roboto-BoldItalic.ttf",
				1294,
			},
		*/
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
			require.NotNil(t, fnt.glyf)

			// Read the glyf table from the font.
			tr := fnt.trec.trMap["glyf"]
			f.Seek(int64(tr.offset), io.SeekStart)
			b := make([]byte, tr.length)
			_, err = io.ReadFull(f, b)
			require.NoError(t, err)

			// Write the glyf table out to glyfBuf.
			var glyfBuf bytes.Buffer
			glyfw := newByteWriter(&glyfBuf)
			err = fnt.writeGlyf(glyfw)
			require.NoError(t, err)
			err = glyfw.flush()
			require.NoError(t, err)

			fmt.Printf("Read (%d):\n", len(b))
			fmt.Printf("Write (%d):\n", len(glyfBuf.Bytes()))
			require.Equal(t, b, glyfBuf.Bytes())

			// Hack to set position right for this buffer.
			tr.offset = 0
			fnt.trec.trMap["glyf"] = tr

			// Try to parse?
			glyfr := newByteReader(bytes.NewReader(glyfBuf.Bytes()))

			glyft, err := fnt.parseGlyf(glyfr)
			require.NoError(t, err)
			require.Equal(t, fnt.glyf, glyft)
		})
	}
}
