package test

import (
	"goh264/Stream2File"
	"goh264/ps2h264"
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

	Stream2File.Stream2File(inPath, outPath)
}
