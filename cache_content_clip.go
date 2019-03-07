package gofpdf

import (
	"fmt"
	"io"
	"math"
	"strings"
)

type cacheContentClipBegin struct{}

func (c *cacheContentClipBegin) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprint(w, "q ")
	return nil
}

type cacheContentClipEnd struct{}

func (c *cacheContentClipEnd) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprint(w, "Q\n")
	return nil
}

type cacheContentClipRect struct {
	pageHeight float64
	x, y, w, h float64
	style      string
}

func (c *cacheContentClipRect) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprintf(w, "%.2f %.2f %.2f %.2f re W %s\n", c.x, c.pageHeight-c.y, c.w, -c.h, c.style)
	return nil
}

type cacheContentClipText struct {
	pageHeight float64
	x, y       float64
	txtStr     string
	style      int
}

func (c *cacheContentClipText) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprintf(w, "BT %.5f %.5f Td %d Tr (%s) Tj ET\n", c.x, c.pageHeight-c.y, c.style, c.escape(c.txtStr))
	return nil
}

// There is probably a better way to do this
func (c *cacheContentClipText) escape(s string) string {
	s = strings.Replace(s, "\\", "\\\\", -1)
	s = strings.Replace(s, "(", "\\(", -1)
	s = strings.Replace(s, ")", "\\)", -1)
	s = strings.Replace(s, "\r", "\\r", -1)
	return s
}

type cacheContentClipRoundedRect struct {
	pageHeight    float64
	style         string
	x, y, w, h, r float64
}

func (c *cacheContentClipRoundedRect) clipArc(w io.Writer, x1, y1, x2, y2, x3, y3 float64) {
	fmt.Fprintf(w, "%.5f %.5f %.5f %.5f %.5f %.5f c \n", x1, c.pageHeight-y1, x2, c.pageHeight-y2, x3, c.pageHeight-y3)
}

func (c *cacheContentClipRoundedRect) write(w io.Writer, protection *PDFProtection) error {
	fmt.Fprint(w)

	myArc := (4.0 / 3.0) * (math.Sqrt2 - 1.0)
	fmt.Fprintf(w, "%.5f %.5f m\n", c.x+c.r, c.pageHeight-c.y)
	xc := c.x + c.w - c.r
	yc := c.y + c.r
	fmt.Fprintf(w, "%.5f %.5f l\n", xc, c.pageHeight-c.y)
	c.clipArc(w, xc+c.r*myArc, yc-c.r, xc+c.r, yc-c.r*myArc, xc+c.r, yc)
	xc = c.x + c.w - c.r
	yc = c.y + c.h - c.r
	fmt.Fprintf(w, "%.5f %.5f l\n", c.x+c.w, c.pageHeight-yc)
	c.clipArc(w, xc+c.r, yc+c.r*myArc, xc+c.r*myArc, yc+c.r, xc, yc+c.r)
	xc = c.x + c.r
	yc = c.y + c.h - c.r
	fmt.Fprintf(w, "%.5f %.5f l\n", xc, c.pageHeight-(c.y+c.h))
	c.clipArc(w, xc-c.r*myArc, yc+c.r, xc-c.r, yc+c.r*myArc, xc-c.r, yc)
	xc = c.x + c.r
	yc = c.y + c.r
	fmt.Fprintf(w, "%.5f %.5f l\n", c.x, c.pageHeight-yc)
	c.clipArc(w, xc-c.r, yc-c.r*myArc, xc-c.r*myArc, yc-c.r, xc, yc-c.r)
	fmt.Fprintf(w, " W %s\n", c.style)

	return nil
}

type cacheContentClipEllipse struct {
	pageHeight   float64
	style        string
	x, y, rx, ry float64
}

func (c *cacheContentClipEllipse) write(w io.Writer, protection *PDFProtection) error {
	lx := (4.0 / 3.0) * c.rx * (math.Sqrt2 - 1)
	ly := (4.0 / 3.0) * c.ry * (math.Sqrt2 - 1)
	h := c.pageHeight
	fmt.Fprintf(w, "%.5f %.5f m %.5f %.5f %.5f %.5f %.5f %.5f c\n",
		c.x+c.rx, h-c.y,
		c.x+c.rx, h-(c.y-ly),
		c.x+lx, h-c.y-c.ry,
		c.x, h-(c.y-c.ry))
	fmt.Fprintf(w, "%.5f %.5f %.5f %.5f %.5f %.5f c\n",
		c.x-lx, h-(c.y-c.ry),
		c.x-c.rx, h-(c.y-ly),
		c.x-c.rx, h-c.y)
	fmt.Fprintf(w, "%.5f %.5f %.5f %.5f %.5f %.5f c\n",
		c.x-c.rx, h-(c.y+ly),
		c.x-lx, h-(c.y+c.ry),
		c.x, h-(c.y+c.ry))
	fmt.Fprintf(w, "%.5f %.5f %.5f %.5f %.5f %.5f c W %s\n",
		c.x+lx, h-(c.y+c.ry),
		c.x+c.rx, h-(c.y+ly),
		c.x+c.rx, h-c.y,
		c.style)
	return nil
}

type cacheContentClipPolygon struct {
	pageHeight float64
	style      string
	points     []Point
}

func (c *cacheContentClipPolygon) write(w io.Writer, protection *PDFProtection) error {
	for j, pt := range c.points {
		d := "l"
		if j == 0 {
			d = "m"
		}

		fmt.Fprintf(w, "%.5f %.5f %s ", pt.X, c.pageHeight-pt.Y, d)
	}

	fmt.Fprintf(w, "h W %s\n", c.style)
	return nil
}

type cacheContentGradientClip struct {
}
