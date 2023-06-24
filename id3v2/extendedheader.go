package id3v2

import (
	"fmt"
	"nekojanai/mumeta/utils"
	"os"
)

type ID3v2ExtendedHeader struct {
	Size [4]byte
	// 1 byte for ID3v2.4.0 and 2 bytes for ID3v2.3.0
	ExtendedFlags []byte
	version       [2]byte
	data          []byte
}

func (t ID3v2ExtendedHeader) IsEmpty() bool {
	return t.Size == [4]byte{0, 0, 0, 0} &&
		t.ExtendedFlags == nil &&
		t.version == [2]byte{0, 0} &&
		t.data == nil
}

func (t ID3v2ExtendedHeader) HumanString() (str string) {
	if t.IsEmpty() {
		return "{}"
	}
	version := fmt.Sprintf("------- ID3v2.%d.%d Extended Header -------\n", t.version[0], t.version[1])
	size := fmt.Sprintf("Size: %d\n", utils.DecodeSyncSafeIntegers(t.Size[:]))
	extendedFlags := fmt.Sprintf("Extended Flags: %s\n", t.ParseExtendedFlags())
	v4, v3 := t.Fields()
	var versionSpecificFields string
	if t.version == [2]byte{0x04, 0x00} {
		versionSpecificFields = fmt.Sprintf("%+v\n", v4)
	}
	if t.version == [2]byte{0x03, 0x00} {
		versionSpecificFields = fmt.Sprintf("%+v\n", v3)
	}
	footer := "-----------------------------------------"
	str = fmt.Sprintf("%s%s%s%+v%s", version, size, extendedFlags, versionSpecificFields, footer)
	return
}

type ID3v2ExtendedHeaderV4ExtendedFlags struct {
	TagIsUpdate     bool
	CRCDataPresent  bool
	TagRestrictions bool
}

func (t ID3v2ExtendedHeaderV4ExtendedFlags) HumanString() string {
	tagIsUpdate := fmt.Sprintf("Tag is an update: %v\n", t.TagIsUpdate)
	crcDataPresent := fmt.Sprintf("\t\tCRC data present: %v\n", t.CRCDataPresent)
	tagRestrictions := fmt.Sprintf("\t\tTag restrictions: %v\n", t.TagRestrictions)
	return fmt.Sprintf("%s%s%s", tagIsUpdate, crcDataPresent, tagRestrictions)
}

type ID3v2ExtendedHeaderV3ExtendedFlags struct {
	CRCDataPresent bool
}

func (t ID3v2ExtendedHeaderV3ExtendedFlags) HumanString() string {
	crcDataPresent := fmt.Sprintf("CRC data present: %v\n", t.CRCDataPresent)
	return fmt.Sprintf("%s", crcDataPresent)
}

func (t ID3v2ExtendedHeader) ParseExtendedFlags() (flags string) {
	if t.IsEmpty() {
		return "{}"
	}
	if t.version == [2]byte{0x04, 0x00} {
		flags = fmt.Sprintf("%v", ID3v2ExtendedHeaderV4ExtendedFlags{
			t.ExtendedFlags[0]&0x40 != 0,
			t.ExtendedFlags[0]&0x20 != 0,
			t.ExtendedFlags[0]&0x10 != 0,
		})
	}
	if t.version == [2]byte{0x03, 0x00} {
		flags = fmt.Sprintf("%v", ID3v2ExtendedHeaderV3ExtendedFlags{
			t.ExtendedFlags[0]&0x80 != 0,
		})
	}
	return
}

