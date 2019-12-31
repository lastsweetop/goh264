package model

import "fmt"

type NalHeader struct {
	ForbiddenBit    uint8
	NalReferenceBit uint8
	NalUnitType     uint8
}

func (n NalHeader) String() string {
	return fmt.Sprintf("{ForbiddenBit : %d, NalReferenceBit : %d, NalUnitType: %d}",
		n.ForbiddenBit, n.NalReferenceBit, n.NalUnitType)
}

