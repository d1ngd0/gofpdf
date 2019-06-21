package gofpdf

import (
	"bytes"
	"compress/zlib" // for constants
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const subsetFont = "SubsetFont"

// the default margin if no margins are set
const defaultMargin = 1 * conversion_Unit_CM

//Fpdf : A simple library for generating PDF written in Go lang
type Fpdf struct {

	//page Margin
	// leftMargin float64
	// topMargin  float64
	margins Margins

	pdfObjs *pdfObjs
	anchors map[string]anchorOption

	curr Current

	indexEncodingObjFonts []int
	//IsUnderline bool

	// Buffer for io.Reader compliance
	buf bytes.Buffer

	//pdf PProtection
	pdfProtection *PDFProtection

	// content streams only
	compressLevel int

	info        *PdfInfo
	appliedOpts []PdfOption
}

// Set a page boundary
func (gp *Fpdf) SetPageBoundary(pb *PageBoundary) {
	if page := gp.currentPage(); page != nil {
		page.pageOption.AddPageBoundary(pb)
	} else {
		gp.curr.pageOption.AddPageBoundary(pb)
	}
}

func (gp *Fpdf) GetPageBoundary(t int) *PageBoundary {
	if page := gp.currentPage(); page != nil {
		if boundary := page.pageOption.GetBoundary(t); boundary != nil {
			return boundary
		}
	}

	return gp.curr.pageOption.GetBoundary(t)
}

func (gp *Fpdf) GetBoundarySize(t int) (rec Rect) {
	if pb := gp.GetPageBoundary(t); pb != nil {
		rec = pb.Size
	}
	return
}

func (gp *Fpdf) GetBoundaryPosition(t int) (p Point) {
	if pb := gp.GetPageBoundary(t); pb != nil {
		p = pb.Position
	}
	return
}

func (gp *Fpdf) GetBoundaryX(t int) float64 {
	return gp.GetBoundaryPosition(t).X
}

func (gp *Fpdf) GetBoundaryY(t int) float64 {
	return gp.GetBoundaryPosition(t).Y
}

func (gp *Fpdf) GetBoundaryWidth(t int) float64 {
	return gp.GetBoundarySize(t).W
}

func (gp *Fpdf) GetBoundaryHeight(t int) float64 {
	return gp.GetBoundarySize(t).H
}

func (gp *Fpdf) SetPageSize(w, h float64) {
	pb := gp.NewPageSizeBoundary(w, h)
	gp.SetPageBoundary(pb)
}

func (gp *Fpdf) SetCropBox(x, y, w, h float64) {
	pb := gp.NewCropPageBoundary(x, y, w, h)
	gp.SetPageBoundary(pb)
}

func (gp *Fpdf) SetBleedBox(x, y, w, h float64) {
	pb := gp.NewBleedPageBoundary(x, y, w, h)
	gp.SetPageBoundary(pb)
}

func (gp *Fpdf) SetTrimBox(x, y, w, h float64) {
	pb := gp.NewTrimPageBoundary(x, y, w, h)
	gp.SetPageBoundary(pb)
}

func (gp *Fpdf) SetArtBox(x, y, w, h float64) {
	pb := gp.NewArtPageBoundary(x, y, w, h)
	gp.SetPageBoundary(pb)
}

//SetLineWidth : set line width
func (gp *Fpdf) SetLineWidth(width float64) {
	gp.curr.lineWidth = gp.UnitsToPoints(width)
	gp.currentContent().AppendStreamSetLineWidth(gp.curr.lineWidth)
}

//SetCompressLevel : set compress Level for content streams
// Possible values for level:
//    -2 HuffmanOnly, -1 DefaultCompression (which is level 6)
//     0 No compression,
//     1 fastest compression, but not very good ratio
//     9 best compression, but slowest
func (gp *Fpdf) SetCompressLevel(level int) {
	errfmt := "compress level too %s, using %s instead\n"
	if level < -2 { //-2 = zlib.HuffmanOnly
		fmt.Fprintf(os.Stderr, errfmt, "small", "DefaultCompression")
		level = zlib.DefaultCompression
	} else if level > zlib.BestCompression {
		fmt.Fprintf(os.Stderr, errfmt, "big", "BestCompression")
		level = zlib.BestCompression
		return
	}
	// sanity check complete
	gp.compressLevel = level
}

//SetNoCompression : compressLevel = 0
func (gp *Fpdf) SetNoCompression() {
	gp.compressLevel = zlib.NoCompression
}

//SetLineType : set line type  ("dashed" ,"dotted")
//  Usage:
//  pdf.SetLineType("dashed")
//  pdf.Line(50, 200, 550, 200)
//  pdf.SetLineType("dotted")
//  pdf.Line(50, 400, 550, 400)
func (gp *Fpdf) SetLineType(linetype string) {
	gp.currentContent().AppendStreamSetLineType(linetype)
}

const (
	CapStyleDefault = iota
	CapStyleRound
	CapStyleSquare
)

// SetLineCapStyle defines the line cap style. styleStr should be "butt",
// "round" or "square". A square style projects from the end of the line. The
// method can be called before the first page is created. The value is
// retained from page to page.
func (gp *Fpdf) SetLineCapStyle(style int) {
	if style != gp.curr.capStyle {
		gp.curr.capStyle = style
		gp.currentContent().AppendStreamSetCapStyle(style)
	}
}

const (
	JoinStyleRound   = 1
	JoinStyleBevel   = 2
	JoinStyleDefault = 0
)

// SetLineJoinStyle defines the line cap style. styleStr should be "miter",
// "round" or "bevel". The method can be called before the first page
// is created. The value is retained from page to page.
func (gp *Fpdf) SetLineJoinStyle(style int) {
	if style != gp.curr.joinStyle {
		gp.curr.joinStyle = style
		gp.currentContent().AppendStreamSetJoinStyle(gp.curr.joinStyle)
	}
}

// Beziergon draws a closed figure defined by a series of cubic Bézier curve
// segments. The first point in the slice defines the starting point of the
// figure. Each three following points p1, p2, p3 represent a curve segment to
// the point p3 using p1 and p2 as the Bézier control points.
//
// The x and y fields of the points use the units established in New().
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the ellipse's perimeter.
// Filling uses the current fill color.
func (gp *Fpdf) Beziergon(pts Points, styleStr string) error {
	// Thanks, Robert Lillack, for contributing this function.
	points := pts.ToPoints(gp.curr.unit)

	if len(points) < 4 {
		return fmt.Errorf("the number of points can not be less than 4. %d found", len(points))
	}

	gp.currentContent().AppendStreamPoint(points[0].XY())

	points = points[1:]
	for len(points) >= 3 {
		cx0, cy0 := points[0].XY()
		cx1, cy1 := points[1].XY()
		x1, y1 := points[2].XY()
		gp.currentContent().AppendStreamCurveBezierCubic(cx0, cy0, cx1, cy1, x1, y1)
		points = points[3:]
	}

	gp.currentContent().AppendStreamDrawPath(styleStr)
	return nil
}

// Beziertext writes text along a path defined by a series of cubic Bézier
// curve segments. Font size is reduced if the text exceeds avaiable arc length.
func (gp *Fpdf) Beziertext(pts Points, startBracket, endBracket float64, text string, opt CellOption, textOpt TextOption) error {
	points := pts.ToPoints(gp.curr.unit)

	if len(points) < 4 {
		return fmt.Errorf("the number of points can not be less than 4. %d found", len(points))
	}

	bs := NewBezierSpline(points)

	numrunes := len([]rune(text))
	err := gp.curr.Font_ISubset.AddChars(text)
	if err != nil {
		return err
	}

	pathlength := bs.Length() / 72.0
	if startBracket < 0.0 {
		startBracket = 0.0
	}
	if endBracket > pathlength || endBracket <= startBracket {
		endBracket = pathlength
	}
	// Make sure width doesnt exceed brackets or spline endpoints
	blength := endBracket - startBracket
	var textwidth float64
	for {
		textwidth, err = gp.MeasureTextWidth(text, textOpt)
		if err != nil {
			return err
		}
		if textwidth <= blength {
			break
		}
		// Reduce font size to fit
		r := blength / textwidth
		r, err = strconv.ParseFloat(fmt.Sprintf("%.3f", r), 64)
		r = r - 0.001
		if r < 0.0 {
			r = 0.0
		}
		gp.curr.Font_Size = gp.curr.Font_Size * r
	}

	endpts := make([]float64, numrunes+1)
	srunes := make([]string, numrunes)
	x := ""
	i := 0 // Explicit counter gives rune index instead of byte index
	for _, c := range text {
		srunes[i] = string(c)
		x += srunes[i]
		v, err := gp.MeasureTextWidth(x, textOpt)
		if err != nil {
			return err
		}
		endpts[i+1] = v
		i++
	}

	n := bezierSampleCardinality
	b := int(startBracket / pathlength * float64(n))
	c := int((pathlength - endBracket) / pathlength * float64(n))
	frac := textwidth / pathlength
	m := int(frac * float64(n))
	g := n - m - b - c
	tsubsUniform := bs.SampleByArcLength(bezierSampleCardinality)
	selected := make([]BezierPoint, numrunes)
	for i := 0; i < numrunes; i++ {
		// Select point from sample, taking into account bracket and alignment
		f := 0
		switch opt.Align {
		case Center, Center | Top, Center | Bottom, Center | Middle:
			f = g / 2
		case Right, Right | Top, Right | Bottom, Right | Middle:
			f = g
		}
		p := (endpts[i] + endpts[i+1]) / 2.0 // Midpoint of rune
		j := b + f + int((p/textwidth)*float64(m))
		if j < 0 {
			j = 0
		}
		if j >= n {
			j = n - 1
		}
		curveIndex, paramValue := tsubsUniform.At(j)
		selected[i] = BezierPoint{bs[curveIndex].At(paramValue), bs[curveIndex].NormalDegrees(paramValue)}
	}

	height := gp.curr.Font_Size
	descent := gp.curr.Font_ISubset.ttfp.TypoDescender()
	upm := gp.curr.Font_ISubset.ttfp.UnitsPerEm()
	cellopt := CellOption{
		Align:  Center,
		Border: opt.Border,
		Float:  opt.Float,
	}
	for i, v := range selected {
		gp.Rotate(v.normaldir, v.pt.X/72.0, v.pt.Y/72.0)
		width := endpts[i+1] - endpts[i]*72.0
		rect := Rect{W: width, H: height}
		gp.curr.X, gp.curr.Y = v.pt.X-(width/2.0), v.pt.Y-height- // Offset cell origin
			float64(descent)/float64(upm)*height // Move down to baseline
		t := srunes[i]
		if err := gp.curr.Font_ISubset.AddChars(t); err != nil {
			return err
		}
		if err = gp.currentContent().AppendStreamSubsetFont(rect, t, cellopt, textOpt); err != nil {
			return err
		}
		gp.RotateReset()
	}
	return nil
}

// CurveBezierCubic draws a single-segment cubic Bézier curve. The curve starts at
// the point (x0, y0) and ends at the point (x1, y1). The control points (cx0,
// cy0) and (cx1, cy1) specify the curvature. At the start point, the curve is
// tangent to the straight line between the start point and the control point
// (cx0, cy0). At the end point, the curve is tangent to the straight line
// between the end point and the control point (cx1, cy1).
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the curve's
// path. Filling uses the current fill color.
//
// This routine performs the same function as CurveCubic() but uses standard
// argument order.
//
// The Circle() example demonstrates this method.
func (gp *Fpdf) CurveBezierCubic(x0, y0, cx0, cy0, cx1, cy1, x1, y1 float64, styleStr string) {
	gp.UnitsToPointsVar(&x0, &y0, &cx0, &cy0, &cx1, &cy1, &x1, &y1)
	gp.currentContent().AppendStreamPoint(x0, y0)
	gp.currentContent().AppendStreamCurveBezierCubic(cx0, cy0, cx1, cy1, x1, y1)
	gp.currentContent().AppendStreamDrawPath(styleStr)
}

// CurveCubic draws a single-segment cubic Bézier curve. This routine performs
// the same function as CurveBezierCubic() but has a nonstandard argument order.
// It is retained to preserve backward compatibility.
func (gp *Fpdf) CurveCubic(x0, y0, cx0, cy0, x1, y1, cx1, cy1 float64, styleStr string) {
	// f.point(x0, y0)
	// f.outf("%.5f %.5f %.5f %.5f %.5f %.5f c %s", cx0*f.k, (f.h-cy0)*f.k,
	// cx1*f.k, (f.h-cy1)*f.k, x1*f.k, (f.h-y1)*f.k, fillDrawOp(styleStr))
	gp.CurveBezierCubic(x0, y0, cx0, cy0, cx1, cy1, x1, y1, styleStr)
}

// CurveBezierCubicTo creates a single-segment cubic Bézier curve. The curve
// starts at the current stylus location and ends at the point (x, y). The
// control points (cx0, cy0) and (cx1, cy1) specify the curvature. At the
// current stylus, the curve is tangent to the straight line between the
// current stylus location and the control point (cx0, cy0). At the end point,
// the curve is tangent to the straight line between the end point and the
// control point (cx1, cy1).
//
// The MoveTo() example demonstrates this method.
func (gp *Fpdf) CurveBezierCubicTo(cx0, cy0, cx1, cy1, x, y float64) {
	gp.UnitsToPointsVar(&cx0, &cy0, &cx1, &cy1, &x, &y)
	gp.currentContent().AppendStreamCurveBezierCubic(cx0, cy0, cx1, cy1, x, y)
	gp.curr.X, gp.curr.Y = x, y
}

// ClosePath creates a line from the current location to the last MoveTo point
// (if not the same) and mark the path as closed so the first and last lines
// join nicely.
//
// The MoveTo() example demonstrates this method.
func (gp *Fpdf) ClosePath() {
	gp.currentContent().AppendStreamClosePath()
}

// DrawPath actually draws the path on the page.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D".
// Path-painting operators as defined in the PDF specification are also
// allowed: "S" (Stroke the path), "s" (Close and stroke the path),
// "f" (fill the path, using the nonzero winding number), "f*"
// (Fill the path, using the even-odd rule), "B" (Fill and then stroke
// the path, using the nonzero winding number rule), "B*" (Fill and
// then stroke the path, using the even-odd rule), "b" (Close, fill,
// and then stroke the path, using the nonzero winding number rule) and
// "b*" (Close, fill, and then stroke the path, using the even-odd
// rule).
// Drawing uses the current draw color, line width, and cap style
// centered on the
// path. Filling uses the current fill color.
//
// The MoveTo() example demonstrates this method.
func (gp *Fpdf) DrawPath(styleStr string) {
	gp.currentContent().AppendStreamDrawPath(styleStr)
}

// Ln performs a line break. The current abscissa goes back to the left margin
// and the ordinate increases by the amount passed in parameter.
// This method is demonstrated in the example for MultiCell.
func (gp *Fpdf) Ln(h float64) {
	gp.ln(h, true)
}

func (gp *Fpdf) ln(h float64, toLeftMargin bool) {
	gp.PointsToUnitsVar(&h)
	if toLeftMargin {
		gp.curr.X = gp.margins.Left
	}

	if gp.curr.Y+h > gp.bottomMarginHeight() {
		page := gp.currentPage()
		gp.addPageWithOption(page.pageOption)
	}

	gp.curr.Y += gp.curr.setLineHeight(h)
}

// Circle draws a circle centered on point (x, y) with radius r.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the circle's perimeter.
// Filling uses the current fill color.
func (gp *Fpdf) Circle(x, y, r float64, styleStr string) {
	gp.Ellipse(x, y, r, r, 0, styleStr)
}

// Ellipse draws an ellipse centered at point (x, y). rx and ry specify its
// horizontal and vertical radii.
//
// degRotate specifies the counter-clockwise angle in degrees that the ellipse
// will be rotated.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color and line width centered on the ellipse's perimeter.
// Filling uses the current fill color.
//
// The Circle() example demonstrates this method.
func (gp *Fpdf) Ellipse(x, y, rx, ry, degRotate float64, styleStr string) {
	gp.UnitsToPointsVar(&x, &y, &rx, &ry)
	gp.currentContent().AppendStreamArcTo(x, y, rx, ry, degRotate, 0, 360, styleStr, false)
}

// Arc draws an elliptical arc centered at point (x, y). rx and ry specify its
// horizontal and vertical radii.
//
// degRotate specifies the angle that the arc will be rotated. degStart and
// degEnd specify the starting and ending angle of the arc. All angles are
// specified in degrees and measured counter-clockwise from the 3 o'clock
// position.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the arc's
// path. Filling uses the current fill color.
//
// The Circle() example demonstrates this method.
func (gp *Fpdf) Arc(x, y, rx, ry, degRotate, degStart, degEnd float64, styleStr string) {
	gp.UnitsToPointsVar(&x, &y, &rx, &ry)
	gp.currentContent().AppendStreamArcTo(x, y, rx, ry, degRotate, degStart, degEnd, styleStr, true)

}

// ArcTo draws an elliptical arc centered at point (x, y). rx and ry specify its
// horizontal and vertical radii. If the start of the arc is not at
// the current position, a connecting line will be drawn.
//
// degRotate specifies the angle that the arc will be rotated. degStart and
// degEnd specify the starting and ending angle of the arc. All angles are
// specified in degrees and measured counter-clockwise from the 3 o'clock
// position.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the arc's
// path. Filling uses the current fill color.
//
// The MoveTo() example demonstrates this method.
func (gp *Fpdf) ArcTo(x, y, rx, ry, degRotate, degStart, degEnd float64) {
	gp.UnitsToPointsVar(&x, &y, &rx, &ry)
	gp.currentContent().AppendStreamArcTo(x, y, rx, ry, degRotate, degStart, degEnd, "", true)
}

//Line : draw line
func (gp *Fpdf) Line(x1 float64, y1 float64, x2 float64, y2 float64) {
	gp.UnitsToPointsVar(&x1, &y1, &x2, &y2)
	gp.currentContent().AppendStreamLine(x1, y1, x2, y2)
}

// MoveTo moves the stylus to (x, y) without drawing the path from the
// previous point. Paths must start with a MoveTo to set the original
// stylus location or the result is undefined.
//
// Create a "path" by moving a virtual stylus around the page (with
// MoveTo, LineTo, CurveTo, CurveBezierCubicTo, ArcTo & ClosePath)
// then draw it or  fill it in (with DrawPath). The main advantage of
// using the path drawing routines rather than multiple Fpdf.Line is
// that PDF creates nice line joins at the angles, rather than just
// overlaying the lines.
func (gp *Fpdf) MoveTo(x, y float64) {
	gp.UnitsToPointsVar(&x, &y)
	gp.currentContent().AppendStreamPoint(x, y)
	gp.curr.X, gp.curr.Y = x, y
}

// LineTo creates a line from the current stylus location to (x, y), which
// becomes the new stylus location. Note that this only creates the line in
// the path; it does not actually draw the line on the page.
//
// The MoveTo() example demonstrates this method.
func (gp *Fpdf) LineTo(x, y float64) {
	gp.UnitsToPointsVar(&x, &y)
	gp.currentContent().AppendStreamLineTo(x, y)
}

//RectFromLowerLeft : draw rectangle from lower-left corner (x, y)
func (gp *Fpdf) RectFromLowerLeft(x float64, y float64, wdth float64, hght float64) {
	gp.UnitsToPointsVar(&x, &y, &wdth, &hght)
	gp.currentContent().AppendStreamRectangle(x, y, wdth, hght, "")
}

//RectFromUpperLeft : draw rectangle from upper-left corner (x, y)
func (gp *Fpdf) RectFromUpperLeft(x float64, y float64, wdth float64, hght float64) {
	gp.UnitsToPointsVar(&x, &y, &wdth, &hght)
	gp.currentContent().AppendStreamRectangle(x, y+hght, wdth, hght, "")
}

//RectFromLowerLeftWithStyle : draw rectangle from lower-left corner (x, y)
// - style: Style of rectangule (draw and/or fill: D, F, DF, FD)
//		D or empty string: draw. This is the default value.
//		F: fill
//		DF or FD: draw and fill
func (gp *Fpdf) RectFromLowerLeftWithStyle(x float64, y float64, wdth float64, hght float64, style string) {
	gp.UnitsToPointsVar(&x, &y, &wdth, &hght)
	gp.currentContent().AppendStreamRectangle(x, y, wdth, hght, style)
}

//RectFromUpperLeftWithStyle : draw rectangle from upper-left corner (x, y)
// - style: Style of rectangule (draw and/or fill: D, F, DF, FD)
//		D or empty string: draw. This is the default value.
//		F: fill
//		DF or FD: draw and fill
func (gp *Fpdf) RectFromUpperLeftWithStyle(x float64, y float64, wdth float64, hght float64, style string) {
	gp.UnitsToPointsVar(&x, &y, &wdth, &hght)
	gp.currentContent().AppendStreamRectangle(x, y+hght, wdth, hght, style)
}

//Oval : draw oval
func (gp *Fpdf) Oval(x1 float64, y1 float64, x2 float64, y2 float64) {
	gp.UnitsToPointsVar(&x1, &y1, &x2, &y2)
	gp.currentContent().AppendStreamOval(x1, y1, x2, y2)
}

//SetGrayFill set the grayscale for the fill, takes a float64 between 0.0 and 1.0
func (gp *Fpdf) SetGrayFill(grayScale float64) {
	gp.curr.grayFill = grayScale
	gp.currentContent().AppendStreamSetGrayFill(grayScale)
}

//SetGrayStroke set the grayscale for the stroke, takes a float64 between 0.0 and 1.0
func (gp *Fpdf) SetGrayStroke(grayScale float64) {
	gp.curr.grayStroke = grayScale
	gp.currentContent().AppendStreamSetGrayStroke(grayScale)
}

//SetX : set current position X
func (gp *Fpdf) SetX(x float64) {
	gp.UnitsToPointsVar(&x)
	gp.curr.setXCount++
	gp.curr.X = x
}

//X : get current position X
func (gp *Fpdf) X() float64 {
	return gp.PointsToUnits(gp.curr.X)
}

//SetY : set current position y
func (gp *Fpdf) SetY(y float64) {
	gp.UnitsToPointsVar(&y)
	gp.curr.Y = y
}

// Y : get current position y
func (gp *Fpdf) Y() float64 {
	return gp.PointsToUnits(gp.curr.Y)
}

// XY gets the current x and y position
func (gp *Fpdf) XY() (float64, float64) {
	return gp.X(), gp.Y()
}

// SetXY sets both x and y
func (gp *Fpdf) SetXY(x, y float64) {
	gp.SetX(x)
	gp.SetY(y)
}

//ImageByHolder : draw image by ImageHolder
func (gp *Fpdf) ImageByHolder(img ImageHolder, x float64, y float64, rect Rect) error {
	gp.UnitsToPointsVar(&x, &y)
	rect = rect.UnitsToPoints(gp.curr.unit)

	return gp.imageByHolder(img, x, y, rect)
}

func (gp *Fpdf) ImageByObj(img *ImageObj, x float64, y float64, rect Rect) error {
	gp.UnitsToPointsVar(&x, &y)
	rect = rect.UnitsToPoints(gp.curr.unit)

	cacheImageIndex, _, err := gp.registerImageByImageObj(img)
	if err != nil {
		return err
	}

	gp.currentContent().AppendStreamImage(cacheImageIndex, x, y, rect)
	return nil
}

func (gp *Fpdf) imageByHolder(img ImageHolder, x float64, y float64, rect Rect) error {
	cacheImageIndex, _, err := gp.registerImageByHolder(img)
	if err != nil {
		return err
	}

	gp.currentContent().AppendStreamImage(cacheImageIndex, x, y, rect)
	return nil
}

// CreateTemplate defines a new template using the current page size.
func (gp *Fpdf) CreateTemplate(fn TplFunc) (Template, error) {
	return newTpl(Point{0, 0}, gp.appliedOpts, fn, gp)
}

// CreateTemplateCustom starts a template, using the given bounds.
func (gp *Fpdf) CreateTemplateCustom(corner Point, fn TplFunc, opts ...PdfOption) (Template, error) {
	corner = corner.ToPoints(gp.curr.unit)
	opts = append(gp.appliedOpts, opts...)
	return newTpl(corner, opts, fn, gp)
}

// CreateTemplate creates a template that is not attached to any document.
func CreateTemplate(corner Point, unit int, fn TplFunc, opts ...PdfOption) (Template, error) {
	corner = corner.ToPoints(unit)
	return newTpl(corner, opts, fn, nil)
}

// UseTemplate adds a template to the current page or another template,
// using the size and position at which it was originally written.
func (gp *Fpdf) UseTemplate(t Template) error {
	if t == nil {
		return errors.New("template is nil")
	}

	corner, size := t.Size()
	return gp.useTemplateScaled(t, corner, size)
}

// UseTemplateScaled adds a template to the current page or another template,
// using the given page coordinates.
func (gp *Fpdf) UseTemplateScaled(t Template, corner Point, size Rect) error {
	corner = corner.ToPoints(gp.curr.unit)
	size = size.UnitsToPoints(gp.curr.unit)

	return gp.useTemplateScaled(t, corner, size)
}

func (gp *Fpdf) useTemplateScaled(t Template, corner Point, size Rect) error {
	if t == nil {
		return errors.New("template is nil")
	}

	templates := t.Templates()
	for x := 0; x < len(templates); x++ {
		if _, _, err := gp.registerTpl(templates[x]); err != nil {
			return err
		}
	}

	imgs := t.Images()
	for x := 0; x < len(imgs); x++ {
		if _, _, err := gp.registerImageByImageObj(imgs[x]); err != nil {
			return err
		}
	}

	fonts := t.Fonts()
	for x := 0; x < len(fonts); x++ {
		if err := gp.AddTTFFontBySubsetFont(fonts[x].Family, fonts[x]); err != nil {
			return err
		}
	}

	id, _, err := gp.registerTpl(t)
	if err != nil {
		return err
	}
	_, tSize := t.Size()
	scalex := size.W / tSize.W
	scaley := size.H / tSize.H

	gp.currentContent().AppendStreamUseTemplate(id, corner.X, corner.Y, size.H, scalex, scaley)
	return nil
}

func (gp *Fpdf) registerTpl(template Template) (string, int, error) {
	//create img object
	tplObj := newTemplateObj(template, gp.protection(), func() *Fpdf {
		return gp
	})
	// I feel like there will be an error here at some point
	a, b := gp.addProcsetObj(tplObj)
	return a, b, nil
}

func (gp *Fpdf) procsetIndex(id string, isFont bool) int {
	if isFont {
		return gp.procset().Realtes.getIndex(id)
	}
	return gp.procset().RealteXobjs.getIndex(id)
}

func (gp *Fpdf) hasProcsetIndex(id string, isFont bool) bool {
	return gp.procsetIndex(id, isFont) != -1
}

func (gp *Fpdf) registerImageByHolder(img ImageHolder) (string, int, error) {
	//create img object
	imgobj, err := NewImageObj(img)
	if err != nil {
		return "", 0, err
	}

	return gp.registerImageByImageObj(imgobj)
}

func (gp *Fpdf) registerImageByImageObj(imgobj *ImageObj) (string, int, error) {
	if pid, id, ok := gp.pdfObjs.hasProcsetObj(imgobj); ok {
		return pid, id, nil
	}

	// sanity check that these exist, they are removed on serialization
	imgobj.getRoot = func() *Fpdf {
		return gp
	}
	imgobj.setProtection(gp.protection())

	if imgobj.haveSMask() {
		smaskObj, err := imgobj.createSMask()
		if err != nil {
			return "", 0, err
		}
		imgobj.imginfo.smarkObjID = gp.addObj(smaskObj)
	}

	if imgobj.isColspaceIndexed() {
		dRGB, err := imgobj.createDeviceRGB()
		if err != nil {
			return "", 0, err
		}
		dRGB.getRoot = func() *Fpdf {
			return gp
		}
		imgobj.imginfo.deviceRGBObjID = gp.addObj(dRGB)
	}

	pid, id := gp.addProcsetObj(imgobj)
	return pid, id, nil
}

//Image : draw image
func (gp *Fpdf) Image(picPath string, x float64, y float64, rect Rect) error {
	gp.UnitsToPointsVar(&x, &y)
	rect = rect.UnitsToPoints(gp.curr.unit)

	imgh, err := ImageHolderByPath(picPath)
	if err != nil {
		return err
	}
	return gp.imageByHolder(imgh, x, y, rect)
}

// ImageByReader adds an image to the pdf with a reader
func (gp *Fpdf) ImageByReader(r io.Reader, x float64, y float64, rect Rect) error {
	gp.UnitsToPointsVar(&x, &y)
	rect = rect.UnitsToPoints(gp.curr.unit)

	imgh, err := newImageBuffByReader(r)
	if err != nil {
		return err
	}

	return gp.imageByHolder(imgh, x, y, rect)
}

// ImageByURL adds an image to the pdf using the given url
func (gp *Fpdf) ImageByURL(url string, x float64, y float64, rect Rect) error {
	gp.UnitsToPointsVar(&x, &y)
	rect = rect.UnitsToPoints(gp.curr.unit)

	imgh, err := newImageBuffByURL(url)
	if err != nil {
		return err
	}

	return gp.imageByHolder(imgh, x, y, rect)
}

//AddPage : add new page
func (gp *Fpdf) AddPage() {
	var po PageOption

	if !gp.curr.pageOption.IsEmpty() {
		po = gp.curr.pageOption
	}

	if cp := gp.currentPage(); cp != nil && !cp.pageOption.IsEmpty() {
		po = cp.pageOption
	}

	gp.AddPageWithOption(po)
}

//AddPageWithOption  : add new page with option
func (gp *Fpdf) AddPageWithOption(opt PageOption) {
	gp.addPageWithOption(opt)
}

func (gp *Fpdf) addPageWithOption(opt PageOption) {
	page := new(PageObj)
	page.init(func() *Fpdf {
		return gp
	})

	page.setOption(gp.curr.pageOption.merge(opt))

	page.ResourcesRelate = fmt.Sprintf("%d 0 R", gp.pdfObjs.indexOfFirst(procSetType)+1)
	gp.pdfObjs.addPageObj(page)
	gp.resetCurrXY()
}

//New creates a new Fpdf Object
func New(opts ...PdfOption) (*Fpdf, error) {
	gp := new(Fpdf)
	gp.init()

	for x := 0; x < len(opts); x++ {
		if err := opts[x].apply(gp); err != nil {
			return nil, err
		}
	}
	gp.appliedOpts = opts

	//สร้าง obj พื้นฐาน
	catalog := new(CatalogObj)
	catalog.init(func() *Fpdf {
		return gp
	})
	pages := new(PagesObj)
	pages.init(func() *Fpdf {
		return gp
	})
	gp.addObj(catalog)
	gp.addObj(pages)
	_ = gp.pdfObjs.procset()

	return gp, nil
}

// SetFontWithStyle : set font style support Regular or Underline
// for Bold|Italic should be loaded apropriate fonts with same styles defined
func (gp *Fpdf) SetFontWithStyle(family string, style int, size float64) error {
	if sub := gp.pdfObjs.getSubsetFontObjByFamilyAndStyle(family, style&^Underline); sub != nil {
		gp.curr.Font_Size = size
		gp.curr.Font_Style = style
		gp.curr.Font_FontCount = sub.CountOfFont
		gp.curr.Font_ISubset = sub
	} else {
		return fmt.Errorf("Could not find font with family: \"%s\" and style \"%d\"", family, style)
	}

	return nil
}

//SetFont : set font style support "" or "U"
// for "B" and "I" should be loaded apropriate fonts with same styles defined
func (gp *Fpdf) SetFont(family string, style string, size float64) error {
	return gp.SetFontWithStyle(family, getConvertedStyle(style), size)
}

//WritePdf : wirte pdf file
func (gp *Fpdf) WritePdf(pdfPath string) error {
	return ioutil.WriteFile(pdfPath, gp.GetBytesPdf(), 0644)
}

func (gp *Fpdf) Write(w io.Writer) error {
	return gp.compilePdf(w)
}

func (gp *Fpdf) Read(p []byte) (int, error) {
	if gp.buf.Len() == 0 && gp.buf.Cap() == 0 {
		if err := gp.compilePdf(&gp.buf); err != nil {
			return 0, err
		}
	}
	return gp.buf.Read(p)
}

// Close closes the pdf buffer
func (gp *Fpdf) Close() error {
	gp.buf = bytes.Buffer{}
	return nil
}

func (gp *Fpdf) compilePdf(w io.Writer) error {

	if gp.isUseProtection() {
		encObj := gp.pdfProtection.encryptionObj()
		gp.addObj(encObj)
	}

	err := gp.Close()
	if err != nil {
		return err
	}
	writer := newCountingWriter(w)
	//io.WriteString(w, "%PDF-1.7\n\n")
	fmt.Fprint(writer, "%PDF-1.7\n\n")
	linelens, err := gp.pdfObjs.write(writer)
	gp.xref(writer, writer.offset, linelens, len(linelens))
	return nil
}

type (
	countingWriter struct {
		offset int
		writer io.Writer
	}
)

func newCountingWriter(w io.Writer) *countingWriter {
	return &countingWriter{writer: w}
}

func (cw *countingWriter) Write(b []byte) (int, error) {
	n, err := cw.writer.Write(b)
	cw.offset += n
	return n, err
}

//GetBytesPdfReturnErr : get bytes of pdf file
func (gp *Fpdf) GetBytesPdfReturnErr() ([]byte, error) {
	err := gp.Close()
	if err != nil {
		return nil, err
	}
	err = gp.compilePdf(&gp.buf)
	return gp.buf.Bytes(), err
}

//GetBytesPdf : get bytes of pdf file
func (gp *Fpdf) GetBytesPdf() []byte {
	b, err := gp.GetBytesPdfReturnErr()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	return b
}

//Text write text start at current x,y ( current y is the baseline of text )
func (gp *Fpdf) Text(x, y float64, text string) error {
	gp.SetXY(x, y)

	err := gp.curr.Font_ISubset.AddChars(text)
	if err != nil {
		return err
	}

	err = gp.currentContent().AppendStreamText(text)
	if err != nil {
		return err
	}

	return nil
}

// WriteTextf is the same as WriteText but it uses fmt.Sprintf
func (gp *Fpdf) WriteTextf(h float64, txtStr string, v ...interface{}) error {
	return gp.WriteText(h, fmt.Sprintf(txtStr, v...))
}

// WriteTextOptsf is the same as WriteText but it uses fmt.Sprintf
func (gp *Fpdf) WriteTextOptsf(h float64, txtStr string, opts CellOption, textOpts TextOption, v ...interface{}) error {
	return gp.WriteTextOpts(h, fmt.Sprintf(txtStr, v...), opts, textOpts)
}

// WriteText prints text from the current position. When the right margin is
// reached (or the \n character is met) a line break occurs and text continues
// from the left margin. Upon method exit, the current position is left just at
// the end of the text.
//
// It is possible to put a link on the text.
//
// h indicates the line height in the unit of measure specified in New().
func (gp *Fpdf) WriteText(h float64, txtStr string) error {
	return gp.MultiCell(0, h, txtStr)
}

// WriteText prints text from the current position. When the right margin is
// reached (or the \n character is met) a line break occurs and text continues
// from the left margin. Upon method exit, the current position is left just at
// the end of the text.
//
// It is possible to put a link on the text.
//
// h indicates the line height in the unit of measure specified in New().
func (gp *Fpdf) WriteTextOpts(h float64, txtStr string, opts CellOption, textOpts TextOption) error {
	return gp.MultiCellOpts(0, h, txtStr, opts, textOpts)
}

// MultiCell supports printing text with line breaks. They can be automatic (as
// soon as the text reaches the right border of the cell) or explicit (via the
// \n character). As many cells as necessary are output, one below the other.
//
// Text can be aligned, centered or justified. The cell block can be framed and
// the background painted. See CellFormat() for more details.
//
// The current position after calling MultiCell() is the beginning of the next
// line, equivalent to calling CellFormat with ln equal to 1.
//
// w is the width of the cells. A value of zero indicates cells that reach to
// the right margin.
//
// h indicates the line height of each cell in the unit of measure specified in New().
func (gp *Fpdf) MultiCell(w, h float64, txtStr string) error {
	defaultopt := CellOption{
		Align:  Left | Top,
		Border: 0,
		Float:  Bottom,
	}

	return gp.MultiCellOpts(w, h, txtStr, defaultopt, TextOption{})
}

// MultiCell supports printing text with line breaks. They can be automatic (as
// soon as the text reaches the right border of the cell) or explicit (via the
// \n character). As many cells as necessary are output, one below the other.
//
// Text can be aligned, centered or justified. The cell block can be framed and
// the background painted. See CellFormat() for more details.
//
// The current position after calling MultiCell() is the beginning of the next
// line, equivalent to calling CellFormat with ln equal to 1.
//
// w is the width of the cells. A value of zero indicates cells that reach to
// the right margin.
//
// h indicates the line height of each cell in the unit of measure specified in New().
func (gp *Fpdf) MultiCellOpts(w, h float64, txtStr string, opts CellOption, textOpts TextOption) error {
	gp.UnitsToPointsVar(&w, &h)
	gp.curr.setLineHeight(h)

	if w == 0 {
		w = gp.rightMarginWidth(gp.curr.X)
	}

	if w <= 0 {
		return errors.New("Cell has a zero or negative width, something is wrong")
	}

	lines, err := gp.splitLines(txtStr, w, textOpts)
	if err != nil {
		return err
	}

	rectangle := Rect{W: w, H: h}

	for x := 0; x < len(lines); x++ {
		if gp.curr.Y+h > gp.bottomMarginHeight() {
			page := gp.currentPage()
			gp.addPageWithOption(page.pageOption)
		}

		err = gp.cellWithOption(rectangle, lines[x], opts, textOpts)

		if err != nil {
			return err
		}
	}

	return nil
}

func (gp *Fpdf) rightMarginWidth(leftOffset float64) float64 {
	return gp.GetBoundaryWidth(PageBoundaryMedia) - gp.margins.Right - leftOffset
}

func (gp *Fpdf) bottomMarginHeight() float64 {
	return gp.GetBoundaryHeight(PageBoundaryMedia) - gp.margins.Bottom
}

func (gp *Fpdf) splitLines(txt string, w float64, textOpts TextOption) ([]string, error) {
	var final []string
	nlb := strings.Split(txt, "\n")

	for x := 0; x < len(nlb); x++ {
		buffer := nlb[x]

		for {
			var line string
			var err error

			line, buffer, err = gp.cutStringBefore(buffer, w, textOpts)
			if err != nil {
				return final, err
			}

			final = append(final, line)
			if buffer == "" {
				break
			}
		}
	}

	return final, nil
}

func (gp *Fpdf) cutStringBefore(txtStr string, w float64, textOpt TextOption) (line string, left string, err error) {
	r := regexp.MustCompile("[^\\s]*\\s*")
	words := r.FindAllString(txtStr, -1)

	for y := 0; y < len(words); y++ {
		var tw float64
		tw, err = gp.measureTextWidth(fmt.Sprintf("%s%s", line, words[y]), Unit_PT, textOpt)

		if err != nil {
			return
		}

		if tw > w {
			if line == "" && left != "" {
				err = fmt.Errorf("width not large enough to fit anything")
				return
			}

			left = strings.Join(words[y:], "")
			break
		}

		line = fmt.Sprintf("%s%s", line, words[y])
	}

	return
}

//CellWithOption create cell of text ( use current x,y is upper-left corner of cell)
func (gp *Fpdf) CellWithOption(w, h float64, text string, opt CellOption, textOpts TextOption) error {
	gp.UnitsToPointsVar(&w, &h)
	rectangle := Rect{W: w, H: h}

	return gp.cellWithOption(rectangle, text, opt, textOpts)
}

func (gp *Fpdf) cellWithOption(rect Rect, text string, opt CellOption, textOpts TextOption) error {
	err := gp.curr.Font_ISubset.AddChars(text)
	if err != nil {
		return err
	}

	err = gp.currentContent().AppendStreamSubsetFont(rect, text, opt, textOpts)
	return err
}

//Cellf : same as cell but using go's Sprintf format
func (gp *Fpdf) Cellf(w, h float64, text string, args ...interface{}) error {
	return gp.Cell(w, h, fmt.Sprintf(text, args...))
}

//Cell : create cell of text ( use current x,y is upper-left corner of cell)
//Note that this has no effect on Rect.H pdf (now). Fix later :-)
func (gp *Fpdf) Cell(w, h float64, text string) error {
	gp.UnitsToPointsVar(&w, &h)
	rectangle := Rect{W: w, H: h}

	defaultopt := CellOption{
		Align:  Left | Top,
		Border: 0,
		Float:  Right,
	}

	return gp.cellWithOption(rectangle, text, defaultopt, TextOption{})
}

//AddLink
func (gp *Fpdf) AddExternalLink(url string, x, y, w, h float64) {
	gp.UnitsToPointsVar(&x, &y, &w, &h)
	page := gp.currentPage()
	page.Links = append(page.Links, linkOption{x, gp.GetBoundaryHeight(PageBoundaryMedia) - y, w, h, url, ""})
}

func (gp *Fpdf) AddInternalLink(anchor string, x, y, w, h float64) {
	gp.UnitsToPointsVar(&x, &y, &w, &h)
	page := gp.currentPage()
	page.Links = append(page.Links, linkOption{x, gp.GetBoundaryHeight(PageBoundaryMedia) - y, w, h, "", anchor})
}

func (gp *Fpdf) SetAnchor(name string) {
	y := gp.GetBoundaryHeight(PageBoundaryMedia) - gp.curr.Y + float64(gp.curr.Font_Size)
	gp.anchors[name] = anchorOption{gp.curr.IndexOfPageObj, y}
}

//AddTTFFontByReader add font file
func (gp *Fpdf) AddTTFFontByReader(family string, rd io.Reader) error {
	return gp.AddTTFFontByReaderWithOption(family, rd, defaultTtfFontOption())
}

//AddTTFFontByReaderWithOption add font file
func (gp *Fpdf) AddTTFFontByReaderWithOption(family string, rd io.Reader, option TtfOption) error {
	subsetFont, err := SubsetFontByReaderWithOption(rd, option)
	if err != nil {
		return err
	}

	return gp.AddTTFFontBySubsetFont(family, subsetFont)
}

func SubsetFontByReader(rd io.Reader) (*SubsetFontObj, error) {

	return SubsetFontByReaderWithOption(rd, defaultTtfFontOption())
}

func SubsetFontByReaderWithOption(rd io.Reader, option TtfOption) (*SubsetFontObj, error) {
	subsetFont := new(SubsetFontObj)
	subsetFont.SetTtfFontOption(option)
	subsetFont.CharacterToGlyphIndex = NewMapOfCharacterToGlyphIndex()
	err := subsetFont.SetTTFByReader(rd)
	return subsetFont, err
}

func (gp *Fpdf) AddTTFFontBySubsetFont(family string, subsetFont *SubsetFontObj) error {
	if _, id, ok := gp.pdfObjs.hasProcsetObj(subsetFont); ok {
		actualSubsetFont := gp.pdfObjs.getSubsetFont(id)
		actualSubsetFont.AddChars(subsetFont.CharacterToGlyphIndex.AllKeysString())
		return nil
	}

	subsetFont.SetFamily(family)

	subsetFont.init(func() *Fpdf {
		return gp
	})

	unicodemap := new(UnicodeMap)
	unicodemap.init(func() *Fpdf {
		return gp
	})
	unicodemap.setProtection(gp.protection())
	unicodemap.SetPtrToSubsetFontObj(subsetFont)
	unicodeindex := gp.addObj(unicodemap)

	pdfdic := new(PdfDictionaryObj)
	pdfdic.init(func() *Fpdf {
		return gp
	})
	pdfdic.setProtection(gp.protection())
	pdfdic.SetPtrToSubsetFontObj(subsetFont)
	pdfdicindex := gp.addObj(pdfdic)

	subfontdesc := new(SubfontDescriptorObj)
	subfontdesc.init(func() *Fpdf {
		return gp
	})
	subfontdesc.SetPtrToSubsetFontObj(subsetFont)
	subfontdesc.SetIndexObjPdfDictionary(pdfdicindex)
	subfontdescindex := gp.addObj(subfontdesc)

	cidfont := new(CIDFontObj)
	cidfont.init(func() *Fpdf {
		return gp
	})
	cidfont.SetPtrToSubsetFontObj(subsetFont)
	cidfont.SetIndexObjSubfontDescriptor(subfontdescindex)
	cidindex := gp.addObj(cidfont)

	subsetFont.SetIndexObjCIDFont(cidindex)
	subsetFont.SetIndexObjUnicodeMap(unicodeindex)
	gp.addProcsetObj(subsetFont) //add หลังสุด
	return nil
}

//AddTTFFontWithOption : add font file
func (gp *Fpdf) AddTTFFontWithOption(family string, ttfpath string, option TtfOption) error {

	if _, err := os.Stat(ttfpath); os.IsNotExist(err) {
		return err
	}
	data, err := ioutil.ReadFile(ttfpath)
	if err != nil {
		return err
	}
	rd := bytes.NewReader(data)
	return gp.AddTTFFontByReaderWithOption(family, rd, option)
}

//AddTTFFont : add font file
func (gp *Fpdf) AddTTFFont(family string, ttfpath string) error {
	return gp.AddTTFFontWithOption(family, ttfpath, defaultTtfFontOption())
}

//KernOverride override kern value
func (gp *Fpdf) KernOverride(family string, fn FuncKernOverride) error {
	fonts := gp.pdfObjs.allOf(subsetFontType)
	max := len(fonts)
	for i := 0; i < max; i++ {
		sub := fonts[i].(*SubsetFontObj)
		if sub.GetFamily() == family {
			sub.funcKernOverride = fn
			return nil
		}
	}
	return errors.New("font family not found")
}

//SetTextColor :  function sets the text color
func (gp *Fpdf) SetTextColor(r uint8, g uint8, b uint8) {
	rgb := Rgb{
		r: r,
		g: g,
		b: b,
	}
	gp.curr.setTextColor(rgb)
}

//SetRBStrokeColor set the color for the stroke
func (gp *Fpdf) SetRGBStrokeColor(r uint8, g uint8, b uint8) {
	gp.currentContent().AppendStreamSetRGBColorStroke(r, g, b)
}

//SetRGBFillColor set the color for the stroke
func (gp *Fpdf) SetRGBFillColor(r uint8, g uint8, b uint8) {
	gp.currentContent().AppendStreamSetRGBColorFill(r, g, b)
}

//SetCMYKStrokeColor set the color for the stroke
func (gp *Fpdf) SetCMYKStrokeColor(c, m, y, k uint8) {
	gp.currentContent().AppendStreamSetCMYKColorStroke(c, m, y, k)
}

//SetCMYKFillColor set the color for the stroke
func (gp *Fpdf) SetCMYKFillColor(c, m, y, k uint8) {
	gp.currentContent().AppendStreamSetCMYKColorFill(c, m, y, k)
}

//MeasureTextWidth : measure Width of text (use current font)
func (gp *Fpdf) MeasureTextWidth(text string, textOpt TextOption) (float64, error) {
	return gp.measureTextWidth(text, gp.curr.unit, textOpt)
}

func (gp *Fpdf) measureTextWidth(text string, units int, textOpt TextOption) (float64, error) {
	err := gp.curr.Font_ISubset.AddChars(text) //AddChars for create CharacterToGlyphIndex
	if err != nil {
		return 0, err
	}

	_, _, textWidthPdfUnit, err := createContent(gp.curr.Font_ISubset, text, gp.curr.Font_Size, nil, textOpt)
	if err != nil {
		return 0, err
	}
	return PointsToUnits(units, textWidthPdfUnit), nil
}

// CurveTo : marks a curve from the x, y position to the new x, y position
func (gp *Fpdf) CurveTo(cx, cy, x, y float64) {
	gp.UnitsToPointsVar(&cx, &cy, &x, &y)
	gp.currentContent().AppendStreamCurve(cx, cy, x, y)
}

// Curve draws a single-segment quadratic Bézier curve. The curve starts at
// the point (x0, y0) and ends at the point (x1, y1). The control point (cx,
// cy) specifies the curvature. At the start point, the curve is tangent to the
// straight line between the start point and the control point. At the end
// point, the curve is tangent to the straight line between the end point and
// the control point.
//
// styleStr can be "F" for filled, "D" for outlined only, or "DF" or "FD" for
// outlined and filled. An empty string will be replaced with "D". Drawing uses
// the current draw color, line width, and cap style centered on the curve's
// path. Filling uses the current fill color.
//
// The Circle() example demonstrates this method.
func (gp *Fpdf) Curve(x0, y0, cx, cy, x1, y1 float64, styleStr string) {
	gp.UnitsToPointsVar(&x0, &y0, &cx, &cy, &x1, &y1)
	gp.currentContent().AppendStreamPoint(x0, y0)
	gp.currentContent().AppendStreamCurve(cx, cy, x1, y1)
	gp.currentContent().AppendStreamDrawPath(styleStr)
}

/*
//SetProtection set permissions as well as user and owner passwords
func (gp *Fpdf) SetProtection(permissions int, userPass []byte, ownerPass []byte) {
	gp.pdfProtection = new(PDFProtection)
	gp.pdfProtection.setProtection(permissions, userPass, ownerPass)
}*/

//Rotate rotate text or image
// angle is angle in degrees.
// x, y is rotation center
func (gp *Fpdf) Rotate(angle, x, y float64) {
	gp.UnitsToPointsVar(&x, &y)
	gp.currentContent().appendRotate(angle, x, y)
}

//RotateReset reset rotate
func (gp *Fpdf) RotateReset() {
	gp.currentContent().appendRotateReset()
}

/*---private---*/

//init
func (gp *Fpdf) init() {
	//default
	gp.margins = Margins{
		Left:   defaultMargin,
		Top:    defaultMargin,
		Right:  defaultMargin,
		Bottom: defaultMargin,
	}

	//init curr
	gp.resetCurrXY()
	gp.curr.IndexOfPageObj = -1
	gp.curr.CountOfFont = 0
	gp.curr.CountOfL = 0
	gp.curr.CountOfImg = 0 //img
	gp.anchors = make(map[string]anchorOption)

	//init index
	gp.pdfObjs = newPdfObjs(func() *Fpdf {
		return gp
	})

	//No underline
	//gp.IsUnderline = false
	gp.curr.lineWidth = 1
	// default to zlib.DefaultCompression
	gp.compressLevel = zlib.DefaultCompression
	// default units is points
	gp.curr.unit = Unit_PT
	gp.curr.pageOption.AddPageBoundary(NewPageSizeBoundary(Unit_PT, PageSizeA4.W, PageSizeA4.H))
}

func (gp *Fpdf) resetCurrXY() {
	gp.curr.X = gp.margins.Left
	gp.curr.Y = gp.margins.Top
}

func (gp *Fpdf) isUseProtection() bool {
	return gp.pdfProtection != nil
}

func (gp *Fpdf) protection() *PDFProtection {
	return gp.pdfProtection
}

func (gp *Fpdf) xref(w io.Writer, xrefbyteoffset int, linelens []int, i int) error {

	io.WriteString(w, "xref\n")
	fmt.Fprintf(w, "0 %d\n", i+1)
	io.WriteString(w, "0000000000 65535 f \n")
	j := 0
	max := len(linelens)
	for j < max {
		linelen := linelens[j]
		fmt.Fprintf(w, "%s 00000 n \n", gp.formatXrefline(linelen))
		j++
	}
	io.WriteString(w, "trailer\n")
	io.WriteString(w, "<<\n")
	fmt.Fprintf(w, "/Size %d\n", max+1)
	io.WriteString(w, "/Root 1 0 R\n")
	if id := gp.pdfObjs.indexOfFirst(encryptionType); id >= 0 {
		fmt.Fprintf(w, "/Encrypt %d 0 R\n", id+1)
		io.WriteString(w, "/ID [()()]\n")
	}
	gp.GetInfo().write(w)
	io.WriteString(w, ">>\n")
	io.WriteString(w, "startxref\n")
	fmt.Fprintf(w, "%d", xrefbyteoffset)
	io.WriteString(w, "\n%%EOF\n")

	return nil
}

//ปรับ xref ให้เป็น 10 หลัก
func (gp *Fpdf) formatXrefline(n int) string {
	str := strconv.Itoa(n)
	for len(str) < 10 {
		str = "0" + str
	}
	return str
}

func (gp *Fpdf) addObj(iobj IObj) int {
	return gp.pdfObjs.addObj(iobj)
}

func (gp *Fpdf) addProcsetObj(iobj ProcsetIObj) (string, int) {
	return gp.pdfObjs.addProcsetObj(iobj)
}

func (gp *Fpdf) currentPage() *PageObj {
	return gp.pdfObjs.currentPage()
}

func (gp *Fpdf) currentContent() *ContentObj {
	return gp.pdfObjs.currentContent()
}

func (gp *Fpdf) procset() *ProcSetObj {
	return gp.pdfObjs.procset()
}

func (gp *Fpdf) getAllPages() []*PageObj {
	return gp.pdfObjs.allPages()
}

func (gp *Fpdf) getAllImages() []*ImageObj {
	return gp.pdfObjs.allImages()
}

func (gp *Fpdf) getAllSubsetFonts() []*SubsetFontObj {
	return gp.pdfObjs.allSubsetFonts()
}

func (gp *Fpdf) getTemplates() ([]Template, error) {
	ts := make([]Template, 0)
	templates := gp.pdfObjs.allOf(templateType)

	for x := 0; x < len(templates); x++ {
		template := templates[x].(*TemplateObj)
		ts = append(ts, template.ToTemplate())
	}

	return ts, nil

}

// UnitsToPoints converts the units to the documents unit type
func (gp *Fpdf) UnitsToPoints(u float64) float64 {
	return UnitsToPoints(gp.curr.unit, u)
}

// UnitsToPointsVar converts the units to the documents unit type for all variables passed in
func (gp *Fpdf) UnitsToPointsVar(u ...*float64) {
	UnitsToPointsVar(gp.curr.unit, u...)
}

// PointsToUnits converts the points to the documents unit type
func (gp *Fpdf) PointsToUnits(u float64) float64 {
	return PointsToUnits(gp.curr.unit, u)
}

// PointsToUnits converts the points to the documents unit type for all variables passed in
func (gp *Fpdf) PointsToUnitsVar(u ...*float64) {
	PointsToUnitsVar(gp.curr.unit, u...)
}

func encodeUtf8(str string) string {
	var buff bytes.Buffer
	for _, r := range str {
		c := fmt.Sprintf("%X", r)
		for len(c) < 4 {
			c = "0" + c
		}
		buff.WriteString(c)
	}
	return buff.String()
}

func infodate(t time.Time) string {
	ft := t.Format("20060102150405-07'00'")
	return ft
}

//tool for validate pdf https://www.pdf-online.com/osa/validate.aspx
