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
	"strconv"
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

	pdfObjs []IObj
	config  Config
	anchors map[string]anchorOption

	/*---index ของ obj สำคัญๆ เก็บเพื่อลด loop ตอนค้นหา---*/
	//index ของ obj pages
	indexOfPagesObj int

	//index ของ obj page อันแรก
	indexOfFirstPageObj int

	//ต่ำแหน่งปัจจุบัน
	curr Current

	indexEncodingObjFonts []int
	indexOfContent        int

	//index ของ procset ซึ่งควรจะมีอันเดียว
	indexOfProcSet int

	//IsUnderline bool

	// Buffer for io.Reader compliance
	buf bytes.Buffer

	//pdf PProtection
	pdfProtection   *PDFProtection
	encryptionObjID int

	// content streams only
	compressLevel int

	//info
	isUseInfo bool
	info      *PdfInfo
}

//SetLineWidth : set line width
func (gp *Fpdf) SetLineWidth(width float64) {
	gp.curr.lineWidth = gp.UnitsToPoints(width)
	gp.getContent().AppendStreamSetLineWidth(gp.UnitsToPoints(width))
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
	gp.getContent().AppendStreamSetLineType(linetype)
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
	points := pts.ToPoints(gp.config.Unit)

	if len(points) < 4 {
		return fmt.Errorf("the number of points can not be less than 4. %d found", len(points))
	}

	gp.getContent().AppendStreamPoint(points[0].XY())

	points = points[1:]
	for len(points) >= 3 {
		cx0, cy0 := points[0].XY()
		cx1, cy1 := points[1].XY()
		x1, y1 := points[2].XY()
		gp.getContent().AppendStreamCurveBezierCubic(cx0, cy0, cx1, cy1, x1, y1)
		points = points[3:]
	}

	gp.getContent().AppendStreamDrawPath(styleStr)
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
	gp.getContent().AppendStreamPoint(x0, y0)
	gp.getContent().AppendStreamCurveBezierCubic(cx0, cy0, cx1, cy1, x1, y1)
	gp.getContent().AppendStreamDrawPath(styleStr)
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
	gp.getContent().AppendStreamCurveBezierCubic(cx0, cy0, cx1, cy1, x, y)
	gp.curr.X, gp.curr.Y = x, y
}

// ClosePath creates a line from the current location to the last MoveTo point
// (if not the same) and mark the path as closed so the first and last lines
// join nicely.
//
// The MoveTo() example demonstrates this method.
func (gp *Fpdf) ClosePath() {
	gp.getContent().AppendStreamClosePath()
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
	gp.getContent().AppendStreamDrawPath(styleStr)
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
	gp.getContent().AppendStreamArcTo(x, y, rx, ry, degRotate, 0, 360, styleStr, false)
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
	gp.getContent().AppendStreamArcTo(x, y, rx, ry, degRotate, degStart, degEnd, styleStr, true)

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
	gp.getContent().AppendStreamArcTo(x, y, rx, ry, degRotate, degStart, degEnd, "", true)
}

//Line : draw line
func (gp *Fpdf) Line(x1 float64, y1 float64, x2 float64, y2 float64) {
	gp.UnitsToPointsVar(&x1, &y1, &x2, &y2)
	gp.getContent().AppendStreamLine(x1, y1, x2, y2)
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
	gp.getContent().AppendStreamPoint(x, y)
	gp.curr.X, gp.curr.Y = x, y
}

// LineTo creates a line from the current stylus location to (x, y), which
// becomes the new stylus location. Note that this only creates the line in
// the path; it does not actually draw the line on the page.
//
// The MoveTo() example demonstrates this method.
func (gp *Fpdf) LineTo(x, y float64) {
	gp.UnitsToPointsVar(&x, &y)
	gp.getContent().AppendStreamLineTo(x, y)
}

