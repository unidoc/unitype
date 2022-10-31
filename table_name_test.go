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

func TestNameTable(t *testing.T) {
	testcases := []struct {
		fontPath   string
		numEntries int
		expected   map[int]string
	}{
		{
			"./testdata/FreeSans.ttf",
			24,
			map[int]string{
				0:  "Copyleft 2002, 2003, 2005 Free Software Foundation.",
				1:  "FreeSans",
				2:  "Medium",
				4:  "Free Sans",
				13: "The use of this font is granted subject to GNU General Public License.",
				19: "The quick brown fox jumps over the lazy dog.",
			},
		},
		{
			"./testdata/wts11.ttf",
			44,
			map[int]string{
				0:  "(C)Copyright Dr. Hann-Tzong Wang, 2002-2004.",
				1:  "HanWang KaiBold-Gb5",
				2:  "Regular",
				3:  "HanWang KaiBold-Gb5",
				4:  "HanWang KaiBold-Gb5",
				6:  "HanWang KaiBold-Gb5",
				7:  "HanWang KaiBold-Gb5 is a registered trademark of HtWang Graphics Laboratory",
				14: "http://www.gnu.org/licenses/gpl.txt",
			},
		},
		{
			"./testdata/roboto/Roboto-BoldItalic.ttf",
			26,
			map[int]string{
				0:  "Copyright 2011 Google Inc. All Rights Reserved.",
				1:  "Roboto",
				2:  "Bold Italic",
				3:  "Roboto Bold Italic",
				4:  "Roboto Bold Italic",
				5:  "Version 2.137; 2017",
				6:  "Roboto-BoldItalic",
				14: "http://www.apache.org/licenses/LICENSE-2.0",
			},
		},
	}

	for _, tcase := range testcases {
		t.Run(tcase.fontPath, func(t *testing.T) {
			f, err := os.Open(tcase.fontPath)
			assert.Equal(t, nil, err)
			defer f.Close()

			br := newByteReader(f)
			fnt, err := parseFont(br)
			assert.Equal(t, nil, err)
			require.NoError(t, err)

			require.NotNil(t, fnt)
			require.NotNil(t, fnt.name)
			require.NotNil(t, fnt.name.nameRecords)

			assert.Equal(t, tcase.numEntries, len(fnt.name.nameRecords))
			for nameID, expStr := range tcase.expected {
				assert.Equal(t, expStr, fnt.GetNameByID(nameID))
			}

			for _, nr := range fnt.name.nameRecords {
				t.Logf("%d/%d/%d - '%s'", nr.platformID, nr.encodingID, nr.nameID, nr.Decoded())
			}
		})
	}
}

