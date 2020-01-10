package test

import (
	"goh264/flv2h264"
	"goh264/h2642flv"
	"goh264/ps2h264"
	"goh264/stream2file"
	"goh264/ts2h264"
	"testing"
)

func TestPs2H264(t *testing.T) {
	inPath := "source/6.ps"
	outPath := "source/6.h264"

	ps2h264.Transfer(inPath, outPath)
}

func TestStream2File(t *testing.T) {
	inPath := "source/6.h264"
	outPath := "source/7.h264"

	stream2file.Transfer(inPath, outPath)
}

func TestH2642flv(t *testing.T) {
	inPath := "source/test2.h264"
	outPath := "source/test2.flv"

	h2642flv.Transfer(inPath, outPath)
}

func TestFlv2h264(t *testing.T) {
	inPath := "source/receive.flv"
	outPath := "source/receive.h264"

	flv2h264.Transfer(inPath, outPath)
}

func TestTS2H264(t *testing.T) {
	inPath := "source/test3.ts"
	outPath := "source/test3.h264"

	ts2h264.Transfer(inPath, outPath)

}
