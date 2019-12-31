package main

import "goh264/Stream2File"

func main() {
	inPath := "/Users/sweetop/Code/mystudy/golang/gopes/source/6.h264"
	outPath := "/Users/sweetop/Code/mystudy/golang/gopes/source/7.h264"

	Stream2File.Stream2File(inPath, outPath)
}
