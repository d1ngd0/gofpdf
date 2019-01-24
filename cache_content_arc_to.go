package gofpdf

import (
	"fmt"
	"io"
	"math"
)

type cacheContentArcTo struct {
	pageHeight                                float64
	currentX, currentY                        *float64
	x, y, rx, ry, degRotate, degStart, degEnd float64
	styleStr                                  string
	path                                      bool
}

func (c *cacheContentArcTo) write(w io.Writer, protection *PDFProtection) error {
	x := c.x
	y := c.pageHeight - c.y

	segments := int(c.degEnd-c.degStart) / 60
	if segments < 2 {
		segments = 2
	}
	angleStart := c.degStart * math.Pi / 180
	angleEnd := c.degEnd * math.Pi / 180
	angleTotal := angleEnd - angleStart
	dt := angleTotal / float64(segments)
	dtm := dt / 3
	if c.degRotate != 0 {
		a := -c.degRotate * math.Pi / 180
		fmt.Fprintf(w, "q %.5f %.5f %.5f %.5f %.5f %.5f cm\n",
			math.Cos(a), -1*math.Sin(a),
			math.Sin(a), math.Cos(a), x, y)
		x = 0
		y = 0
	}
	t := angleStart
	a0 := x + c.rx*math.Cos(t)
	b0 := y + c.ry*math.Sin(t)
	c0 := -c.rx * math.Sin(t)
	d0 := c.ry * math.Cos(t)
	sx := a0 // start point of arc
	sy := c.pageHeight - b0
	if c.path {
		if *c.currentX != sx || *c.currentY != sy {
			lineCache := &cacheContentLineTo{c.pageHeight, sx, sy}
			lineCache.write(w, protection)
		}
	} else {
		pointCache := &cacheContentPoint{c.pageHeight, sx, sy}
		pointCache.write(w, protection)
	}
	for j := 1; j <= segments; j++ {
		// Draw this bit of the total curve
		t = (float64(j) * dt) + angleStart
		a1 := x + c.rx*math.Cos(t)
		b1 := y + c.ry*math.Sin(t)
		c1 := -c.rx * math.Sin(t)
		d1 := c.ry * math.Cos(t)
		curveTo := &cacheContentCurveBezierCubic{
			c.pageHeight,
			a0 + (c0 * dtm),
			c.pageHeight - (b0 + (d0 * dtm)),
			a1 - (c1 * dtm),
			c.pageHeight - (b1 - (d1 * dtm)),
			a1,
			c.pageHeight - b1,
		}
		curveTo.write(w, protection)
		a0 = a1
		b0 = b1
		c0 = c1
		d0 = d1
		if c.path {
			*c.currentX = a1
			*c.currentY = c.pageHeight - b1
		}
	}

	if !c.path {
		drawPath := cacheContentDrawPath{c.styleStr}
		drawPath.write(w, protection)
	}

	if c.degRotate != 0 {
		fmt.Fprintf(w, "Q\n")
	}

	return nil
}
