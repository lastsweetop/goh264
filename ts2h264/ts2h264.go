package ts2h264

import (
	"bufio"
	"goh264/tools"
	"log"
	"math"
	"os"
	"sync"
)

var (
	bufChan    = make(chan []byte, 10)
	pesBufChan = make(chan []byte, 10)

	closeSplitChan = false
	closeWriteChan = false
	wg             = sync.WaitGroup{}
	syncTs         = true
)

func Transfer(inPath, outPath string) {
	fi, err := os.Open(inPath)
	if err != nil {
		log.Println("open error : " + err.Error())
	}
	defer fi.Close()
	r := bufio.NewReader(fi)

	go splitTs()
	go splitPes(outPath)

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

func splitTs() {
	wg.Add(1)
	defer func() {
		wg.Done()
	}()

	temp := make([]byte, 0)
	pmtId := 0
	videoId := 0
	payload := make([]byte, 0)
	for {
		select {
		case buf := <-bufChan:
			buf = tools.Merge(temp, buf)
			temp = make([]byte, 0)
			i := 0
			for i < len(buf) {
				if i > len(buf)-188 {
					temp = buf[i:] //剩余不足4个字节的数据
					break
				}
				if syncTs && buf[i] == 0x47 && buf[i+188] == 0x47 && buf[i+376] == 0x47 {
					syncTs = false
				}
				tspkg := buf[i : i+188]
				//log.Println(tspkg)
				//transport_error_indicator := (tspkg[1] & 0x80) >> 7
				//log.Println("transport_error_indicator", transport_error_indicator)
				payloadUnitStartIndicator := (tspkg[1] & 0x40) >> 6
				//log.Println("payloadUnitStartIndicator", payloadUnitStartIndicator)
				//transport_priority := (tspkg[1] & 0x20) >> 5
				//log.Println("transport_priority", transport_priority)
				pid := int(tspkg[1]&0x1F)<<8 + int(tspkg[2])
				//log.Println("pid", pid)
				//transport_scrambling_control:=(tspkg[3]&0xC0)>>6
				//log.Println("transport_scrambling_control",transport_scrambling_control)
				adaptationFieldControl := (tspkg[3] & 0x30) >> 4
				//log.Println("adaptationFieldControl", adaptationFieldControl)
				//continuity_counter:=tspkg[3]&0x0F
				//log.Println("continuity_counter",continuity_counter)
				var adaptationFieldLength byte = 0
				if adaptationFieldControl == 3 || adaptationFieldControl == 2 {
					adaptationFieldLength = tspkg[4] + 1
				}

				if pid == 0 {
					log.Println("PAT")
					payload = tspkg[4+adaptationFieldLength+payloadUnitStartIndicator:]
					pmtId = getPMTIdfromPAT(payload)
				} else if pid == pmtId {
					log.Println("PMT")
					payload = tspkg[4+adaptationFieldLength+payloadUnitStartIndicator:]
					videoId = parserPMT(payload)
				} else if pid == videoId {
					payload = tspkg[4+adaptationFieldLength:]
					//log.Println(payload)
					pesBufChan <- payload
				}

				i += 188
			}
			break
		default:
			if closeSplitChan {
				closeWriteChan = true
				return
			}
		}
	}
}

func getPMTIdfromPAT(pat []byte) int {
	sectionLength := int(pat[1]&0x0F)<<8 + int(pat[2])
	//log.Println("section_length", sectionLength)
	for i := 8; i < sectionLength-1; i += 4 {
		if pat[i] == 0x00 && pat[i+1] == 0x00 {
			//log.Println("NIT")
		} else if pat[i] == 0x00 && pat[i+1] == 0x01 {
			PID := int(pat[i+2]&0x1F)<<8 + int(pat[i+3])
			//log.Println("PMT PID", PID)
			return PID
		}
	}
	return 0
}

func parserPMT(pmt []byte) int {
	//tableId := pmt[0]
	//log.Println("table_id", tableId)
	sectionLength := int(pmt[1]&0x0F)<<8 + int(pmt[2])
	//log.Println("sectionLength", sectionLength)
	//programNumber := int(pmt[3])<<8 + int(pmt[4])
	//log.Println("programNumber", programNumber)
	//PcrPid := int(pmt[8]&0X1F)<<8 + int(pmt[9])
	//log.Println("PcrPid", PcrPid)
	//programInfoLength := int(pmt[10]&0X0F)<<8 + int(pmt[11])
	//log.Println("programInfoLength", programInfoLength)
	for i := 12; i < sectionLength-1; i += 6 {
		streamType := pmt[i]
		//log.Println("streamType", streamType)
		elementaryPid := int(pmt[i+1]&0x1F)<<8 + int(pmt[i+2])
		//log.Println("elementaryPid", elementaryPid)
		if streamType == 0x1b {
			return elementaryPid
		}
	}
	return 0
}

func splitPes(outPath string) {
	fi, err := os.Create(outPath)
	if err != nil {
		log.Println("open error : " + err.Error())
	}
	w := bufio.NewWriter(fi)

	defer func() {
		wg.Done()
	}()
	wg.Add(1)

	i := 0
	left := 0
	for {
		select {
		case buf := <-pesBufChan:
			//log.Println(buf)
			//log.Println("streamId", streamId)
			if left >= len(buf) {
				if left == math.MaxInt32 {
					j := 0
					for ; j+3 < len(buf); j++ {
						if buf[j] == 0x00 && buf[j+1] == 0x00 && buf[j+2] == 0x01 && buf[j+3] == 0xE0 {
							log.Println("pes")
							left = j
							break
						}
					}
					//log.Println("vedio data")
					if j+3 == len(buf) {
						//left = left - len(buf)
						w.Write(buf)
						break
					}
				} else {
					left = left - len(buf)
					w.Write(buf)
					break
				}
			}

			if left > 0 {
				w.Write(buf[:left])
			}
			i = left
			streamId := buf[i+3]
			//log.Println("streamId", streamId)
			if streamId == 0xc0 {
				log.Println("音频")
			} else if streamId == 0xe0 {
				log.Println("视频")
			}
			pesPacketLength := int(buf[i+4])<<8 + int(buf[i+5])

			log.Println("pesPacketLength", pesPacketLength)
			pesHeaderlength := int(buf[i+8])
			log.Println("pesHeaderlength", pesHeaderlength)
			log.Println("h264 header :", buf[i+15+pesHeaderlength:i+20+pesHeaderlength], "size ", i+6+pesPacketLength-i+9+pesHeaderlength)
			log.Println(buf[i+15+pesHeaderlength:])
			w.Write(buf[i+15+pesHeaderlength:])
			//w.Write(buf[i:])

			if pesPacketLength == 0 {
				left = math.MaxInt32
			} else {
				left = i + 6 + pesPacketLength - len(buf)
			}
			break
		default:
			if closeWriteChan {
				return
			}
		}
	}
}
