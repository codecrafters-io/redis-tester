package internal

// redisBitAt returns the bit at offset in value using Redis bitmap layout:
// bit 0 is the most significant bit of the first byte. Offsets past the end
// of value are treated as 0.
func redisBitAt(value string, offset int) int {
	if offset < 0 {
		return 0
	}

	byteIndex := offset / 8
	if byteIndex >= len(value) {
		return 0
	}

	bitIndex := offset % 8
	return int((value[byteIndex] >> (7 - bitIndex)) & 1)
}

func redisBitOffsets(value string, bit int) []int {
	var offsets []int
	for offset := 0; offset < len(value)*8; offset++ {
		if redisBitAt(value, offset) == bit {
			offsets = append(offsets, offset)
		}
	}
	return offsets
}
