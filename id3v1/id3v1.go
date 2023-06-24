package id3v1

import (
	"fmt"
	"os"
)

const (
	id3v1TagSize = 128

	id3v1HeaderSize  = 3
	id3v1TitleSize   = 30
	id3v1ArtistSize  = 30
	id3v1AlbumSize   = 30
	id3v1YearSize    = 4
	id3v1CommentSize = 30
	id3v1GenreSize   = 1

	id3v11CommentSize = 28
	id3v11TrackSize   = 2
)

var id3v1HeaderValue = [3]byte{0x54, 0x41, 0x47} // TAG

type ID3v1Tag struct {
	title  [30]byte
	artist [30]byte
	album  [30]byte
	year   [4]byte
	// [28]byte for ID3v1.1, [30]byte for ID3v1
	comment []byte
	// only for ID3v1.1, first byte always 0
	track [2]byte
	genre byte
}

func ReadID3v1Tag(f *os.File) (tag ID3v1Tag, err error) {
	b := make([]byte, id3v1HeaderSize)
	fileInfo, err := f.Stat()
	if err != nil {
		return
	}
	n, err := f.ReadAt(b, fileInfo.Size()-id3v1TagSize)
	if err != nil || n != id3v1HeaderSize {
		return
	}
	if b[0] != id3v1HeaderValue[0] ||
		b[1] != id3v1HeaderValue[1] ||
		b[2] != id3v1HeaderValue[2] {
		err = fmt.Errorf("No ID3v1 tag")
		return
	}

	offset := fileInfo.Size() - id3v1TagSize + id3v1HeaderSize
	if err != nil {
		return
	}
	title := make([]byte, id3v1TitleSize)
	n, err = f.ReadAt(title, offset)
	if err != nil || n != id3v1TitleSize {
		return
	}
	offset += id3v1TitleSize
	artist := make([]byte, id3v1ArtistSize)
	n, err = f.ReadAt(artist, offset)
	if err != nil || n != id3v1ArtistSize {
		return
	}
	offset += id3v1ArtistSize
	album := make([]byte, id3v1AlbumSize)
	n, err = f.ReadAt(album, offset)
	if err != nil || n != id3v1AlbumSize {
		return
	}
	offset += id3v1AlbumSize
	year := make([]byte, id3v1YearSize)
	n, err = f.ReadAt(year, offset)
	if err != nil || n != id3v1YearSize {
		return
	}
	offset += id3v1YearSize

	// Check if ID3v1.1
	nextTwoBytes := make([]byte, 2)
	n, err = f.ReadAt(nextTwoBytes, offset+id3v11CommentSize)
	if err != nil || n != 2 {
		return
	}
	var comment, track []byte

	if nextTwoBytes[0] == 0 && nextTwoBytes[1] != 0 {
		comment = make([]byte, id3v11CommentSize)
		n, err = f.ReadAt(comment, offset)
		if err != nil || n != id3v11CommentSize {
			return
		}
		offset += id3v11CommentSize
		// The first byte of the track number is always 0
		track = make([]byte, id3v11TrackSize)
		n, err = f.ReadAt(track, offset)
		if err != nil || n != id3v11TrackSize {
			return
		}
		offset += id3v11TrackSize
	} else {
		comment = make([]byte, id3v1CommentSize)
		n, err = f.ReadAt(comment, offset)
		if err != nil || n != id3v1CommentSize {
			return
		}
		offset += id3v1CommentSize
	}
	genre := make([]byte, id3v1GenreSize)
	n, err = f.ReadAt(genre, offset)
	if err != nil || n != id3v1GenreSize {
		return
	}
	tag = ID3v1Tag{
		title:   [30]byte(title),
		artist:  [30]byte(artist),
		album:   [30]byte(album),
		year:    [4]byte(year),
		comment: comment,
		track:   [2]byte(track),
		genre:   genre[0],
	}
	return
}

func (t ID3v1Tag) String() string {
	if t.isEmpty() {
		return "{}"
	}
	version := fmt.Sprintf("------------------ %s ------------------\n", t.version())
	title := fmt.Sprintf("title:  \t%s\n", t.title)
	artist := fmt.Sprintf("artist: \t%s\n", t.artist)
	album := fmt.Sprintf("album:  \t%s\n", t.album)
	year := fmt.Sprintf("year:   \t%s\n", t.year)
	var track string
	if t.version() == "ID3v1.1" {
		track = fmt.Sprintf("track:  \t%d\n", t.track[1])
	}
	genre := fmt.Sprintf("genre:  \t%s\n", t.genreString())
	comment := fmt.Sprintf("comment:\t%s\n", t.comment)
	footer := "---------------------------------------------"
	return fmt.Sprintf("%s%s%s%s%s%s%s%s%s", version, title, artist, album, year, track, genre, comment, footer)
}

func (t ID3v1Tag) isEmpty() bool {
	if t.title == [30]byte{} &&
		t.artist == [30]byte{} &&
		t.album == [30]byte{} &&
		t.year == [4]byte{} &&
		t.comment == nil &&
		t.track == [2]byte{} &&
		t.genre == 0 {
		return true
	}
	return false
}

func (t ID3v1Tag) version() string {
	if t.track[0] == 0 && t.track[1] != 0 {
		return "ID3v1.1"
	}
	return "ID3v1"
}

func (t ID3v1Tag) genreString() string {
	genre, ok := genres[t.genre]
	if !ok {
		return "Unknown"
	}
	return genre
}
