package id3v2

import (
	"fmt"
	"nekojanai/mumeta/utils"
	"os"
)

const id3v2HeaderSize = 10

var id3v2HeaderIDValue = [3]byte{0x49, 0x44, 0x33}

type ID3v2Header struct {
	Version [2]byte
	flags   byte
	size    [4]byte
}

func ReadID3v2Header(f *os.File) (header ID3v2Header, err error) {
	fileInfo, err := f.Stat()
	if err != nil || fileInfo.Size() < id3v2HeaderSize {
		return
	}
	h := make([]byte, id3v2HeaderSize)
	n, err := f.Read(h)
	if err != nil || n != id3v2HeaderSize {
		return
	}
	if h[0] == id3v2HeaderIDValue[0] &&
		h[1] == id3v2HeaderIDValue[1] &&
		h[2] == id3v2HeaderIDValue[2] &&
		h[3] < 0xFF &&
		h[4] < 0xFF &&
		h[5]&0x0F == 0 &&
		h[6] < 0x80 &&
		h[7] < 0x80 &&
		h[8] < 0x80 &&
		h[9] < 0x80 {
		header = ID3v2Header{
			Version: [2]byte{h[3], h[4]},
			flags:   h[5],
			size:    [4]byte{h[6], h[7], h[8], h[9]},
		}
	} else {
		err = fmt.Errorf("No ID3v2 header")
	}
	return
}

func (h ID3v2Header) HumanString() string {
	if h.IsEmpty() {
		return "{}"
	}
	version := fmt.Sprintf("------- ID3v2.%d.%d Header -------\n", h.Version[0], h.Version[1])
	size := fmt.Sprintf("Size: \t%d\n", utils.DecodeSyncSafeIntegers(h.size[:]))
	flags := fmt.Sprintf("Flags: \t%v\n", h.Flags())
	footer := "--------------------------------"
	return fmt.Sprintf("%s%s%s%s", version, size, flags, footer)
}

func (h ID3v2Header) IsEmpty() bool {
	return h.Version == [2]byte{} &&
		h.flags == 0 &&
		h.size == [4]byte{}
}

func (h ID3v2Header) Flags() ID3v2HeaderFlags {
	return ID3v2HeaderFlags{
		Unsynchronisation: h.flags&0x80 != 0,
		ExtendedHeader:    h.flags&0x40 != 0,
		Experimental:      h.flags&0x20 != 0,
		Footer:            h.flags&0x10 != 0,
	}
}

type ID3v2HeaderFlags struct {
	// Supported added in ID3v2.2
	Unsynchronisation bool
	// Support added in ID3v2.3
	ExtendedHeader bool
	// Support added in ID3v2.3
	Experimental bool
	// Support added in ID3v2.4
	Footer bool
}

func (h ID3v2HeaderFlags) HumanString() string {
	unsynchronisation := fmt.Sprintf("Unsynchronisation: %v\n", h.Unsynchronisation)
	extendedHeader := fmt.Sprintf("\tExtendedHeader:    %v\n", h.ExtendedHeader)
	experimental := fmt.Sprintf("\tExperimental:      %v\n", h.Experimental)
	footer := fmt.Sprintf("\tFooter:            %v", h.Footer)
	return fmt.Sprintf("%s%s%s%s",
		unsynchronisation, extendedHeader, experimental, footer)
}
