/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"io"
	"os"
	"testing"
)

// FuzzParse asserts Parse, and Write on whatever Parse accepts, return errors
// on malformed input rather than panicking.
func FuzzParse(f *testing.F) {
	for _, path := range []string{"testdata/roboto/Roboto-Regular.ttf", "testdata/FreeSans.ttf"} {
		data, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		fnt, err := Parse(bytes.NewReader(data))
		if err != nil {
			return
		}
		_ = fnt.Write(io.Discard)
	})
}