//RectFromLowerLeft : draw rectangle from lower-left corner (x, y)
func (gp *Fpdf) RectFromLowerLeft(x float64, y float64, wdth float64, hght float64) {
	gp.UnitsToPointsVar(&x, &y, &wdth, &hght)
	gp.getContent().AppendStreamRectangle(x, y, wdth, hght, "")
}

//RectFromUpperLeft : draw rectangle from upper-left corner (x, y)
func (gp *Fpdf) RectFromUpperLeft(x float64, y float64, wdth float64, hght float64) {
	gp.UnitsToPointsVar(&x, &y, &wdth, &hght)
	gp.getContent().AppendStreamRectangle(x, y+hght, wdth, hght, "")
}

//RectFromLowerLeftWithStyle : draw rectangle from lower-left corner (x, y)
// - style: Style of rectangule (draw and/or fill: D, F, DF, FD)
//		D or empty string: draw. This is the default value.
//		F: fill
//		DF or FD: draw and fill
func (gp *Fpdf) RectFromLowerLeftWithStyle(x float64, y float64, wdth float64, hght float64, style string) {
	gp.UnitsToPointsVar(&x, &y, &wdth, &hght)
	gp.getContent().AppendStreamRectangle(x, y, wdth, hght, style)
}

//RectFromUpperLeftWithStyle : draw rectangle from upper-left corner (x, y)
// - style: Style of rectangule (draw and/or fill: D, F, DF, FD)
//		D or empty string: draw. This is the default value.
//		F: fill
//		DF or FD: draw and fill
func (gp *Fpdf) RectFromUpperLeftWithStyle(x float64, y float64, wdth float64, hght float64, style string) {
	gp.UnitsToPointsVar(&x, &y, &wdth, &hght)
	gp.getContent().AppendStreamRectangle(x, y+hght, wdth, hght, style)
}

//Oval : draw oval
func (gp *Fpdf) Oval(x1 float64, y1 float64, x2 float64, y2 float64) {
	gp.UnitsToPointsVar(&x1, &y1, &x2, &y2)
	gp.getContent().AppendStreamOval(x1, y1, x2, y2)
}

//Br : new line
func (gp *Fpdf) Br(h float64) {
	gp.UnitsToPointsVar(&h)
	gp.curr.Y += h
	gp.curr.X = gp.margins.Left
}

//SetGrayFill set the grayscale for the fill, takes a float64 between 0.0 and 1.0
func (gp *Fpdf) SetGrayFill(grayScale float64) {
	gp.curr.grayFill = grayScale
	gp.getContent().AppendStreamSetGrayFill(grayScale)
}

//SetGrayStroke set the grayscale for the stroke, takes a float64 between 0.0 and 1.0
func (gp *Fpdf) SetGrayStroke(grayScale float64) {
	gp.curr.grayStroke = grayScale
	gp.getContent().AppendStreamSetGrayStroke(grayScale)
}

//SetX : set current position X
func (gp *Fpdf) SetX(x float64) {
	gp.UnitsToPointsVar(&x)
	gp.curr.setXCount++
	gp.curr.X = x
}

//GetX : get current position X
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

func (gp *Fpdf) SetXY(x, y float64) {
	gp.SetX(x)
	gp.SetY(y)
}

//ImageByHolder : draw image by ImageHolder
func (gp *Fpdf) ImageByHolder(img ImageHolder, x float64, y float64, rect *Rect) error {
	gp.UnitsToPointsVar(&x, &y)
	rect = rect.UnitsToPoints(gp.config.Unit)

	return gp.imageByHolder(img, x, y, rect)
}

