package gofpdf

// ClipRect begins a rectangular clipping operation. The rectangle is of width
// w and height h. Its upper left corner is positioned at point (x, y). outline
// is true to draw a border with the current draw color and line width centered
// on the rectangle's perimeter. Only the outer half of the border will be
// shown. After calling this method, all rendering operations (for example,
// Image(), LinearGradient(), etc) will be clipped by the specified rectangle.
// Call ClipEnd() to restore unclipped operations.
//
// This ClipText() example demonstrates this method.
func (gp *Fpdf) ClipRect(x, y, w, h float64, outline bool) {
	gp.UnitsToPointsVar(&x, &y, &w, &h)
	style := "n"
	if outline {
		style = "S"
	}

	gp.currentContent().AppendStreamClipRect(x, y, w, h, style)
}

// ClipText begins a clipping operation in which rendering is confined to the
// character string specified by txtStr. The origin (x, y) is on the left of
// the first character at the baseline. The current font is used. outline is
// true to draw a border with the current draw color and line width centered on
// the perimeters of the text characters. Only the outer half of the border
// will be shown. After calling this method, all rendering operations (for
// example, Image(), LinearGradient(), etc) will be clipped. Call ClipEnd() to
// restore unclipped operations.
func (gp *Fpdf) ClipText(x, y float64, txtStr string, outline bool) {
	gp.UnitsToPointsVar(&x, &y)
	style := 7
	if outline {
		style = 5
	}

	gp.currentContent().AppendStreamClipText(x, y, txtStr, style)
}

// ClipRoundedRect begins a rectangular clipping operation. The rectangle is of
// width w and height h. Its upper left corner is positioned at point (x, y).
// The rounded corners of the rectangle are specified by radius r. outline is
// true to draw a border with the current draw color and line width centered on
// the rectangle's perimeter. Only the outer half of the border will be shown.
// After calling this method, all rendering operations (for example, Image(),
// LinearGradient(), etc) will be clipped by the specified rectangle. Call
// ClipEnd() to restore unclipped operations.
//
// This ClipText() example demonstrates this method.
func (gp *Fpdf) ClipRoundedRect(x, y, w, h, r float64, outline bool) {
	gp.UnitsToPointsVar(&x, &y, &w, &h, &r)
	style := "n"
	if outline {
		style = "S"
	}

	gp.currentContent().AppendStreamClipRoundedRect(x, y, w, h, r, style)
}

// ClipEllipse begins an elliptical clipping operation. The ellipse is centered
// at (x, y). Its horizontal and vertical radii are specified by rx and ry.
// outline is true to draw a border with the current draw color and line width
// centered on the ellipse's perimeter. Only the outer half of the border will
// be shown. After calling this method, all rendering operations (for example,
// Image(), LinearGradient(), etc) will be clipped by the specified ellipse.
// Call ClipEnd() to restore unclipped operations.
//
// This ClipText() example demonstrates this method.
func (gp *Fpdf) ClipEllipse(x, y, rx, ry float64, outline bool) {
	gp.UnitsToPointsVar(&x, &y, &rx, &ry)
	style := "n"
	if outline {
		style = "S"
	}

	gp.currentContent().AppendStreamClipEllipse(x, y, rx, ry, style)
}

// ClipCircle begins a circular clipping operation. The circle is centered at
// (x, y) and has radius r. outline is true to draw a border with the current
// draw color and line width centered on the circle's perimeter. Only the outer
// half of the border will be shown. After calling this method, all rendering
// operations (for example, Image(), LinearGradient(), etc) will be clipped by
// the specified circle. Call ClipEnd() to restore unclipped operations.
//
// The ClipText() example demonstrates this method.
func (gp *Fpdf) ClipCircle(x, y, r float64, outline bool) {
	gp.ClipEllipse(x, y, r, r, outline)
}

// ClipPolygon begins a clipping operation within a polygon. The figure is
// defined by a series of vertices specified by points. The x and y fields of
// the points use the units established in New(). The last point in the slice
// will be implicitly joined to the first to close the polygon. outline is true
// to draw a border with the current draw color and line width centered on the
// polygon's perimeter. Only the outer half of the border will be shown. After
// calling this method, all rendering operations (for example, Image(),
// LinearGradient(), etc) will be clipped by the specified polygon. Call
// ClipEnd() to restore unclipped operations.
//
// The ClipText() example demonstrates this method.
func (gp *Fpdf) ClipPolygon(points []Point, outline bool) {
	style := "n"
	if outline {
		style = "S"
	}

	for x := 0; x < len(points); x++ {
		gp.UnitsToPointsVar(&points[x].X, &points[x].Y)
	}

	gp.currentContent().AppendStreamClipPolygon(points, style)
}

// ClipEnd ends a clipping operation that was started with a call to
// ClipRect(), ClipRoundedRect(), ClipText(), ClipEllipse(), ClipCircle() or
// ClipPolygon(). Clipping operations can be nested. The document cannot be
// successfully output while a clipping operation is active.
//
// The ClipText() example demonstrates this method.
func (gp *Fpdf) ClipEnd() {
	gp.currentContent().AppendStreamClipEnd()
}
