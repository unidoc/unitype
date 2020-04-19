package unitype

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestReadWrite(t *testing.T) {
	testcases := []struct {
		fontPath string
	}{
		{
			"./testdata/FreeSans.ttf",
		},
		/*
			{
				"./testdata/wts11.ttf",
			},
			{
				"./testdata/roboto/Roboto-BoldItalic.ttf",
			},
		*/
	}

	for _, tcase := range testcases {
		t.Logf("%s", tcase.fontPath)
		fnt, err := ParseFile(tcase.fontPath)
		require.NoError(t, err)

		logrus.Debug("Write")
		outPath := "/tmp/1.ttf"

		t.Logf("WriteFile -> %s", outPath)
		err = fnt.WriteFile(outPath)
		require.NoError(t, err)

		t.Logf("Validating %s...", outPath)
		err = ValidateFile(outPath)
		require.NoError(t, err)
	}
}
