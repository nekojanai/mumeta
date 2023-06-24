package utils

func DecodeSyncSafeIntegers(b []byte) int {
	var result int
	for _, v := range b {
		result = result<<7 | int(v&0x7F)
	}
	return result
}
