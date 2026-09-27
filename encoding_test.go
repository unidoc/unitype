/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMacEncoding(t *testing.T) {
	// Spot checks based on: https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6post.html
	assert.Equal(t, 258, len(macGlyphNames))
	assert.Equal(t, GlyphName(".notdef"), macGlyphNames[0])
	assert.Equal(t, GlyphName("space"), macGlyphNames[3])
	assert.Equal(t, GlyphName("comma"), macGlyphNames[15])
	assert.Equal(t, GlyphName("a"), macGlyphNames[68])
	assert.Equal(t, GlyphName("z"), macGlyphNames[93])
	assert.Equal(t, GlyphName("dcroat"), macGlyphNames[257])
}

// TestDecodeCharcode asserts decodeCharcode matches DecodeRune(ToBytes(c))
// for every encoding, across the UCS-2 range, the Unicode range with its
// surrogates, and codes beyond it.
func TestDecodeCharcode(t *testing.T) {
	var codes []uint32
	for c := uint32(0); c <= 0x2FFFF; c++ {
		codes = append(codes, c)
	}
	for c := uint32(0x30000); c <= 0x10FFFF; c += 97 {
		codes = append(codes, c)
	}
	codes = append(codes, 0x10FFFF, 0x110000, 0x110001, 0x7FFFFFFF, 0x80000000, 0xFFFFFFFF)

	for _, e := range []cmapEncoding{cmapEncodingUCS2, cmapEncodingUCS4, cmapEncodingMacRoman, cmapEncodingShiftJIS,
		cmapEncodingPRC, cmapEncodingBig5, cmapEncodingJohab, cmapEncodingUnsupported} {
		d := e.GetRuneDecoder()
		for _, c := range codes {
			if want, got := d.DecodeRune(d.ToBytes(c)), d.decodeCharcode(c); want != got {
				t.Fatalf("encoding %d, charcode 0x%X: decodeCharcode = %U, DecodeRune(ToBytes) = %U", e, c, got, want)
			}
		}
	}
}
