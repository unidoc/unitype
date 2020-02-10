/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package truetype

// maxUint16 returns the maximium value in `sl` or 0 if empty.
func maxUint16(sl []uint16) uint16 {
	var max uint16
	for i, val := range sl {
		if i == 0 {
			max = val
			continue
		}

		if val > max {
			max = val
		}
	}
	return max
}
