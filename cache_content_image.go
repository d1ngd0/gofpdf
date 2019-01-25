package gofpdf

import (
	"fmt"
	"io"
)

type cacheContentImage struct {
	id   string
	x    float64
	y    float64
	h    float64
	rect Rect
}

func (c *cacheContentImage) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprintf(w, "q %0.2f 0 0 %0.2f %0.2f %0.2f cm /%s Do Q\n", c.rect.W, c.rect.H, c.x, c.h-(c.y+c.rect.H), c.id)
	return nil
}
