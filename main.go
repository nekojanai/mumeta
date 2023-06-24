package main

import (
	"fmt"
	"nekojanai/mumeta/id3v2"
	"os"
)

func main() {
	f, error := os.Open("test_data/id3v2-xheader.mp3")
	defer f.Close()
	if error != nil {
		fmt.Println(error)
		panic(error)
	}
	defer f.Close()

	h, e := id3v2.ReadID3v2Header(f)
	if e != nil {
		panic(e)
	}
	fmt.Printf("%+v\n", h)
	eh, err := id3v2.ReadID3v2ExtendedHeader(f, h)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v\n", eh)
	fmt.Printf("%+v\n", eh.ParseExtendedFlags())
}
