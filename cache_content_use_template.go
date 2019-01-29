package gofpdf

import (
	"fmt"
	"io"
)

type cacheContentUseTemplate struct {
	pageHeight float64
	id         string
	x          float64
	y          float64
	sx         float64
	sy         float64
	h          float64
}

func (c *cacheContentUseTemplate) write(w io.Writer, protection *PDFProtection) error {
	tx := c.x
	ty := c.pageHeight - c.y - c.h

	fmt.Fprintf(w, "q %.4f 0 0 %.4f %.4f %.4f cm\n", c.sx, c.sy, tx, ty)
	fmt.Fprintf(w, "/%s Do Q\n", c.id)

	return nil
}
