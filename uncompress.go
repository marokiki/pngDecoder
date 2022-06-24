package main

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"math"
)

// zlib圧縮の展開
func Uncompress(data []byte) ([]byte, error) {
	dataBuffer := bytes.NewReader(data)
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
func BitsPerPixel(colorType int, depth int) (int, error) {
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

	return imageData, nil
}
