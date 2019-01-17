package gofpdf

import (
	"io"
)

//IObj inteface for all pdf object
type IObj interface {
	init(func() *Fpdf)
	getType() string
	write(w io.Writer, objID int) error
}
