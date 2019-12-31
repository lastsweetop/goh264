package test

import (
	"fmt"
	"goh264/model"
	"testing"
)

func TestU(t *testing.T) {
	streamData := model.StreamData{Data: []byte{0XFF}, Index: 0}

	fmt.Printf("0xff %d \n",0x80)

	data:=streamData.F(8)

	fmt.Println(data)
}
