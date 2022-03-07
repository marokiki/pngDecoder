package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"math"
	"os"
)

// 1バイト文字をintに変換
func byte1toint(b []byte) uint32 {
	_ = b[0]
	return uint32(b[0])
}

// 3バイト文字をintに変換
func byte3toint(b []byte) uint32 {
	_ = b[2]
	return uint32(b[2]) | uint32(b[1])<<8 | uint32(b[0])<<16
}

// バイト列中の先頭nバイトを読む
func readBytes(r io.Reader, n int) []byte {
	buf := make([]byte, n)
	_, err := r.Read(buf)
	if err != nil {
		return nil
	}
	return buf
}

// バイト列中の先頭nバイトをintとして読む
func readBytesAsInt(r io.Reader, n int) int {
	if n == 4 {
		return int(binary.BigEndian.Uint32(readBytes(r, n)))
	} else if n == 1 {
		return int(byte1toint(readBytes(r, n)))
	} else if n == 2 {
		return int(binary.BigEndian.Uint16(readBytes(r, n)))
	} else if n == 3 {
		return int(byte3toint(readBytes(r, n)))
	} else {
		return 0
	}
}

// zlib圧縮の展開
func uncompress(data []byte) ([]byte, error) {
	dataBuffer := bytes.NewReader(data)
	fmt.Println(len(data))
	r, err := zlib.NewReader(dataBuffer)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// 1ピクセル中のビット数を計算
func bitsPerPixel(colorType int, depth int) (int, error) {
	switch colorType {
	case 0:
		return depth, nil
	case 2:
		return depth * 3, nil
	case 3:
		return depth, nil
	case 4:
		return depth * 2, nil
	case 6:
		return depth * 4, nil
	default:
		return 0, fmt.Errorf("unknown color type")
	}
}

// フィルタリング処理を展開
func applyFilter(data []byte, width, height, bitsPerPixel, bytesPerPixel int) ([]byte, error) {
	rowSize := 1 + (bitsPerPixel*width+7)/8
	imageData := make([]byte, width*height*bytesPerPixel)
	rowData := make([]byte, rowSize)
	prevRowData := make([]byte, rowSize)
	for y := 0; y < height; y++ {
		offset := y * rowSize
		rowData = data[offset : offset+rowSize]
		filterType := int(rowData[0])

		currentScanData := rowData[1:]
		prevScanData := prevRowData[1:]

		switch filterType {
		case 0:
			// No-op.
		case 1:
			for i := bytesPerPixel; i < len(currentScanData); i++ {
				currentScanData[i] += currentScanData[i-bytesPerPixel]
			}
		case 2:
			for i, p := range prevScanData {
				currentScanData[i] += p
			}
		case 3:
			for i := 0; i < bytesPerPixel; i++ {
				currentScanData[i] += prevScanData[i] / 2
			}
			for i := bytesPerPixel; i < len(currentScanData); i++ {
				currentScanData[i] += uint8((int(currentScanData[i-bytesPerPixel]) + int(prevScanData[i])) / 2)
			}
		case 4:
			var a, b, c, pa, pb, pc int
			for i := 0; i < bytesPerPixel; i++ {
				a, c = 0, 0
				for j := i; j < len(currentScanData); j += bytesPerPixel {
					b = int(prevScanData[j])
					pa = b - c
					pb = a - c
					pc = int(math.Abs(float64(pa + pb)))
					pa = int(math.Abs(float64(pa)))
					pb = int(math.Abs(float64(pb)))
					if pa <= pb && pa <= pc {
						// No-op.
					} else if pb <= pc {
						a = b
					} else {
						a = c
					}
					a += int(currentScanData[j])
					a &= 0xff
					currentScanData[j] = uint8(a)
					c = b
				}
			}
		default:
			return nil, fmt.Errorf("bad filter type")
		}

		copy(imageData[y*len(currentScanData):], currentScanData)

		prevRowData = rowData
		rowData = prevRowData
	}
	//fmt.Println(rowData)

	return imageData, nil
}

func main() {
	fileName := os.Args[1]
	buf, err := ioutil.ReadFile(fileName)
	if err != nil {
		return
	}
	r := bytes.NewReader(buf)

	if !bytes.Equal(readBytes(r, 8), []byte("\x89PNG\r\n\x1a\n")) {
		fmt.Println("This file is not PNG")
		return
	}

	loop := true
	var colorType int
	var width int
	var height int
	var depth int
	var img image.Image
	imgdata := make([]byte, 0)

	for loop {
		Length := readBytesAsInt(r, 4)
		Type := string(readBytes(r, 4))
		data := readBytes(r, Length)
		_ = readBytes(r, 4)
		fmt.Println("Chunk:", Type)

		switch Type {
		case "IHDR":
			ihdrNR := bytes.NewReader(data)
			width = readBytesAsInt(ihdrNR, 4)
			height = readBytesAsInt(ihdrNR, 4)
			depth = readBytesAsInt(ihdrNR, 1)
			colorType = readBytesAsInt(ihdrNR, 1)
			compression := readBytesAsInt(ihdrNR, 1)
			filter := readBytesAsInt(ihdrNR, 1)
			interlace := readBytesAsInt(ihdrNR, 1)
			fmt.Println("Width:", width, "Height:", height, "depth:", depth, "ColorType:", colorType, "Compression:", compression, "FilterType:", filter, "Interlace:", interlace)

		case "PLTE":
			plteNR := bytes.NewReader(data)
			paletteData := readBytes(plteNR, Length)
			fmt.Println("paletteData:", paletteData)

		case "tRNS":
			trnsNR := bytes.NewReader(data)
			if colorType == 3 {
				var PaletteAlpha []byte
				for i := 0; i < Length; i++ {
					PaletteAlpha[i] = readBytes(trnsNR, 1)[0]
					fmt.Println("PaletteNo.", i, " Alpha:", PaletteAlpha[i])
				}
			} else if colorType == 0 {
				var GlayAlpha []byte
				for i := 0; i < Length/2; i++ {
					GlayAlpha[i] = readBytes(trnsNR, 2)[0]
					fmt.Println("GlayLevel.", i, " Alpha:", GlayAlpha[i])
				}
			} else if colorType == 2 {
				var TransAlphaR []byte
				var TransAlphaG []byte
				var TransAlphaB []byte
				for i := 0; i < Length/6; i++ {
					TransAlphaR[i] = readBytes(trnsNR, 2)[0]
					TransAlphaG[i] = readBytes(trnsNR, 2)[0]
					TransAlphaB[i] = readBytes(trnsNR, 2)[0]
					fmt.Println("No.", i, " Alpha R:", TransAlphaR[i], ", G:", TransAlphaG[i], ", B:", TransAlphaB[i])
				}
			}

		case "gAMA":
			gamaNR := bytes.NewReader(data)
			gamma := readBytesAsInt(gamaNR, Length)
			fmt.Println("gammaValue:", gamma)

		case "cHRM":
			chrmNR := bytes.NewReader(data)
			whitePointX := readBytesAsInt(chrmNR, 4)
			whitePointY := readBytesAsInt(chrmNR, 4)
			redX := readBytesAsInt(chrmNR, 4)
			redY := readBytesAsInt(chrmNR, 4)
			greenX := readBytesAsInt(chrmNR, 4)
			greenY := readBytesAsInt(chrmNR, 4)
			blueX := readBytesAsInt(chrmNR, 4)
			blueY := readBytesAsInt(chrmNR, 4)
			fmt.Println("White Point X:", whitePointX, "White Point Y:", whitePointY, "Red X:", redX, "Red Y:", redY, "Green X:", greenX, "Green Y:", greenY, "Blue X:", blueX, "Blue Y:", blueY)

		case "sRGB":
			rgbNR := bytes.NewReader(data)
			rendering := readBytes(rgbNR, Length)
			fmt.Println("Rendering Effects:", string(rendering))

		case "iCCP":
			iccpNR := bytes.NewReader(data)
			profile := readBytes(iccpNR, Length)
			fmt.Println("profile:", string(profile))

		case "tEXt":
			textNR := bytes.NewReader(data)
			keyWords := readBytes(textNR, Length)
			fmt.Println("KeyWords:", string(keyWords))

		case "zTXt":
			textNR := bytes.NewReader(data)
			KeyWords := readBytes(textNR, Length)
			fmt.Println("Compressed KeyWords:", string(KeyWords))

		case "iTXt":
			textNR := bytes.NewReader(data)
			KeyWords := readBytes(textNR, Length)
			fmt.Println("International KeyWords:", string(KeyWords))

		case "bKGD":
			bkgdNR := bytes.NewReader(data)
			if colorType == 3 {
				paletteNo := readBytesAsInt(bkgdNR, 1)
				fmt.Println("BackGround Palette No:", paletteNo)
			} else if colorType == 0 || colorType == 4 {
				glayLevel := readBytesAsInt(bkgdNR, 2)
				fmt.Println("BackGround Glay Level:", glayLevel)
			} else if colorType == 2 || colorType == 6 {
				r := readBytesAsInt(bkgdNR, 2)
				g := readBytesAsInt(bkgdNR, 2)
				b := readBytesAsInt(bkgdNR, 2)
				fmt.Println("BackGround Color R:", r, " G:", g, " B:", b)
			} else {
				fmt.Println("Invalid ColorType")
			}

		case "pHYs":
			pixelNR := bytes.NewReader(data)
			pixelX := readBytesAsInt(pixelNR, 4)
			pixelY := readBytesAsInt(pixelNR, 4)
			unit := readBytesAsInt(pixelNR, 1)
			u := "Undefined"
			if unit == 1 {
				u = "Meter"
			}
			fmt.Println("Pixel per unit X:", pixelX, "Y:", pixelY, "unit:", u)

		case "sBIT":
			bitNR := bytes.NewReader(data)
			if colorType == 0 {
				glaybit := readBytesAsInt(bitNR, 1)
				fmt.Println("Glay Bit Number:", glaybit)
			} else if colorType == 2 {
				r := readBytesAsInt(bitNR, 1)
				g := readBytesAsInt(bitNR, 1)
				b := readBytesAsInt(bitNR, 1)
				fmt.Println("Bit Number Red:", r, "Green:", g, "Blue:", b)
			} else if colorType == 3 {
				r := readBytesAsInt(bitNR, 1)
				g := readBytesAsInt(bitNR, 1)
				b := readBytesAsInt(bitNR, 1)
				fmt.Println("Bit Number on Palette Red:", r, "Green:", g, "Blue:", b)
			} else if colorType == 4 {
				glay := readBytesAsInt(bitNR, 1)
				alpha := readBytesAsInt(bitNR, 1)
				fmt.Println("Glay Bit Number:", glay, "Alpha Bit Number:", alpha)
			} else if colorType == 6 {
				r := readBytesAsInt(bitNR, 1)
				g := readBytesAsInt(bitNR, 1)
				b := readBytesAsInt(bitNR, 1)
				alpha := readBytesAsInt(bitNR, 1)
				fmt.Println("Bit Number Red:", r, "Green:", g, "Blue:", b, "Alpha:", alpha)
			} else {
				fmt.Println("Invalid ColorType")
			}

		case "tIME":
			timeNR := bytes.NewReader(data)
			year := readBytesAsInt(timeNR, 2)
			month := readBytesAsInt(timeNR, 1)
			day := readBytesAsInt(timeNR, 1)
			hour := readBytesAsInt(timeNR, 1)
			minute := readBytesAsInt(timeNR, 1)
			second := readBytesAsInt(timeNR, 1)
			fmt.Println("Last-Modification Time Year:", year, " Month:", month, " Day:", day, " Hour:", hour, " Minute:", minute, " Second:", second)

		case "sPLT":
			paletteNR := bytes.NewReader(data)
			paletteName := readBytes(paletteNR, Length)
			fmt.Println("Recommended Palatte:", paletteName)

		case "hIST":
			histNR := bytes.NewReader(data)
			paletteNum := Length / 3
			hist := make([]int, 0)
			for i := 0; i < paletteNum; i++ {
				hist = append(hist, readBytesAsInt(histNR, 2))
			}
			if hist[0] == 0 {
				fmt.Println("No Used")
			} else {
				for i := 0; i < paletteNum; i++ {
					fmt.Println("Palette Num ", i, " Use Frequency:", hist[i])
				}
			}

		case "IDAT":
			//idatNR := bytes.NewReader(data)
			//imgData := readBytes(idatNR, Length)
			imgdata = append(imgdata, data...)
			fmt.Println("Data Length:", len(data))

		case "IEND":
			loop = false
			fmt.Println("END")
		}
	}
	uncompressedData, err := uncompress(imgdata)
	if err != nil {
		fmt.Println("Uncompress Error")
		return
	}
	fmt.Println("Uncompressed Data Length:", len(uncompressedData))

	// apply filter type
	bitsPerPixel, err := bitsPerPixel(colorType, depth)
	if err != nil {
		return
	}
	bytesPerPixel := (bitsPerPixel + 7) / 8
	ndata, err := applyFilter(uncompressedData, width, height, bitsPerPixel, bytesPerPixel)
	if err != nil {
		return
	}
	fmt.Println("appled filter type data length:", len(ndata))

	switch colorType {
	case 0: // GlayScale Image
		nglay := image.NewGray(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				offset := bytesPerPixel*width*y + bytesPerPixel*x
				pixel := ndata[offset : offset+bytesPerPixel]
				i := y*nglay.Stride + x*1
				nglay.Pix[i] = pixel[0] // Glay
			}
		}
		img = nglay

	case 2: // True Color Image
		nrgb := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				offset := bytesPerPixel*width*y + bytesPerPixel*x
				pixel := ndata[offset : offset+bytesPerPixel]
				i := y*nrgb.Stride + x*4
				nrgb.Pix[i] = pixel[0]   // R
				nrgb.Pix[i+1] = pixel[1] // G
				nrgb.Pix[i+2] = pixel[2] // B
				nrgb.Pix[i+3] = 255      // Alpha
			}
		}
		img = nrgb

	// TO DO: case 3(インデックスカラー画像)

	// TO DO: case 4 image.NewGray 修正
	case 4: // GlayScale + Alpha Image
		nglayalpha := image.NewGray(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				offset := bytesPerPixel*width*y + bytesPerPixel*x
				pixel := ndata[offset : offset+bytesPerPixel]
				i := y*nglayalpha.Stride + x*2
				nglayalpha.Pix[i] = pixel[0] // Glay
				nglayalpha.Pix[i+1] = 255    // Alpha
			}
		}
		img = nglayalpha

	case 6: // True Color + Alpha Image
		nrgba := image.NewRGBA(image.Rect(0, 0, width, height))
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				offset := bytesPerPixel*width*y + bytesPerPixel*x
				pixel := ndata[offset : offset+bytesPerPixel]
				i := y*nrgba.Stride + x*4
				nrgba.Pix[i] = pixel[0]   // R
				nrgba.Pix[i+1] = pixel[1] // G
				nrgba.Pix[i+2] = pixel[2] // B
				nrgba.Pix[i+3] = pixel[3] // A
			}
		}
		img = nrgba
	}

	outputFile, err := os.Create(("output.png"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outputFile.Close()

	png.Encode(outputFile, img)

	fmt.Println("Complete")
}
