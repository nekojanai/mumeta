package id3v2

import (
	"nekojanai/mumeta/utils"
	"os"
)

type ID3v2Frame struct {
	FrameID [4]byte
	Size    [4]byte
	Flags   [2]byte
	Data    []byte
}

func ReadID3v2Frames(f *os.File, header ID3v2Header) (frames []ID3v2Frame, err error) {
	if header.IsEmpty() {
		header, err = ReadID3v2Header(f)
		if err != nil {
			return
		}
	}
	framesOffset := int64(id3v2HeaderSize)
	if header.Flags().ExtendedHeader {
		var extendedHeader ID3v2ExtendedHeader
		extendedHeader, err = ReadID3v2ExtendedHeader(f, header)
		if err != nil {
			return
		}
		framesOffset += int64(utils.DecodeSyncSafeIntegers(extendedHeader.Size[:]))
		if header.Version == [2]byte{0x03, 0x00} {
			framesOffset += 4
		}
	}

	return
}

func readID3v2Frame(f *os.File, offset int64) (frame ID3v2Frame, err error) {
	frameID := make([]byte, 4)
	_, err = f.ReadAt(frameID, offset)
	if err != nil {
		return
	}
	offset += 4
	size := make([]byte, 4)
	_, err = f.ReadAt(size, offset)
	if err != nil {
		return
	}
	// TODO: continue here
	return
}
