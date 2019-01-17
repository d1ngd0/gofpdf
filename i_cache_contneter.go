package gofpdf

import (
	"io"
)

type iCacheContent interface {
	write(w io.Writer, protection *PDFProtection) error
}
