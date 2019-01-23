package gofpdf

import (
	"fmt"
	"io"
)

type cacheContentCurve struct {
	pageHeight   float64
	x, y, cx, cy float64
}

func (c *cacheContentCurve) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprintf(w, "%.5f %.5f %.5f %.5f v\n", c.cx, c.pageHeight-c.cy, c.x, c.pageHeight-c.y)

	return nil
}
