package flv2h264

import (
	"bufio"
	"log"
	"os"
)

func Flv2h264(inPath, outPath string) {
	fi, err := os.Open(inPath)
	if err != nil {
		log.Println("open error : " + err.Error())
	}
	defer fi.Close()
	r := bufio.NewReader(fi)

	for {
		buf := make([]byte, 20480)
		n, err := r.Read(buf)
		if err != nil && err.Error() != "EOF" {
			log.Println("read error : " + err.Error())
		}
		if 0 == n {
			break
		}
		log.Println(buf[:9])
		log.Println(buf[9:13])
		taglen := int(buf[14])<<16 + int(buf[15])<<8 + int(buf[16])
		log.Println("taglen", taglen)
		buf[4] = 0x01
		log.Println(buf[:25+taglen])
		return
	}
}
