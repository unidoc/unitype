/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkParse measures Parse on bundled fonts of increasing size.
func BenchmarkParse(b *testing.B) {
	for _, path := range []string{"testdata/roboto/Roboto-Regular.ttf", "testdata/FreeSans.ttf", "testdata/wts11.ttf"} {
		data, err := os.ReadFile(path)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(filepath.Base(path), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			for i := 0; i < b.N; i++ {
				if _, err := Parse(bytes.NewReader(data)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
