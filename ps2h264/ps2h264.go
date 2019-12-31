package ps2h264

import (
	"bufio"
	"log"
	"os"
	"sync"
)

var (
	bufChan        = make(chan []byte, 10)
	pspkgChan      = make(chan []byte, 10)
	closeSplitChan = false
	closeWriteChan = false
	wg             = sync.WaitGroup{}
)

func Ps2H264(inPath, outPath string) {
	fi, err := os.Open(inPath)
	if err != nil {
		log.Println("open error : " + err.Error())
	}
	defer fi.Close()
	r := bufio.NewReader(fi)

	go splitPS()
	go writePES(outPath)

	for {
		buf := make([]byte, 20480)
		n, err := r.Read(buf)
		if err != nil && err.Error() != "EOF" {
			log.Println("read error : " + err.Error())
		}
		if 0 == n {
			break
		}
		bufChan <- buf[:n]
	}
	closeSplitChan = true
	wg.Wait()
}

func splitPS() {
	temp := make([]byte, 0)
	pspkg := make([]byte, 0)
	pspkgNum := 0

	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case buf := <-bufChan:
			buf = append(temp, buf...)
			temp = make([]byte, 0)
			i := 0
			for i < len(buf) {
				if i > len(buf)-4 {
					temp = buf[i:] //剩余不足4个字节的数据
					break
				}
				if buf[i] == 0x00 && buf[i+1] == 0x00 && buf[i+2] == 0x01 && buf[i+3] == 0xba {
					if i != 0 {
						pspkgChan <- pspkg
						pspkg = make([]byte, 0)
					}
					pspkgNum++ //frame数量+1
					i += 4
				} else {
					pspkg = append(pspkg, buf[i])
					i++
				}
			}
			break
		default:
			if closeSplitChan {
				pspkg = append(pspkg, temp...)
				pspkgNum++
				pspkgChan <- pspkg
				closeWriteChan = true
				log.Println("关闭切割split协程")
				log.Println("ps包数量：", pspkgNum)
				return
			}
		}

	}
}

func writePES(outPath string) {
	fi, err := os.Create(outPath)
	if err != nil {
		log.Println("open error : " + err.Error())
	}
	defer func() {
		fi.Close()
		wg.Done()
	}()
	wg.Add(1)

	w := bufio.NewWriter(fi)
	start := false

	for {
		select {
		case pspkg := <-pspkgChan:
			extlen := pspkg[9] & 0x07

			log.Println("extlen", extlen)
			psSystemHeaderLenght := byte(0)
			programStreamMapLength := byte(0)
			pespkg := []byte{}

			if pspkg[10+extlen] == 0x00 && pspkg[11+extlen] == 0x00 && pspkg[12+extlen] == 0x01 && pspkg[13+extlen] == 0xBB {
				psSystemHeaderLenght = pspkg[14+extlen]<<8 + pspkg[15+extlen]
			}
			if pspkg[16+extlen+psSystemHeaderLenght] == 0x00 && pspkg[17+extlen+psSystemHeaderLenght] == 0x00 &&
				pspkg[18+extlen+psSystemHeaderLenght] == 0x01 && pspkg[19+extlen+psSystemHeaderLenght] == 0xBC {
				programStreamMapLength = pspkg[20+extlen+psSystemHeaderLenght]<<8 + pspkg[21+extlen+psSystemHeaderLenght]
				pespkg = pspkg[22+extlen+psSystemHeaderLenght+programStreamMapLength:]
			} else {
				pespkg = pspkg[10+extlen:]
			}

			log.Println("len(pespkg)", len(pespkg))
			for i := 0; i < len(pespkg); {
				peslen := int(pespkg[i+4])<<8 + int(pespkg[i+5])
				pesheaderlength := int(pespkg[i+8])

				if i+9+pesheaderlength > len(pespkg) {
					log.Println("===========error PS  1=============")
					break
				}
				if i+9+pesheaderlength > i+6+peslen {
					log.Println("===========error PS  2=============")
					break
				}
				if i+6+peslen > len(pespkg) {
					log.Println("===========error PS  3=============")
					break
				} else {
					if start || pespkg[i+13+pesheaderlength] == 0x67 {
						w.Write(pespkg[i+9+pesheaderlength : i+6+peslen])
						start = true
					}

				}
				i = i + 6 + peslen
			}
			break
		default:
			if closeWriteChan {
				log.Println("关闭写入pes协程")
				return
			}
		}
	}
}