func TestGetNameRecords(t *testing.T) {
	testcases := []struct {
		fontPath string
		numNames int
		expected map[uint16]map[uint16]string
	}{
		{
			"./testdata/FreeSans.ttf",
			3,
			map[uint16]map[uint16]string{
				0: map[uint16]string{
					0:  "Copyleft 2002, 2003, 2005 Free Software Foundation.",
					1:  "FreeSans",
					2:  "Medium",
					3:  "FontForge 2.0 : Free Sans : 18-5-2007",
					4:  "Free Sans",
					5:  "Version $Revision: 1.79 $ ",
					6:  "FreeSans",
					13: "The use of this font is granted subject to GNU General Public License.",
					14: "http://www.gnu.org/copyleft/gpl.html",
					19: "The quick brown fox jumps over the lazy dog.",
				},
				1033: map[uint16]string{
					0:  "Copyleft 2002, 2003, 2005 Free Software Foundation.",
					1:  "FreeSans",
					2:  "Medium",
					3:  "FontForge 2.0 : Free Sans : 18-5-2007",
					4:  "Free Sans",
					5:  "Version $Revision: 1.79 $ ",
					6:  "FreeSans",
					13: "The use of this font is granted subject to GNU General Public License.",
					14: "http://www.gnu.org/copyleft/gpl.html",
					19: "The quick brown fox jumps over the lazy dog.",
				},
				1060: map[uint16]string{
					2:  "navadno",
					13: "Dovoljena je uporaba v skladu z licenco GNU General Public License.",
					14: "http://www.gnu.org/copyleft/gpl.html",
					19: "Šerif bo za vajo spet kuhal domače žgance.",
				},
			},
		},
		{
			"./testdata/wts11.ttf",
			4,
			map[uint16]map[uint16]string{
				0: map[uint16]string{
					0:  "(C)Copyright Dr. Hann-Tzong Wang, 2002-2004.",
					1:  "HanWang KaiBold-Gb5",
					2:  "Regular",
					3:  "HanWang KaiBold-Gb5",
					4:  "HanWang KaiBold-Gb5",
					5:  "Version 1.3(license under GNU GPL)",
					6:  "HanWang KaiBold-Gb5",
					7:  "HanWang KaiBold-Gb5 is a registered trademark of HtWang Graphics Laboratory",
					10: "HtWang Fonts(1), March 8, 2002; 1.00, initial release; HtWang Fonts(17), March 5, 2004; GJL(040519). Maintain by CLE Project.",
					13: "(C)Copyright Dr. Hann-Tzong Wang, 2002-2004.\nThis program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or any later version.",
					14: "http://www.gnu.org/licenses/gpl.txt",
				},
				1028: map[uint16]string{
					0:  "(C)Copyright Dr. Hann-Tzong Wang, 2002-2004.",
					1:  "王漢宗粗楷體簡",
					2:  "Regular",
					3:  "王漢宗粗楷體簡",
					4:  "王漢宗粗楷體簡",
					5:  "Version 1.3(license under GNU GPL)",
					6:  "王漢宗粗楷體簡",
					7:  "王漢宗粗楷體簡 is a registered trademark of HtWang Graphics Laboratory",
					10: "HtWang Fonts(1), March 8, 2002; 1.00, initial release; HtWang Fonts(17), March 5, 2004; GJL(040519). Maintain by CLE Project.",
					13: "(C)Copyright Dr. Hann-Tzong Wang, 2002-2004.\nThis program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or any later version.",
					14: "http://www.gnu.org/licenses/gpl.txt",
				},
				1033: map[uint16]string{
					0:  "(C)Copyright Dr. Hann-Tzong Wang, 2002-2004.",
					1:  "HanWang KaiBold-Gb5",
					2:  "Regular",
					3:  "HanWang KaiBold-Gb5",
					4:  "HanWang KaiBold-Gb5",
					5:  "Version 1.3(license under GNU GPL)",
					6:  "HanWang KaiBold-Gb5",
					7:  "HanWang KaiBold-Gb5 is a registered trademark of HtWang Graphics Laboratory",
					10: "HtWang Fonts(1), March 8, 2002; 1.00, initial release; HtWang Fonts(17), March 5, 2004; GJL(040519). Maintain by CLE Project.",
					13: "(C)Copyright Dr. Hann-Tzong Wang, 2002-2004.\nThis program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or any later version.",
					14: "http://www.gnu.org/licenses/gpl.txt",
				},
				2052: map[uint16]string{
					0:  "(C)Copyright Dr. Hann-Tzong Wang, 2002-2004.",
					1:  "王汉宗粗楷体简",
					2:  "Regular",
					3:  "王汉宗粗楷体简",
					4:  "王汉宗粗楷体简",
					5:  "Version 1.3(license under GNU GPL)",
					6:  "王汉宗粗楷体简",
					7:  "王汉宗粗楷体简 is a registered trademark of HtWang Graphics Laboratory",
					10: "HtWang Fonts(1), March 8, 2002; 1.00, initial release; HtWang Fonts(17), March 5, 2004; GJL(040519). Maintain by CLE Project.",
					13: "(C)Copyright Dr. Hann-Tzong Wang, 2002-2004.\nThis program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 2 of the License, or any later version.",
					14: "http://www.gnu.org/licenses/gpl.txt",
				},
			},
		},
	}

	for _, tc := range testcases {
		f, err := os.Open(tc.fontPath)
		require.NoError(t, err)
		defer f.Close()

		br := newByteReader(f)
		fnt, err := parseFont(br)
		require.NoError(t, err)

		// Get name records.
		nameRecords := fnt.GetNameRecords()

		assert.Equal(t, tc.numNames, len(nameRecords))
		assert.Equal(t, tc.expected, nameRecords)
	}
}
