package gofpdf

import (
	"strconv"

	"github.com/d1ngd0/gofpdf/bp"
)

func FontConvertHelper_Cw2Str(cw FontCw) string {
	buff := bp.GetBuffer()
	defer bp.PutBuffer(buff)
	buff.WriteString(" ")
	for i := 32; i <= 255; i++ {
		buff.WriteString(strconv.Itoa(cw[Chr(i)]) + " ")
	}
	return buff.String()
}