func (gp *Fpdf) imageByHolder(img ImageHolder, x float64, y float64, rect *Rect) error {
	cacheImageIndex := -1
	for _, imgcache := range gp.curr.ImgCaches {
		if img.ID() == imgcache.Path {
			cacheImageIndex = imgcache.Index
			break
		}
	}

	if cacheImageIndex == -1 { //new image

		//create img object
		imgobj := new(ImageObj)
		imgobj.init(func() *Fpdf {
			return gp
		})
		imgobj.setProtection(gp.protection())

		err := imgobj.SetImage(img)
		if err != nil {
			return err
		}

		var imgRect *Rect
		if rect == nil {
			imgRect, err = imgobj.getRect()
			if err != nil {
				return err
			}
		} else {
			imgRect = rect
		}

		err = imgobj.parse()
		if err != nil {
			return err
		}
		index := gp.addObj(imgobj)
		if gp.indexOfProcSet != -1 {
			//ยัดรูป
			procset := gp.pdfObjs[gp.indexOfProcSet].(*ProcSetObj)
			gp.getContent().AppendStreamImage(gp.curr.CountOfImg, x, y, imgRect)
			procset.RealteXobjs = append(procset.RealteXobjs, RealteXobject{IndexOfObj: index})
			//เก็บข้อมูลรูปเอาไว้
			var imgcache ImageCache
			imgcache.Index = gp.curr.CountOfImg
			imgcache.Path = img.ID()
			imgcache.Rect = imgRect
			gp.curr.ImgCaches = append(gp.curr.ImgCaches, imgcache)
			gp.curr.CountOfImg++
		}

		if imgobj.haveSMask() {
			smaskObj, err := imgobj.createSMask()
			if err != nil {
				return err
			}
			imgobj.imginfo.smarkObjID = gp.addObj(smaskObj)
		}

		if imgobj.isColspaceIndexed() {
			dRGB, err := imgobj.createDeviceRGB()
			if err != nil {
				return err
			}
			dRGB.getRoot = func() *Fpdf {
				return gp
			}
			imgobj.imginfo.deviceRGBObjID = gp.addObj(dRGB)
		}

	} else { //same img
		var imgRect *Rect
		if rect == nil {
			imgRect = gp.curr.ImgCaches[cacheImageIndex].Rect
		} else {
			imgRect = rect
		}
		gp.getContent().AppendStreamImage(cacheImageIndex, x, y, imgRect)
	}
	return nil
}

//Image : draw image
func (gp *Fpdf) Image(picPath string, x float64, y float64, rect *Rect) error {
	gp.UnitsToPointsVar(&x, &y)
	rect = rect.UnitsToPoints(gp.config.Unit)

	imgh, err := ImageHolderByPath(picPath)
	if err != nil {
		return err
	}
	return gp.imageByHolder(imgh, x, y, rect)
}

//AddPage : add new page
func (gp *Fpdf) AddPage() {
	emptyOpt := PageOption{}
	gp.AddPageWithOption(emptyOpt)
}

//AddPageWithOption  : add new page with option
func (gp *Fpdf) AddPageWithOption(opt PageOption) {
	opt.PageSize = opt.PageSize.UnitsToPoints(gp.config.Unit)

	page := new(PageObj)
	page.init(func() *Fpdf {
		return gp
	})

	if !opt.isEmpty() { //use page option
		page.setOption(opt)
		gp.curr.pageSize = opt.PageSize
	} else { //use default
		gp.curr.pageSize = gp.config.PageSize
	}

	page.ResourcesRelate = strconv.Itoa(gp.indexOfProcSet+1) + " 0 R"
	index := gp.addObj(page)
	if gp.indexOfFirstPageObj == -1 {
		gp.indexOfFirstPageObj = index
	}
	gp.curr.IndexOfPageObj = index

	//reset
	gp.indexOfContent = -1
	gp.resetCurrXY()
}

//New creates a new Fpdf Object
func New(config Config) *Fpdf {
	gp := new(Fpdf)
	gp.config = config
	gp.init()
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
	gp.indexOfPagesObj = gp.addObj(pages)

	//indexOfProcSet
	procset := new(ProcSetObj)
	procset.init(func() *Fpdf {
		return gp
	})
	gp.indexOfProcSet = gp.addObj(procset)

	if gp.isUseProtection() {
		gp.pdfProtection = gp.createProtection()
	}

	return gp
}

