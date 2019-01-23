package gofpdf

import (
	"fmt"
	"io"
)

type cacheContentPoint struct {
	pageHeight float64
	x          float64
	y          float64
}

func (c *cacheContentPoint) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprintf(w, "%.2f %.2f m\n", c.x, c.pageHeight-c.y)
	return nil
}