func ReadID3v2ExtendedHeader(f *os.File, header ID3v2Header) (extendedHeader ID3v2ExtendedHeader, err error) {
	if header.IsEmpty() {
		header, err = ReadID3v2Header(f)
		if err != nil {
			return
		}
	}
	if !header.Flags().ExtendedHeader {
		err = fmt.Errorf("extended header flag not set")
		return
	}
	offset := int64(10)
	size := make([]byte, 4)
	n, err := f.ReadAt(size, offset)
	if err != nil || n != 4 {
		return
	}
	offset += 4

	nextByte := make([]byte, 1)
	n, err = f.ReadAt(nextByte, offset)
	if err != nil || n != 1 {
		return
	}
	if header.Version == [2]byte{0x04, 0x00} && nextByte[0] != 0x01 {
		err = fmt.Errorf("invalid number of flag bytes (ID3v2.4.0)")
		return
	}
	if header.Version == [2]byte{0x03, 0x00} && nextByte[0]&0x7F != 0x00 {
		err = fmt.Errorf("invalid first byte of extended flags (ID3v2.3.0)")
		return
	}
	offset++

	if header.Version == [2]byte{0x03, 0x00} {
		nextFiveBytes := make([]byte, 5)
		n, err = f.ReadAt(nextFiveBytes, offset)
		if err != nil || n != 5 {
			return
		}
		if nextFiveBytes[0] != 0x00 {
			err = fmt.Errorf("invalid second byte of extended flags (ID3v2.3.0)")
			return
		}
		offset += 5
		decodedExtendedHeaderSize := utils.DecodeSyncSafeIntegers(size) - 2
		lastBytes := make([]byte, decodedExtendedHeaderSize)
		n, err = f.ReadAt(lastBytes, offset)
		if err != nil || n != decodedExtendedHeaderSize {
			return
		}
		extendedHeader = ID3v2ExtendedHeader{
			Size:          [4]byte(size),
			ExtendedFlags: append(nextByte, nextFiveBytes[0]),
			data:          append(nextFiveBytes[1:], lastBytes...),
		}
	}
	if header.Version == [2]byte{0x04, 0x00} {
		extendedHeader.Size = [4]byte(size)
		extendedFlags := make([]byte, 1)
		n, err = f.ReadAt(extendedFlags, offset)
		if err != nil || n != 1 {
			return
		}
		extendedHeader.ExtendedFlags = extendedFlags
		offset++
		decodedExtendedHeaderSize := utils.DecodeSyncSafeIntegers(size) - 6
		lastBytes := make([]byte, decodedExtendedHeaderSize)
		n, err = f.ReadAt(lastBytes, offset)
		if err != nil || n != decodedExtendedHeaderSize {
			return
		}
		extendedHeader.data = append(nextByte, lastBytes...)
	}
	extendedHeader.version = header.Version
	return
}

type ID3v230Fields struct {
	SizeOfPadding [4]byte
	TotalFrameCRC [4]byte
}

func (t ID3v230Fields) HumanString() string {
	sizeOfPadding := fmt.Sprintf("Size of padding: %v\n", t.SizeOfPadding)
	totalFrameCRC := fmt.Sprintf("Total frame CRC: %v", t.TotalFrameCRC)
	return fmt.Sprintf("%s%s", sizeOfPadding, totalFrameCRC)
}

type ID3v240Fields struct {
	NumberOfFlagBytes byte
	Flags             []ID3v2ExtendedHeaderFlag
}

func (t ID3v240Fields) HumanString() string {
	numberOfFlagBytes := fmt.Sprintf("Number of flag bytes: %v\n", t.NumberOfFlagBytes)
	flags := fmt.Sprintf("Flags: %v", t.Flags)
	return fmt.Sprintf("%s%s", numberOfFlagBytes, flags)
}

func (h ID3v2ExtendedHeader) Fields() (id3v240Fields ID3v240Fields, id3v230Fields ID3v230Fields) {
	if h.version == [2]byte{0x03, 0x00} {
		id3v230Fields = ID3v230Fields{
			SizeOfPadding: [4]byte{h.data[0], h.data[1], h.data[2], h.data[3]},
		}
		if h.ExtendedFlags[1] == 0x80 {
			id3v230Fields.TotalFrameCRC = [4]byte{h.data[4], h.data[5], h.data[6], h.data[7]}
		}
		return
	} else if h.version == [2]byte{0x04, 0x00} {
		index := 0
		id3v240Fields = ID3v240Fields{
			NumberOfFlagBytes: h.data[index],
		}
		index++
		if h.ExtendedFlags[0]&0x40 == 0x01 && h.data[index] == 0x00 {
			id3v240Fields.Flags = append(id3v240Fields.Flags, ID3v2ExtendedHeaderFlag{
				Length: 0x01,
			})
			index++
		}
		if h.ExtendedFlags[0]&0x20 == 0x01 && h.data[index] == 0x05 {
			id3v240Fields.Flags = append(id3v240Fields.Flags, ID3v2ExtendedHeaderFlag{
				Length: 0x05,
				Data:   []byte{h.data[index+1], h.data[index+2], h.data[index+3], h.data[index+4], h.data[index+5]},
			})
			index += 6
		}
		if h.ExtendedFlags[0]&0x10 == 0x01 && h.data[index] == 0x01 {
			id3v240Fields.Flags = append(id3v240Fields.Flags, ID3v2ExtendedHeaderFlag{
				Length: 0x01,
				Data:   []byte{h.data[index+1]},
			})
		}
		return
	}
	return
}

