package gofpdf

import (
	"io"
)

//IObj inteface for all pdf object
type IObj interface {
	getType() string
	write(w io.Writer, objID int) error
}