// SetFontWithStyle : set font style support Regular or Underline
// for Bold|Italic should be loaded apropriate fonts with same styles defined
func (gp *Fpdf) SetFontWithStyle(family string, style int, size int) error {

	found := false
	i := 0
	max := len(gp.pdfObjs)
	for i < max {
		if gp.pdfObjs[i].getType() == subsetFont {
			obj := gp.pdfObjs[i]
			sub, ok := obj.(*SubsetFontObj)
			if ok {
				if sub.GetFamily() == family && sub.GetTtfFontOption().Style == style&^Underline {
					gp.curr.Font_Size = size
					gp.curr.Font_Style = style
					gp.curr.Font_FontCount = sub.CountOfFont
					gp.curr.Font_ISubset = sub
					found = true
					break
				}
			}
		}
		i++
	}

	if !found {
		return errors.New("not found font family")
	}

	return nil
}

//SetFont : set font style support "" or "U"
// for "B" and "I" should be loaded apropriate fonts with same styles defined
func (gp *Fpdf) SetFont(family string, style string, size int) error {
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

func (gp *Fpdf) Close() error {
	gp.buf = bytes.Buffer{}
	return nil
}

func (gp *Fpdf) compilePdf(w io.Writer) error {
	gp.prepare()
	err := gp.Close()
	if err != nil {
		return err
	}
	max := len(gp.pdfObjs)
	writer := newCountingWriter(w)
	//io.WriteString(w, "%PDF-1.7\n\n")
	fmt.Fprint(writer, "%PDF-1.7\n\n")
	linelens := make([]int, max)
	i := 0

	for i < max {
		objID := i + 1
		linelens[i] = writer.offset
		pdfObj := gp.pdfObjs[i]
		fmt.Fprintf(writer, "%d 0 obj\n", objID)
		pdfObj.write(writer, objID)
		io.WriteString(writer, "endobj\n\n")
		i++
	}
	gp.xref(writer, writer.offset, linelens, i)
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

	err = gp.getContent().AppendStreamText(text)
	if err != nil {
		return err
	}

	return nil
}

//CellWithOptionf : same as CellWithOption but using go's Sprintf format
func (gp *Fpdf) CellWithOptionf(rectangle *Rect, text string, opt CellOption, args ...interface{}) error {
	return gp.CellWithOption(rectangle, fmt.Sprintf(text, args...), opt)
}

//CellWithOption create cell of text ( use current x,y is upper-left corner of cell)
func (gp *Fpdf) CellWithOption(rectangle *Rect, text string, opt CellOption) error {
	rectangle = rectangle.UnitsToPoints(gp.config.Unit)
	err := gp.curr.Font_ISubset.AddChars(text)
	if err != nil {
		return err
	}
	err = gp.getContent().AppendStreamSubsetFont(rectangle, text, opt)
	if err != nil {
		return err
	}
	return nil
}

//Cellf : same as cell but using go's Sprintf format
func (gp *Fpdf) Cellf(rectangle *Rect, text string, args ...interface{}) error {
	return gp.Cell(rectangle, fmt.Sprintf(text, args...))
}

//Cell : create cell of text ( use current x,y is upper-left corner of cell)
//Note that this has no effect on Rect.H pdf (now). Fix later :-)
func (gp *Fpdf) Cell(rectangle *Rect, text string) error {
	rectangle = rectangle.UnitsToPoints(gp.config.Unit)

	defaultopt := CellOption{
		Align:  Left | Top,
		Border: 0,
		Float:  Right,
	}

	err := gp.curr.Font_ISubset.AddChars(text)
	if err != nil {
		return err
	}
	err = gp.getContent().AppendStreamSubsetFont(rectangle, text, defaultopt)
	if err != nil {
		return err
	}

	return nil
}

//AddLink
func (gp *Fpdf) AddExternalLink(url string, x, y, w, h float64) {
	gp.UnitsToPointsVar(&x, &y, &w, &h)
	page := gp.pdfObjs[gp.curr.IndexOfPageObj].(*PageObj)
	page.Links = append(page.Links, linkOption{x, gp.config.PageSize.H - y, w, h, url, ""})
}

func (gp *Fpdf) AddInternalLink(anchor string, x, y, w, h float64) {
	gp.UnitsToPointsVar(&x, &y, &w, &h)
	page := gp.pdfObjs[gp.curr.IndexOfPageObj].(*PageObj)
	page.Links = append(page.Links, linkOption{x, gp.config.PageSize.H - y, w, h, "", anchor})
}

func (gp *Fpdf) SetAnchor(name string) {
	y := gp.config.PageSize.H - gp.curr.Y + float64(gp.curr.Font_Size)
	gp.anchors[name] = anchorOption{gp.curr.IndexOfPageObj, y}
}

//AddTTFFontByReader add font file
func (gp *Fpdf) AddTTFFontByReader(family string, rd io.Reader) error {
	return gp.AddTTFFontByReaderWithOption(family, rd, defaultTtfFontOption())
}

//AddTTFFontByReaderWithOption add font file
func (gp *Fpdf) AddTTFFontByReaderWithOption(family string, rd io.Reader, option TtfOption) error {
	subsetFont := new(SubsetFontObj)
	subsetFont.init(func() *Fpdf {
		return gp
	})
	subsetFont.SetTtfFontOption(option)
	subsetFont.SetFamily(family)
	err := subsetFont.SetTTFByReader(rd)
	if err != nil {
		return err
	}

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
	index := gp.addObj(subsetFont) //add หลังสุด

	if gp.indexOfProcSet != -1 {
		procset := gp.pdfObjs[gp.indexOfProcSet].(*ProcSetObj)
		if !procset.Realtes.IsContainsFamilyAndStyle(family, option.Style&^Underline) {
			procset.Realtes = append(procset.Realtes, RelateFont{Family: family, IndexOfObj: index, CountOfFont: gp.curr.CountOfFont, Style: option.Style &^ Underline})
			subsetFont.CountOfFont = gp.curr.CountOfFont
			gp.curr.CountOfFont++
		}
	}
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
	i := 0
	max := len(gp.pdfObjs)
	for i < max {
		if gp.pdfObjs[i].getType() == subsetFont {
			obj := gp.pdfObjs[i]
			sub, ok := obj.(*SubsetFontObj)
			if ok {
				if sub.GetFamily() == family {
					sub.funcKernOverride = fn
					return nil
				}
			}
		}
		i++
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
	gp.getContent().AppendStreamSetRGBColorStroke(r, g, b)
}

//SetRGBFillColor set the color for the stroke
func (gp *Fpdf) SetRGBFillColor(r uint8, g uint8, b uint8) {
	gp.getContent().AppendStreamSetRGBColorFill(r, g, b)
}

//SetCMYKStrokeColor set the color for the stroke
func (gp *Fpdf) SetCMYKStrokeColor(c, m, y, k uint8) {
	gp.getContent().AppendStreamSetCMYKColorStroke(c, m, y, k)
}

//SetCMYKFillColor set the color for the stroke
func (gp *Fpdf) SetCMYKFillColor(c, m, y, k uint8) {
	gp.getContent().AppendStreamSetCMYKColorFill(c, m, y, k)
}

//MeasureTextWidth : measure Width of text (use current font)
func (gp *Fpdf) MeasureTextWidth(text string, units int) (float64, error) {

	err := gp.curr.Font_ISubset.AddChars(text) //AddChars for create CharacterToGlyphIndex
	if err != nil {
		return 0, err
	}

	_, _, textWidthPdfUnit, err := createContent(gp.curr.Font_ISubset, text, gp.curr.Font_Size, nil)
	if err != nil {
		return 0, err
	}
	return PointsToUnits(units, textWidthPdfUnit), nil
}

// CurveTo : marks a curve from the x, y position to the new x, y position
func (gp *Fpdf) CurveTo(cx, cy, x, y float64) {
	gp.UnitsToPointsVar(&cx, &cy, &x, &y)
	gp.getContent().AppendStreamCurve(cx, cy, x, y)
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
	gp.getContent().AppendStreamPoint(x0, y0)
	gp.getContent().AppendStreamCurve(cx, cy, x1, y1)
	gp.getContent().AppendStreamDrawPath(styleStr)
}

/*
//SetProtection set permissions as well as user and owner passwords
func (gp *Fpdf) SetProtection(permissions int, userPass []byte, ownerPass []byte) {
	gp.pdfProtection = new(PDFProtection)
	gp.pdfProtection.setProtection(permissions, userPass, ownerPass)
}*/

//SetInfo set Document Information Dictionary
func (gp *Fpdf) SetInfo(info PdfInfo) {
	gp.info = &info
	gp.isUseInfo = true
}

//Rotate rotate text or image
// angle is angle in degrees.
// x, y is rotation center
func (gp *Fpdf) Rotate(angle, x, y float64) {
	gp.UnitsToPointsVar(&x, &y)
	gp.getContent().appendRotate(angle, x, y)
}

//RotateReset reset rotate
func (gp *Fpdf) RotateReset() {
	gp.getContent().appendRotateReset()
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
	gp.curr.ImgCaches = *new([]ImageCache)
	gp.anchors = make(map[string]anchorOption)

	//init index
	gp.indexOfPagesObj = -1
	gp.indexOfFirstPageObj = -1
	gp.indexOfContent = -1

	//No underline
	//gp.IsUnderline = false
	gp.curr.lineWidth = 1

	// default to zlib.DefaultCompression
	gp.compressLevel = zlib.DefaultCompression

	// change the unit type
	gp.config.PageSize = gp.config.PageSize.UnitsToPoints(gp.config.Unit)
}

func (gp *Fpdf) resetCurrXY() {
	gp.curr.X = gp.margins.Left
	gp.curr.Y = gp.margins.Top
}

func (gp *Fpdf) isUseProtection() bool {
	return gp.config.Protection.UseProtection
}

func (gp *Fpdf) createProtection() *PDFProtection {
	var prot PDFProtection
	prot.setProtection(
		gp.config.Protection.Permissions,
		gp.config.Protection.UserPass,
		gp.config.Protection.OwnerPass,
	)
	return &prot
}

func (gp *Fpdf) protection() *PDFProtection {
	return gp.pdfProtection
}

func (gp *Fpdf) prepare() {

	if gp.isUseProtection() {
		encObj := gp.pdfProtection.encryptionObj()
		gp.addObj(encObj)
	}

	if gp.indexOfPagesObj != -1 {
		indexCurrPage := -1
		var pagesObj *PagesObj
		pagesObj = gp.pdfObjs[gp.indexOfPagesObj].(*PagesObj)
		i := 0 //gp.indexOfFirstPageObj
		max := len(gp.pdfObjs)
		for i < max {
			objtype := gp.pdfObjs[i].getType()
			//fmt.Printf(" objtype = %s , %d \n", objtype , i)
			if objtype == "Page" {
				pagesObj.Kids = fmt.Sprintf("%s %d 0 R ", pagesObj.Kids, i+1)
				pagesObj.PageCount++
				indexCurrPage = i
			} else if objtype == "Content" {
				if indexCurrPage != -1 {
					gp.pdfObjs[indexCurrPage].(*PageObj).Contents = fmt.Sprintf("%s %d 0 R ", gp.pdfObjs[indexCurrPage].(*PageObj).Contents, i+1)
				}
			} else if objtype == "Font" {
				tmpfont := gp.pdfObjs[i].(*FontObj)
				j := 0
				jmax := len(gp.indexEncodingObjFonts)
				for j < jmax {
					tmpencoding := gp.pdfObjs[gp.indexEncodingObjFonts[j]].(*EncodingObj).GetFont()
					if tmpfont.Family == tmpencoding.GetFamily() { //ใส่ ข้อมูลของ embed font
						tmpfont.IsEmbedFont = true
						tmpfont.SetIndexObjEncoding(gp.indexEncodingObjFonts[j] + 1)
						tmpfont.SetIndexObjWidth(gp.indexEncodingObjFonts[j] + 2)
						tmpfont.SetIndexObjFontDescriptor(gp.indexEncodingObjFonts[j] + 3)
						break
					}
					j++
				}
			} else if objtype == "Encryption" {
				gp.encryptionObjID = i + 1
			}
			i++
		}
	}
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
	if gp.isUseProtection() {
		fmt.Fprintf(w, "/Encrypt %d 0 R\n", gp.encryptionObjID)
		io.WriteString(w, "/ID [()()]\n")
	}
	if gp.isUseInfo {
		gp.writeInfo(w)
	}
	io.WriteString(w, ">>\n")
	io.WriteString(w, "startxref\n")
	fmt.Fprintf(w, "%d", xrefbyteoffset)
	io.WriteString(w, "\n%%EOF\n")

	return nil
}

func (gp *Fpdf) writeInfo(w io.Writer) {
	var zerotime time.Time
	io.WriteString(w, "/Info <<\n")

	if gp.info.Author != "" {
		fmt.Fprintf(w, "/Author <FEFF%s>\n", encodeUtf8(gp.info.Author))
	}

	if gp.info.Title != "" {
		fmt.Fprintf(w, "/Title <FEFF%s>\n", encodeUtf8(gp.info.Title))
	}

	if gp.info.Subject != "" {
		fmt.Fprintf(w, "/Subject <FEFF%s>\n", encodeUtf8(gp.info.Subject))
	}

	if gp.info.Creator != "" {
		fmt.Fprintf(w, "/Creator <FEFF%s>\n", encodeUtf8(gp.info.Creator))
	}

	if gp.info.Producer != "" {
		fmt.Fprintf(w, "/Producer <FEFF%s>\n", encodeUtf8(gp.info.Producer))
	}

	if !zerotime.Equal(gp.info.CreationDate) {
		fmt.Fprintf(w, "/CreationDate(D:%s)>>\n", infodate(gp.info.CreationDate))
	}

	io.WriteString(w, " >>\n")
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
	index := len(gp.pdfObjs)
	gp.pdfObjs = append(gp.pdfObjs, iobj)
	return index
}

func (gp *Fpdf) getContent() *ContentObj {
	var content *ContentObj
	if gp.indexOfContent <= -1 {
		content = new(ContentObj)
		content.init(func() *Fpdf {
			return gp
		})
		gp.indexOfContent = gp.addObj(content)
	} else {
		content = gp.pdfObjs[gp.indexOfContent].(*ContentObj)
	}
	return content
}

// UnitsToPoints converts the units to the documents unit type
func (gp *Fpdf) UnitsToPoints(u float64) float64 {
	return UnitsToPoints(gp.config.Unit, u)
}

// UnitsToPointsVar converts the units to the documents unit type for all variables passed in
func (gp *Fpdf) UnitsToPointsVar(u ...*float64) {
	UnitsToPointsVar(gp.config.Unit, u...)
}

// PointsToUnits converts the points to the documents unit type
func (gp *Fpdf) PointsToUnits(u float64) float64 {
	return PointsToUnits(gp.config.Unit, u)
}

// PointsToUnits converts the points to the documents unit type for all variables passed in
func (gp *Fpdf) PointsToUnitsVar(u ...*float64) {
	PointsToUnitsVar(gp.config.Unit, u...)
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
