package stream2file

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
)

func Stream2File(inPath string, outPath string) {
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
	startI := false

	defer func() {
		fi.Close()
		wg.Done()
	}()

	wg.Add(1)
	w := bufio.NewWriter(fi)

	i := 0
	for {
		select {
		case frame := <-frameChan:
			if startI {
				w.Write([]byte{0x00, 0x00, 0x00, 0x01})
				w.Write(frame)
			}

			i++
			log.Println()
			log.Println("帧", i, "的数据长度：", len(frame))

			dataStream := model.StreamData{Data: frame, Index: 0}
			_ = dataStream.F(1)
			//log.Printf("forbidden_bit %d\n", ForbiddenBit)
			_ = dataStream.U(2)
			//log.Printf("nal_reference_bit %d\n", NalReferenceBit)
			NalUnitType := dataStream.U(5)
			//log.Printf("nal_unit_type %d\n", NalUnitType)

			switch NalUnitType {
			case NalUnitTypes.SPS:
				if !startI {
					w.Write([]byte{0x00, 0x00, 0x00, 0x01})
					w.Write(frame)
				}
				startI = true

				log.Println("SPS")
				//spsStream := model.SpsStream{Data: frame[1:], Index: 0}
				profileIdc := dataStream.U(8)
				log.Printf("profile_idc %d\n", profileIdc)
				constraintSet0Flag := dataStream.U(1)
				log.Printf("constraint_set0_flag %d\n", constraintSet0Flag)
				constraintSet1Flag := dataStream.U(1)
				log.Printf("constraint_set1_flag %d\n", constraintSet1Flag)
				constraintSet2Flag := dataStream.U(1)
				log.Printf("constraint_set2_flag %d\n", constraintSet2Flag)
				constraintSet3Flag := dataStream.U(1)
				log.Printf("constraint_set3_flag %d\n", constraintSet3Flag)
				constraintSet4Flag := dataStream.U(1)
				log.Printf("constraint_set4_flag %d\n", constraintSet4Flag)
				constraintSet5Flag := dataStream.U(1)
				log.Printf("constraint_set5_flag %d\n", constraintSet5Flag)
				reservedZero2bits := dataStream.U(2)
				log.Printf("reserved_zero_2bits %d\n", reservedZero2bits)
				levelIdc := dataStream.U(8)
				log.Printf("level_idc %d\n", levelIdc)
				seqParameterSetId := dataStream.UE()
				log.Printf("seq_parameter_set_id %d\n", seqParameterSetId)
				log2MaxFrameNumMinus4 := dataStream.UE()
				log.Printf("log2_max_frame_num_minus4 %d\n", log2MaxFrameNumMinus4)
				picOrderCntType := dataStream.UE()
				log.Printf("pic_order_cnt_type %d\n", picOrderCntType)
				if picOrderCntType == 0 {
					log2MaxPicOrderCntLsbMinus4 := dataStream.UE()
					log.Printf("log2_max_pic_order_cnt_lsb_minus4 %d\n", log2MaxPicOrderCntLsbMinus4)
				}
				maxNumRefFrames := dataStream.UE()
				log.Printf("max_num_ref_frames %d\n", maxNumRefFrames)
				gapsInFrameNumValueAllowedFlag := dataStream.U(1)
				log.Printf("gaps_in_frame_num_value_allowed_flag %d\n", gapsInFrameNumValueAllowedFlag)

				picWidthInMbsMinus1 := dataStream.UE()
				log.Printf("pic_width_in_mbs_minus1 %d\n", picWidthInMbsMinus1)

				log.Println("宽", (picWidthInMbsMinus1+1)*16)

				picHeightInMapUnitsMinus1 := dataStream.UE()
				log.Printf("pic_height_in_map_units_minus1 %d\n", picHeightInMapUnitsMinus1)

				frameMbsOnlyFlag := dataStream.U(1)
				log.Printf("frame_mbs_only_flag %d\n", frameMbsOnlyFlag)
				log.Println("高", (2-frameMbsOnlyFlag)*(picHeightInMapUnitsMinus1+1)*16)

				direct8x8InferenceFlag := dataStream.U(1)
				log.Printf("direct_8x8_inference_flag %d\n", direct8x8InferenceFlag)

				frameCroppingFlag := dataStream.U(1)
				log.Printf("frame_cropping_flag %d\n", frameCroppingFlag)

				vuiParametersPresentFlag := dataStream.U(1)
				log.Printf("vui_parameters_present_flag %d\n", vuiParametersPresentFlag)

			case NalUnitTypes.PPS:
			case NalUnitTypes.SEI:
			case NalUnitTypes.IDR:
				log.Println("IDR")
			case NalUnitTypes.NOIDR:
			default:
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
