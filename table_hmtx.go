/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package unitype

import (
	"errors"
	"io"

	"github.com/sirupsen/logrus"
)

type hmtxTable struct {
	hMetrics         []longHorMetric // length is numberOfHMetrics from hhea table.
	leftSideBearings []int16         // length is (numGlyphs - numberOfHmetrics) from maxp and hhea tables.
}

type longHorMetric struct {
	advanceWidth uint16
	lsb          int16
}

func (f *font) parseHmtx(r *byteReader) (*hmtxTable, error) {
	if f.maxp == nil || f.hhea == nil {
		logrus.Debug("maxp or hhea table missing")
		return nil, errRequiredField
	}

	tr, has, err := f.seekToTable(r, "hmtx")
	if err != nil {
		return nil, err
	}
	if !has {
		logrus.Debug("hmtx table absent")
		return nil, nil
	}

	// hmtx's size has no length field of its own (implied by hhea/maxp
	// instead, per the OpenType spec). numberOfHMetrics is clamped to
	// numGlyphs and to the entries the table holds, and hhea is updated to
	// match so Write stays consistent; a short trailing lsb array is
	// zero-padded, whether cut short by the declared length or by the end of
	// the file.
	numberOfHMetrics := int(f.hhea.numberOfHMetrics)
	if n := int(f.maxp.numGlyphs); numberOfHMetrics > n {
		numberOfHMetrics = n
	}
	clampedByLength := false
	if n := int64(tr.length) / 4; int64(numberOfHMetrics) > n {
		numberOfHMetrics = int(n)
		clampedByLength = true
		if numberOfHMetrics == 0 {
			logrus.Debug("hmtx table holds no advance widths")
			return nil, errRangeCheck
		}
	}
	if numberOfHMetrics != int(f.hhea.numberOfHMetrics) {
		logrus.Debugf("hmtx: clamping numberOfHMetrics from %d to %d", f.hhea.numberOfHMetrics, numberOfHMetrics)
		f.hhea.numberOfHMetrics = uint16(numberOfHMetrics)
	}
	wantHMetricsLen := 4 * numberOfHMetrics

	lsbLen := int(f.maxp.numGlyphs) - numberOfHMetrics
	readLsbLen := lsbLen
	avail := (int64(tr.length) - int64(wantHMetricsLen)) / 2
	if clampedByLength {
		// The leftover bytes are part of a truncated hMetrics entry, not
		// leftSideBearings.
		avail = 0
	}
	if int64(readLsbLen) > avail {
		logrus.Debug("hmtx leftSideBearings shorter than numGlyphs implies, zero-padding")
		readLsbLen = int(avail)
	}

	t := &hmtxTable{}

	for i := 0; i < numberOfHMetrics; i++ {
		var lhm longHorMetric
		err := r.read(&lhm.advanceWidth, &lhm.lsb)
		if err != nil {
			return nil, err
		}

		t.hMetrics = append(t.hMetrics, lhm)
	}

	if readLsbLen > 0 {
		err = r.readSlice(&t.leftSideBearings, readLsbLen)
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, err
		}
	}
	if missing := lsbLen - len(t.leftSideBearings); missing > 0 {
		t.leftSideBearings = append(t.leftSideBearings, make([]int16, missing)...)
	}

	return t, nil
}

// optimizeHmtx optimizes the htmx table.
func (f *font) optimizeHmtx() {
	i := len(f.hmtx.hMetrics) - 1
	if i <= 0 {
		return
	}
	lastWidth := f.hmtx.hMetrics[i].advanceWidth
	j := i - 1
	for j >= 0 && f.hmtx.hMetrics[j].advanceWidth == lastWidth {
		j--
	}
	numStrip := i - j - 1
	if numStrip == 0 {
		return
	}

	f.hhea.numberOfHMetrics = uint16(j + 2)
	var lsbPrepend []int16
	for k := j + 2; k <= i; k++ {
		lsbPrepend = append(lsbPrepend, f.hmtx.hMetrics[k].lsb)
	}
	f.hmtx.leftSideBearings = append(lsbPrepend, f.hmtx.leftSideBearings...)
	f.hmtx.hMetrics = f.hmtx.hMetrics[0 : j+2]
}

// writeHmtx writes the font's hmtx table  to `w`.
func (f *font) writeHmtx(w *byteWriter) error {
	if f.hmtx == nil || f.hhea == nil {
		return nil
	}

	for _, lhm := range f.hmtx.hMetrics {
		err := w.write(lhm.advanceWidth, lhm.lsb)
		if err != nil {
			return err
		}
	}

	return w.writeSlice(f.hmtx.leftSideBearings)
}
