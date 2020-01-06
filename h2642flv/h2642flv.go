package h2642flv

import (
	"bufio"
	"goh264/enum/NalUnitTypes"
	"goh264/model"
	"log"
	"os"
	"sync"
)

var (
	bufChan         = make(chan []byte, 10)
	frameChan       = make(chan []byte, 10)
	closeSplitChan  = false
	closeDecodeChan = false
	wg              = sync.WaitGroup{}

	prelen    = 304
	timestamp = 0
)

func H2642flv(inPath, outPath string) {
	fi, err := os.Open(inPath)
	if err != nil {
		log.Println("open error : " + err.Error())
	}
	defer fi.Close()
	r := bufio.NewReader(fi)

	go splitFrame()
	go decodeFrame(outPath)

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

func splitFrame() {
	temp := make([]byte, 0)
	frame := make([]byte, 0)
	frameNum := 0

	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	for {
		select {
		case buf := <-bufChan:
			buf = append(temp, buf...) //头部追加上个包剩余不满4个字节的数据
			temp = make([]byte, 0)
			i := 0
			for i < len(buf) {
				if i > len(buf)-4 {
					temp = buf[i:] //剩余不足4个字节的数据
					break
				}

				if buf[i] == 0x00 && buf[i+1] == 0x00 && buf[i+2] == 0x00 && buf[i+3] == 0x01 {
					if i != 0 {
						frameChan <- frame
						frame = make([]byte, 0)
					}
					frameNum++ //frame数量+1
					i += 4
				} else if buf[i] == 0x00 && buf[i+1] == 0x00 && buf[i+2] == 0x01 {
					if i != 0 {
						frameChan <- frame
						frame = make([]byte, 0)
					}
					frameNum++ //frame数量+1
					i += 3
				} else {
					frame = append(frame, buf[i])
					i++
				}
			}
			break
		default:
			if closeSplitChan {
				frame = append(frame, temp...)
				frameChan <- frame
				closeDecodeChan = true
				log.Println("关闭切割split协程")
				log.Println("frame数量：", frameNum)
				return
			}

		}
	}
}

func decodeFrame(outPath string) {
	fi, err := os.Create(outPath)
	if err != nil {
		log.Println("open error : " + err.Error())
	}
	w := bufio.NewWriter(fi)
	defer func() {
		log.Println("exit")
		w.Flush()
		fi.Close()
		wg.Done()
	}()

	wg.Add(1)

	first := true

	w.Write([]byte{0x46, 0x4c, 0x56, 0x01, 0x01, 0x00, 0x00, 0x00, 0x09})
	//w.Write(getHeader())
	i := 0
	tempsps := []byte{}
	for {
		select {
		case frame := <-frameChan:

			flen := len(frame)
			i++
			log.Println("帧", i, "的数据长度：", flen)
			dataStream := model.StreamData{Data: frame, Index: 0}
			_ = dataStream.F(1)
			//log.Printf("forbidden_bit %d\n", ForbiddenBit)
			_ = dataStream.U(2)
			//log.Printf("nal_reference_bit %d\n", NalReferenceBit)
			NalUnitType := dataStream.U(5)
			//log.Printf("nal_unit_type %d\n", NalUnitType)

			switch NalUnitType {
			case NalUnitTypes.SPS:
				tempsps = frame
				log.Println("SPS")
				break
			case NalUnitTypes.PPS:
				log.Println("PPS")
				if first {
					flvspspps(w, tempsps, frame)
					first = false
				}
				break
			case NalUnitTypes.SEI:
				log.Println("SEI")
				break
			case NalUnitTypes.IDR:
				log.Println("IDR")
				flvnalu(w, 0x17, frame)
				break
			case NalUnitTypes.NOIDR:
				log.Println("NOIDR")
				flvnalu(w, 0x27, frame)
				break
			default:
				break
			}
			break
		default:
			if closeDecodeChan {
				log.Println("关闭切割decode协程")
				return
			}
		}
	}
}

func flvnalu(w *bufio.Writer, fc byte, nalu []byte) {
	nlen := len(nalu)
	dlen := 5 + nlen + 4

	w.Write([]byte{byte(prelen >> 24), byte(prelen >> 16), byte(prelen >> 8), byte(prelen)})
	w.Write([]byte{0x09, //type
		byte(dlen >> 16), byte(dlen >> 8), byte(dlen), //tag data len
		byte(timestamp >> 16), byte(timestamp >> 8), byte(timestamp), byte(timestamp >> 24), //timestamp
		0x00, 0x00, 0x00}) //stream iD
	w.Write([]byte{fc,
		0x01,              //AVCPacketType
		0x00, 0x00, 0x00}) //Composition Time

	//i := 0
	//for ; i+step < nlen; i += step {
	//	w.Write([]byte{byte(step >> 24), byte(step >> 16), byte(step >> 8), byte(step)})
	//	w.Write(nalu[:i+step])
	//}
	//temp := nlen - i
	//w.Write([]byte{byte(temp >> 24), byte(temp >> 16), byte(temp >> 8), byte(temp)})
	//w.Write(nalu[i:])

	w.Write([]byte{byte(nlen >> 24), byte(nlen >> 16), byte(nlen >> 8), byte(nlen)})
	w.Write(nalu)
	timestamp += 40
	prelen = dlen + 11
}

func flvspspps(w *bufio.Writer, sps []byte, pps []byte) {
	slen := len(sps)
	plen := len(pps)
	dlen := 5 + slen + 8 + plen + 3

	if prelen == 398 {
		panic(0)
	}

	w.Write([]byte{byte(prelen >> 24), byte(prelen >> 16), byte(prelen >> 8), byte(prelen)})
	w.Write([]byte{0x09, //type
		byte(dlen >> 16), byte(dlen >> 8), byte(dlen), //tag data len
		byte(timestamp >> 16), byte(timestamp >> 8), byte(timestamp), byte(timestamp >> 24), //timestamp
		0x00, 0x00, 0x00}) //stream iD

	w.Write([]byte{0x17, 0x00, 0x00, 0x00, 0x00})
	w.Write([]byte{
		0x01,                                //configurationVersion
		sps[1],                              //AVCProfileIndication
		sps[2],                              //profile_compatibiltity
		sps[3],                              //AVCLevelIndication
		0xFF,                                //lengthSizeMinusOne
		0xE1,                                //numOfSequenceParameterSets
		byte(len(sps) >> 8), byte(len(sps)), //sequenceParameterSetLength
	})
	w.Write(sps)
	w.Write([]byte{
		0x01,                                //numOfPictureParamterSets
		byte(len(pps) >> 8), byte(len(pps)), //pictureParameterSetLength
	})
	//timestamp += 125
	w.Write(pps)
	prelen = dlen + 11

}

func getHeader() []byte {
	fi, err := os.Open("source/receive.flv")
	if err != nil {
		log.Println("open error : " + err.Error())
	}
	defer fi.Close()
	r := bufio.NewReader(fi)

	buf := make([]byte, 20480)
	_, err = r.Read(buf)
	if err != nil && err.Error() != "EOF" {
		log.Println("read error : " + err.Error())
	}
	log.Println(buf[:9])
	log.Println(buf[9:13])
	taglen := int(buf[14])<<16 + int(buf[15])<<8 + int(buf[16])
	log.Println("taglen", taglen)
	buf[4] = 0x01
	log.Println(buf[:24+taglen])
	return buf[:24+taglen]
}
