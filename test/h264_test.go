package test

import (
	"goh264/flv2h264"
	"goh264/h2642flv"
	"goh264/ps2h264"
	"goh264/stream2file"
	"testing"
)

func TestPs2H264(t *testing.T) {
	inPath := "source/6.ps"
	outPath := "source/6.h264"

	ps2h264.Ps2H264(inPath, outPath)
}

func TestStream2File(t *testing.T) {
	inPath := "source/6.h264"
	outPath := "source/7.h264"

	stream2file.Stream2File(inPath, outPath)
}

func TestH2642flv(t *testing.T) {
	inPath := "source/test2.h264"
	outPath := "source/test2.flv"

	h2642flv.H2642flv(inPath, outPath)
}

func TestFlv2h264(t *testing.T) {
	inPath := "source/receive.flv"
	outPath := "source/receive.h264"

	flv2h264.Flv2h264(inPath, outPath)
}
