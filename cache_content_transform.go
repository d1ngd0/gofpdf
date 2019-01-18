package gofpdf

import (
	"fmt"
	"io"
)

type cacheContentTransformBegin struct{}

func (c *cacheContentTransformBegin) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprint(w, "q\n")
	return nil
}

type cacheContentTransformEnd struct{}

func (c *cacheContentTransformEnd) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprint(w, "Q\n")
	return nil
}

type cacheContentTransform struct {
	matrix TransformMatrix
}

func (c *cacheContentTransform) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprintf(w, "%.5f %.5f %.5f %.5f %.5f %.5f cm\n", c.matrix.A, c.matrix.B, c.matrix.C, c.matrix.D, c.matrix.E, c.matrix.F)
	return nil
}
