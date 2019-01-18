package gofpdf

import (
	"fmt"
	"io"
)

// Fill types and manipulants to get the numbers pdf spec wants
const (
	colorTypeStrokeRGB  = "RG"
	colorTypeStrokeCMYK = "K"

	colorTypeFillRGB  = "rg"
	colorTypeFillCMYK = "k"

	rgbManipulant  = 0.00392156862745
	cmykManipulant = 100
)

type cacheContentColor struct {
	colorType  string
	r, g, b    uint8
	c, m, y, k uint8
}

func (c *cacheContentColor) write(w io.Writer, protection *PDFProtection) error {
	switch c.colorType {
	case colorTypeStrokeRGB, colorTypeFillRGB:
		rFloat := float64(c.r) * rgbManipulant
		gFloat := float64(c.g) * rgbManipulant
		bFloat := float64(c.b) * rgbManipulant

		fmt.Fprintf(w, "%.2f %.2f %.2f %s\n", rFloat, gFloat, bFloat, c.colorType)
	case colorTypeStrokeCMYK, colorTypeFillCMYK:
		cFloat := float64(c.c) / cmykManipulant
		mFloat := float64(c.m) / cmykManipulant
		yFloat := float64(c.y) / cmykManipulant
		kFloat := float64(c.k) / cmykManipulant

		fmt.Fprintf(w, "%.2f %.2f %.2f %.2f %s\n", cFloat, mFloat, yFloat, kFloat, c.colorType)
	}
	return nil
}
