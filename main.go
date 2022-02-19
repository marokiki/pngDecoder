package main

import (
	"encoding/binary"
	"io"
	"bytes"
	"fmt"
	"io/ioutil"
)

func readBytes(r io.Reader, n int) []byte {
	buf := make([]byte, n)
	_, err := r.Read(buf)
	if err != nil{
		return nil
	}
	return buf
}

func readBytesAsInt(r io.Reader, n int) int {
	if n == 4{
	return int(binary.BigEndian.Uint32(readBytes(r, n)))
	} else {
		return int(readBytes(r,n)[0])
	}
}

func main(){
	buf, err := ioutil.ReadFile("test.png")
	if err != nil {
		return 
	}
	r := bytes.NewReader(buf)

	if !bytes.Equal(readBytes(r, 8), []byte("\x89PNG\r\n\x1a\n")) {
		fmt.Println("This file is not PNG")
		return
	}

	loop := true

	for loop{
		Length := readBytesAsInt(r,4)
		Type := string(readBytes(r, 4))
		data := readBytes(r, Length)
		_ = readBytes(r, 4)
		fmt.Println("Chunk:",Type)

		// TO DO: 必須チャンク以外の追加
		switch Type {
		case "IHDR":
			ihdrNR := bytes.NewReader(data)
			width := readBytesAsInt(ihdrNR, 4)
			height := readBytesAsInt(ihdrNR, 4)
			depth := readBytesAsInt(ihdrNR, 1)
			colorType := readBytesAsInt(ihdrNR, 1)
			compression := readBytesAsInt(ihdrNR, 1)
			filter := readBytesAsInt(ihdrNR, 1)
			interlace := readBytesAsInt(ihdrNR, 1)
			fmt.Println("Width:",width,"Height:",height,"depth:",depth,"ColorType:",colorType,"Compression:",compression,"FilterType:",filter,"Interlace:",interlace)
		
		// TO DO: Data部の展開
		case "IDAT":
			idatNR := bytes.NewReader(data)
			imgData := readBytes(idatNR, Length)
			fmt.Println("imageData:",imgData)
		
		case "IEND":
			loop = false
		}
	}
	fmt.Println("Complete")
}