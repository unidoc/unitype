/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"testing"
)

// fuzzAllocBase and fuzzAllocPerByte set the most FuzzParse lets Parse and
// Write allocate for one input: fuzzAllocBase plus fuzzAllocPerByte bytes per
// input byte. Parsing FreeSans (460 KB) allocates about 17 MB.
const (
	fuzzAllocBase    = 64 << 20
	fuzzAllocPerByte = 128
)

// FuzzParse asserts Parse, and Write on whatever Parse accepts, return errors
// on malformed input rather than panicking, and stay within an allocation
// budget proportional to the input size.
func FuzzParse(f *testing.F) {
	for _, path := range []string{"testdata/roboto/Roboto-Regular.ttf", "testdata/FreeSans.ttf"} {
		data, err := os.ReadFile(path)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		if fnt, err := Parse(bytes.NewReader(data)); err == nil {
			_ = fnt.Write(io.Discard)
		}
		runtime.ReadMemStats(&after)

		budget := uint64(fuzzAllocBase + fuzzAllocPerByte*len(data))
		if alloc := after.TotalAlloc - before.TotalAlloc; alloc > budget {
			t.Fatalf("allocated %d bytes for a %d-byte input (budget %d)", alloc, len(data), budget)
		}
	})
}
