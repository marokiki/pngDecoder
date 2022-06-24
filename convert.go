package main

import (
	"encoding/binary"
	"io"
)

// 1バイト文字をintに変換
func Byte1toint(b []byte) uint32 {
	_ = b[0]
	return uint32(b[0])
}

// 3バイト文字をintに変換
func Byte3toint(b []byte) uint32 {
	_ = b[2]
	return uint32(b[2]) | uint32(b[1])<<8 | uint32(b[0])<<16
}

// バイト列中の先頭nバイトを読む
func ReadBytes(r io.Reader, n int) []byte {
	buf := make([]byte, n)
	_, err := r.Read(buf)
	if err != nil {
		return nil
	}
	return buf
}

// バイト列中の先頭nバイトをintとして読む
func ReadBytesAsInt(r io.Reader, n int) int {
	if n == 4 {
		return int(binary.BigEndian.Uint32(ReadBytes(r, n)))
	} else if n == 1 {
		return int(Byte1toint(ReadBytes(r, n)))
	} else if n == 2 {
		return int(binary.BigEndian.Uint16(ReadBytes(r, n)))
	} else if n == 3 {
		return int(Byte3toint(ReadBytes(r, n)))
	} else {
		return 0
	}
}
