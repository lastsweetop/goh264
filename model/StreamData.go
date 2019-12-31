package model

type StreamData struct {
	Data  []byte
	Index uint32
}

func (s *StreamData) F(bitCount byte) int {
	var val int

	for i := byte(0); i < bitCount; i++ {
		val <<= 1

		if (s.Data[s.Index/8])&(0x80>>(s.Index%8)) != 0 {
			val |= 1
		}

		s.Index++
	}
	return val

}

func (s *StreamData) U(bitCount byte) uint {
	var val uint

	for i := byte(0); i < bitCount; i++ {
		val <<= 1

		if (s.Data[s.Index/8])&(0x80>>(s.Index%8)) != 0 {
			val |= 1
		}
		s.Index++
	}

	return val
}

func (s *StreamData) UE() uint {
	var zeroNum uint

	for s.U(1) == 0 && zeroNum < 32 {
		zeroNum++
	}
	return 1<<zeroNum - 1 + s.U(byte(zeroNum))
}