type ID3v2ExtendedHeaderFlag struct {
	Length byte
	Data   []byte
}

var extendedHeaderFlagTypes = map[byte]string{
	0x00: "Update",
	0x05: "CRC",
	0x01: "Restrictions",
}

var extendedHeaderFlagDescriptions = map[byte]string{
	0x00: "Indicates that the tag has been updated and needs to be saved again.",
	0x05: "Contains a CRC-32 checksum of the complete tag excluding the header.",
	0x01: "Indicates that the tag is restricted in some way.",
}

func (t ID3v2ExtendedHeaderFlag) HumanString() string {
	return fmt.Sprintf("Type: %s\nData: %s\n", extendedHeaderFlagTypes[t.Length], extendedHeaderFlagDescriptions[t.Length])
}

var extendedHeaderFlagRestrictionsTagSize = map[byte]string{
	0x00: "No more than 128 frames and 1 MB total tag size.",
	0x01: "No more than 64 frames and 128 KB total tag size.",
	0x02: "No more than 32 frames and 40 KB total tag size.",
	0x03: "No more than 32 frames and 4 KB total tag size.",
}

var extendedHeaderFlagRestrictionsTextEncoding = map[byte]string{
	0x00: "No restrictions.",
	0x01: "Strings are only encoded with ISO-8859-1 [ISO-8859-1] or UTF-8 [UTF-8].",
}

var extendedHeaderFlagRestrictionsTextFieldsSize = map[byte]string{
	0x00: "No restrictions.",
	0x01: "No string is longer than 1024 characters.",
	0x02: "No string is longer than 128 characters.",
	0x03: "No string is longer than 30 characters.",
}

var extendedHeaderFlagRestrictionsImageEncoding = map[byte]string{
	0x00: "No restrictions.",
	0x01: "Images are encoded only with PNG [PNG] or JPEG [JFIF].",
}

var extendedHeaderFlagRestrictionsImageSize = map[byte]string{
	0x00: "No restrictions.",
	0x01: "All images are 256x256 pixels or smaller.",
	0x02: "All images are 64x64 pixels or smaller.",
	0x03: "All images are exactly 64x64 pixels, unless required otherwise.",
}

type ID3v2ExtendedHeaderFlagRestrictions struct {
	TagSizeByte       byte
	TextEncodingByte  byte
	TextFieldSizeByte byte
	ImageEncodingByte byte
	ImageSizeByte     byte
}

func (t ID3v2ExtendedHeaderFlagRestrictions) HumanString() string {
	tagSize := fmt.Sprintf("Tag Size: %s\n", extendedHeaderFlagDescriptions[t.TagSizeByte])
	textEncoding := fmt.Sprintf("Text Encoding: %s\n", extendedHeaderFlagRestrictionsTextEncoding[t.TextEncodingByte])
	textFieldSize := fmt.Sprintf("Text Field Size: %s\n", extendedHeaderFlagRestrictionsTextFieldsSize[t.TextFieldSizeByte])
	imageEncoding := fmt.Sprintf("Image Encoding: %s\n", extendedHeaderFlagRestrictionsImageEncoding[t.ImageEncodingByte])
	imageSize := fmt.Sprintf("Image Size: %s\n", extendedHeaderFlagRestrictionsImageSize[t.ImageSizeByte])
	return fmt.Sprintf("%s%s%s%s%s", tagSize, textEncoding, textFieldSize, imageEncoding, imageSize)
}

func (t ID3v2ExtendedHeaderFlag) Restrictions() (restrictions ID3v2ExtendedHeaderFlagRestrictions) {
	if t.Length != 0x01 {
		return
	}
	tagSizeByte := (t.Data[0] & 0xC0) >> 6
	textEncodingByte := (t.Data[0] & 0x20) >> 5
	textFieldSizeByte := (t.Data[0] & 0x18) >> 3
	imageEncodingByte := (t.Data[0] & 0x04) >> 2
	imageSizeByte := t.Data[0] & 0x03
	restrictions = ID3v2ExtendedHeaderFlagRestrictions{
		TagSizeByte:       tagSizeByte,
		TextEncodingByte:  textEncodingByte,
		TextFieldSizeByte: textFieldSizeByte,
		ImageEncodingByte: imageEncodingByte,
		ImageSizeByte:     imageSizeByte,
	}
	return
}
