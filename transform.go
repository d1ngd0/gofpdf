package gofpdf

import (
	"fmt"
	"math"
)

// Routines in this file are translated from the work of Moritz Wagner and
// Andreas Würmser.

// TransformMatrix is used for generalized transformations of text, drawings
// and images.
type TransformMatrix struct {
	A, B, C, D, E, F float64
}

// TransformBegin sets up a transformation context for subsequent text,
// drawings and images. The typical usage is to immediately follow a call to
// this method with a call to one or more of the transformation methods such as
// TransformScale(), TransformSkew(), etc. This is followed by text, drawing or
// image output and finally a call to TransformEnd(). All transformation
// contexts must be properly ended prior to outputting the document.
func (gp *Fpdf) TransformBegin() {
	gp.currentContent().AppendStreamTransformBegin()
}

// TransformEnd applies a transformation that was begun with a call to TransformBegin().
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformEnd() {
	gp.currentContent().AppendStreamTransformEnd()
}

// TransformScaleX scales the width of the following text, drawings and images.
// scaleWd is the percentage scaling factor. (x, y) is center of scaling.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformScaleX(scaleWd, x, y float64) {
	gp.TransformScale(scaleWd, 100, x, y)
}

// TransformScaleY scales the height of the following text, drawings and
// images. scaleHt is the percentage scaling factor. (x, y) is center of
// scaling.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformScaleY(scaleHt, x, y float64) {
	gp.TransformScale(100, scaleHt, x, y)
}

// TransformScaleXY uniformly scales the width and height of the following
// text, drawings and images. s is the percentage scaling factor for both width
// and height. (x, y) is center of scaling.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformScaleXY(s, x, y float64) {
	gp.TransformScale(s, s, x, y)
}

// TransformScale generally scales the following text, drawings and images.
// scaleWd and scaleHt are the percentage scaling factors for width and height.
// (x, y) is center of scaling.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformScale(scaleWd, scaleHt, x, y float64) error {
	gp.UnitsToPointsVar(&x, &y)

	if scaleWd == 0 || scaleHt == 0 {
		return fmt.Errorf("scale factor cannot be zero")
	}

	y = gp.GetBoundaryHeight(PageBoundaryMedia) - y

	scaleWd /= 100
	scaleHt /= 100

	gp.Transform(TransformMatrix{scaleWd, 0, 0,
		scaleHt, x * (1 - scaleWd), y * (1 - scaleHt)})
	return nil
}

// TransformMirrorHorizontal horizontally mirrors the following text, drawings
// and images. x is the axis of reflection.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformMirrorHorizontal(x float64) {
	gp.TransformScale(-100, 100, x, gp.curr.Y)
}

// TransformMirrorVertical vertically mirrors the following text, drawings and
// images. y is the axis of reflection.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformMirrorVertical(y float64) {
	gp.TransformScale(100, -100, gp.curr.X, y)
}

// TransformMirrorPoint symmetrically mirrors the following text, drawings and
// images on the point specified by (x, y).
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformMirrorPoint(x, y float64) {
	gp.TransformScale(-100, -100, x, y)
}

// TransformMirrorLine symmetrically mirrors the following text, drawings and
// images on the line defined by angle and the point (x, y). angles is
// specified in degrees and measured counter-clockwise from the 3 o'clock
// position.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformMirrorLine(angle, x, y float64) {
	gp.TransformScale(-100, 100, x, y)
	gp.TransformRotate(-2*(angle-90), x, y)
}

// TransformTranslateX moves the following text, drawings and images
// horizontally by the amount specified by tx.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformTranslateX(tx float64) {
	gp.TransformTranslate(tx, 0)
}

// TransformTranslateY moves the following text, drawings and images vertically
// by the amount specified by ty.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformTranslateY(ty float64) {
	gp.TransformTranslate(0, ty)
}

// TransformTranslate moves the following text, drawings and images
// horizontally and vertically by the amounts specified by tx and ty.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformTranslate(tx, ty float64) {
	gp.Transform(TransformMatrix{1, 0, 0, 1, gp.UnitsToPoints(tx), gp.UnitsToPoints(-ty)})
}

// TransformRotate rotates the following text, drawings and images around the
// center point (x, y). angle is specified in degrees and measured
// counter-clockwise from the 3 o'clock position.
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformRotate(angle, x, y float64) {
	gp.UnitsToPointsVar(&x, &y)
	y = gp.GetBoundaryHeight(PageBoundaryMedia) - y
	angle = angle * math.Pi / 180
	var tm TransformMatrix
	tm.A = math.Cos(angle)
	tm.B = math.Sin(angle)
	tm.C = -tm.B
	tm.D = tm.A
	tm.E = x + tm.B*y - tm.A*x
	tm.F = y - tm.A*y - tm.B*x
	gp.Transform(tm)
}

// TransformSkewX horizontally skews the following text, drawings and images
// keeping the point (x, y) stationary. angleX ranges from -90 degrees (skew to
// the left) to 90 degrees (skew to the right).
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformSkewX(angleX, x, y float64) error {
	return gp.TransformSkew(angleX, 0, x, y)
}

// TransformSkewY vertically skews the following text, drawings and images
// keeping the point (x, y) stationary. angleY ranges from -90 degrees (skew to
// the bottom) to 90 degrees (skew to the top).
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformSkewY(angleY, x, y float64) error {
	return gp.TransformSkew(0, angleY, x, y)
}

// TransformSkew generally skews the following text, drawings and images
// keeping the point (x, y) stationary. angleX ranges from -90 degrees (skew to
// the left) to 90 degrees (skew to the right). angleY ranges from -90 degrees
// (skew to the bottom) to 90 degrees (skew to the top).
//
// The TransformBegin() example demonstrates this method.
func (gp *Fpdf) TransformSkew(angleX, angleY, x, y float64) error {
	if angleX <= -90 || angleX >= 90 || angleY <= -90 || angleY >= 90 {
		return fmt.Errorf("skew values must be between -90° and 90°")
	}

	gp.UnitsToPointsVar(&x, &y)
	y = gp.GetBoundaryHeight(PageBoundaryMedia) - y

	var tm TransformMatrix
	tm.A = 1
	tm.B = math.Tan(angleY * math.Pi / 180)
	tm.C = math.Tan(angleX * math.Pi / 180)
	tm.D = 1
	tm.E = -tm.C * y
	tm.F = -tm.B * x
	gp.Transform(tm)

	return nil
}

// Transform generally transforms the following text, drawings and images
// according to the specified matrix. It is typically easier to use the various
// methods such as TransformRotate() and TransformMirrorVertical() instead.
func (gp *Fpdf) Transform(tm TransformMatrix) {
	gp.currentContent().AppendStreamTransform(tm)
}
