package utils

import "fmt"

func ReverseByte(b byte) byte {
	b = (b&0xF0)>>4 | (b&0x0F)<<4
	b = (b&0xCC)>>2 | (b&0x33)<<2
	b = (b&0xAA)>>1 | (b&0x55)<<1
	return b
}

func GetBitAt(b byte, offset uint8) bool {
	if offset > 7 {
		return false
	}
	return (b & (1 << offset)) == 1
}

func PrintBytes(b []byte) {
	for _, v := range b {
		fmt.Printf("%s ", ToBinaryString(v))
	}
	fmt.Printf("\n")
}

func ToBinaryString(b byte) string {
	switch {
	case b < 1:
		return "00000000"
	case b < 2:
		return fmt.Sprintf("0000000%b", b)
	case b < 4:
		return fmt.Sprintf("000000%b", b)
	case b < 8:
		return fmt.Sprintf("00000%b", b)
	case b < 16:
		return fmt.Sprintf("0000%b", b)
	case b < 32:
		return fmt.Sprintf("000%b", b)
	case b < 64:
		return fmt.Sprintf("00%b", b)
	case b < 128:
		return fmt.Sprintf("0%b", b)
	default:
		return fmt.Sprintf("%b", b)
	}
}
