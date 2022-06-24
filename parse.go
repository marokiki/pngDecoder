package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
)

func Parse(buffer []byte) (err error) {
	r := bytes.NewReader(buffer)

	if !bytes.Equal(ReadBytes(r, 8), []byte("\x89PNG\r\n\x1a\n")) {
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
		Length := ReadBytesAsInt(r, 4)
		Type := string(ReadBytes(r, 4))
		data := ReadBytes(r, Length)
		_ = ReadBytes(r, 4)
		fmt.Println("Chunk:", Type)

		switch Type {
		case "IHDR":
			ihdrNR := bytes.NewReader(data)
			width = ReadBytesAsInt(ihdrNR, 4)
			height = ReadBytesAsInt(ihdrNR, 4)
			depth = ReadBytesAsInt(ihdrNR, 1)
			colorType = ReadBytesAsInt(ihdrNR, 1)
			compression := ReadBytesAsInt(ihdrNR, 1)
			filter := ReadBytesAsInt(ihdrNR, 1)
			interlace := ReadBytesAsInt(ihdrNR, 1)
			fmt.Println("Width:", width, "Height:", height, "depth:", depth, "ColorType:", colorType, "Compression:", compression, "FilterType:", filter, "Interlace:", interlace)

		case "PLTE":
			plteNR := bytes.NewReader(data)
			paletteData := ReadBytes(plteNR, Length)
			fmt.Println("paletteData:", paletteData)

		case "tRNS":
			trnsNR := bytes.NewReader(data)
			if colorType == 3 {
				var PaletteAlpha []byte
				for i := 0; i < Length; i++ {
					PaletteAlpha[i] = ReadBytes(trnsNR, 1)[0]
					fmt.Println("PaletteNo.", i, " Alpha:", PaletteAlpha[i])
				}
			} else if colorType == 0 {
				var GlayAlpha []byte
				for i := 0; i < Length/2; i++ {
					GlayAlpha[i] = ReadBytes(trnsNR, 2)[0]
					fmt.Println("GlayLevel.", i, " Alpha:", GlayAlpha[i])
				}
			} else if colorType == 2 {
				var TransAlphaR []byte
				var TransAlphaG []byte
				var TransAlphaB []byte
				for i := 0; i < Length/6; i++ {
					TransAlphaR[i] = ReadBytes(trnsNR, 2)[0]
					TransAlphaG[i] = ReadBytes(trnsNR, 2)[0]
					TransAlphaB[i] = ReadBytes(trnsNR, 2)[0]
					fmt.Println("No.", i, " Alpha R:", TransAlphaR[i], ", G:", TransAlphaG[i], ", B:", TransAlphaB[i])
				}
			}

		case "gAMA":
			gamaNR := bytes.NewReader(data)
			gamma := ReadBytesAsInt(gamaNR, Length)
			fmt.Println("gammaValue:", gamma)

		case "cHRM":
			chrmNR := bytes.NewReader(data)
			whitePointX := ReadBytesAsInt(chrmNR, 4)
			whitePointY := ReadBytesAsInt(chrmNR, 4)
			redX := ReadBytesAsInt(chrmNR, 4)
			redY := ReadBytesAsInt(chrmNR, 4)
			greenX := ReadBytesAsInt(chrmNR, 4)
			greenY := ReadBytesAsInt(chrmNR, 4)
			blueX := ReadBytesAsInt(chrmNR, 4)
			blueY := ReadBytesAsInt(chrmNR, 4)
			fmt.Println("White Point X:", whitePointX, "White Point Y:", whitePointY, "Red X:", redX, "Red Y:", redY, "Green X:", greenX, "Green Y:", greenY, "Blue X:", blueX, "Blue Y:", blueY)

		case "sRGB":
			rgbNR := bytes.NewReader(data)
			rendering := ReadBytes(rgbNR, Length)
			fmt.Println("Rendering Effects:", string(rendering))

		case "iCCP":
			iccpNR := bytes.NewReader(data)
			profile := ReadBytes(iccpNR, Length)
			fmt.Println("profile:", string(profile))

		case "tEXt":
			textNR := bytes.NewReader(data)
			keyWords := ReadBytes(textNR, Length)
			fmt.Println("KeyWords:", string(keyWords))

		case "zTXt":
			textNR := bytes.NewReader(data)
			KeyWords := ReadBytes(textNR, Length)
			fmt.Println("Compressed KeyWords:", string(KeyWords))

		case "iTXt":
			textNR := bytes.NewReader(data)
			KeyWords := ReadBytes(textNR, Length)
			fmt.Println("International KeyWords:", string(KeyWords))

		case "bKGD":
			bkgdNR := bytes.NewReader(data)
			if colorType == 3 {
				paletteNo := ReadBytesAsInt(bkgdNR, 1)
				fmt.Println("BackGround Palette No:", paletteNo)
			} else if colorType == 0 || colorType == 4 {
				glayLevel := ReadBytesAsInt(bkgdNR, 2)
				fmt.Println("BackGround Glay Level:", glayLevel)
			} else if colorType == 2 || colorType == 6 {
				r := ReadBytesAsInt(bkgdNR, 2)
				g := ReadBytesAsInt(bkgdNR, 2)
				b := ReadBytesAsInt(bkgdNR, 2)
				fmt.Println("BackGround Color R:", r, " G:", g, " B:", b)
			} else {
				fmt.Println("Invalid ColorType")
			}

		case "pHYs":
			pixelNR := bytes.NewReader(data)
			pixelX := ReadBytesAsInt(pixelNR, 4)
			pixelY := ReadBytesAsInt(pixelNR, 4)
			unit := ReadBytesAsInt(pixelNR, 1)
			u := "Undefined"
			if unit == 1 {
				u = "Meter"
			}
			fmt.Println("Pixel per unit X:", pixelX, "Y:", pixelY, "unit:", u)

		case "sBIT":
			bitNR := bytes.NewReader(data)
			if colorType == 0 {
				glaybit := ReadBytesAsInt(bitNR, 1)
				fmt.Println("Glay Bit Number:", glaybit)
			} else if colorType == 2 {
				r := ReadBytesAsInt(bitNR, 1)
				g := ReadBytesAsInt(bitNR, 1)
				b := ReadBytesAsInt(bitNR, 1)
				fmt.Println("Bit Number Red:", r, "Green:", g, "Blue:", b)
			} else if colorType == 3 {
				r := ReadBytesAsInt(bitNR, 1)
				g := ReadBytesAsInt(bitNR, 1)
				b := ReadBytesAsInt(bitNR, 1)
				fmt.Println("Bit Number on Palette Red:", r, "Green:", g, "Blue:", b)
			} else if colorType == 4 {
				glay := ReadBytesAsInt(bitNR, 1)
				alpha := ReadBytesAsInt(bitNR, 1)
				fmt.Println("Glay Bit Number:", glay, "Alpha Bit Number:", alpha)
			} else if colorType == 6 {
				r := ReadBytesAsInt(bitNR, 1)
				g := ReadBytesAsInt(bitNR, 1)
				b := ReadBytesAsInt(bitNR, 1)
				alpha := ReadBytesAsInt(bitNR, 1)
				fmt.Println("Bit Number Red:", r, "Green:", g, "Blue:", b, "Alpha:", alpha)
			} else {
				fmt.Println("Invalid ColorType")
			}

		case "tIME":
			timeNR := bytes.NewReader(data)
			year := ReadBytesAsInt(timeNR, 2)
			month := ReadBytesAsInt(timeNR, 1)
			day := ReadBytesAsInt(timeNR, 1)
			hour := ReadBytesAsInt(timeNR, 1)
			minute := ReadBytesAsInt(timeNR, 1)
			second := ReadBytesAsInt(timeNR, 1)
			fmt.Println("Last-Modification Time Year:", year, " Month:", month, " Day:", day, " Hour:", hour, " Minute:", minute, " Second:", second)

		case "sPLT":
			paletteNR := bytes.NewReader(data)
			paletteName := ReadBytes(paletteNR, Length)
			fmt.Println("Recommended Palatte:", paletteName)

		case "hIST":
			histNR := bytes.NewReader(data)
			paletteNum := Length / 3
			hist := make([]int, 0)
			for i := 0; i < paletteNum; i++ {
				hist = append(hist, ReadBytesAsInt(histNR, 2))
			}
			if hist[0] == 0 {
				fmt.Println("No Used")
			} else {
				for i := 0; i < paletteNum; i++ {
					fmt.Println("Palette Num ", i, " Use Frequency:", hist[i])
				}
			}

		case "IDAT":
			imgdata = append(imgdata, data...)
			fmt.Println("Data Length:", len(data))

		case "IEND":
			loop = false
			fmt.Println("END")
		}
	}
	UncompressedData, err := Uncompress(imgdata)
	if err != nil {
		fmt.Println("Uncompress Error")
		return
	}
	fmt.Println("Uncompressed Data Length:", len(UncompressedData))

	// apply filter type
	bitsPerPixel, err := BitsPerPixel(colorType, depth)
	if err != nil {
		return
	}
	bytesPerPixel := (bitsPerPixel + 7) / 8
	ndata, err := applyFilter(UncompressedData, width, height, bitsPerPixel, bytesPerPixel)
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
	outputFile, err := os.Create("output.png")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer outputFile.Close()

	png.Encode(outputFile, img)

	fmt.Println("Complete")

	return err
}
