package gofpdf

import (
	"compress/zlib"
	"fmt"
	"io"
)

//ContentObj content object
type ContentObj struct { //impl IObj
	listCache listCacheContent
	//text bytes.Buffer
	getRoot func() *Fpdf
}

func (c *ContentObj) protection() *PDFProtection {
	return c.getRoot().protection()
}

func (c *ContentObj) init(funcGetRoot func() *Fpdf) {
	c.getRoot = funcGetRoot
}

func (c *ContentObj) write(w io.Writer, objID int) error {
	buff := GetBuffer()
	defer PutBuffer(buff)

	isFlate := (c.getRoot().compressLevel != zlib.NoCompression)
	if isFlate {
		ww, err := zlib.NewWriterLevel(buff, c.getRoot().compressLevel)
		if err != nil {
			// should never happen...
			return err
		}
		if err := c.listCache.write(ww, c.protection()); err != nil {
			return err
		}
		ww.Close()
	} else {
		if err := c.listCache.write(buff, c.protection()); err != nil {
			return err
		}
	}

	streamlen := buff.Len()

	io.WriteString(w, "<<\n")
	if isFlate {
		io.WriteString(w, "/Filter/FlateDecode")
	}
	fmt.Fprintf(w, "/Length %d\n", streamlen)
	io.WriteString(w, ">>\n")
	io.WriteString(w, "stream\n")
	if c.protection() != nil {
		tmp, err := rc4Cip(c.protection().objectkey(objID), buff.Bytes())
		if err != nil {
			return err
		}
		w.Write(tmp)
		io.WriteString(w, "\n")
	} else {
		buff.WriteTo(w)
		if isFlate {
			io.WriteString(w, "\n")
		}
	}
	io.WriteString(w, "endstream\n")

	return nil
}

func (c *ContentObj) getType() string {
	return "Content"
}

//AppendStreamText append text
func (c *ContentObj) AppendStreamText(text string) error {

	//support only CURRENT_FONT_TYPE_SUBSET
	textColor := c.getRoot().curr.textColor()
	grayFill := c.getRoot().curr.grayFill
	fontCountIndex := c.getRoot().curr.Font_FontCount + 1
	fontSize := c.getRoot().curr.Font_Size
	fontStyle := c.getRoot().curr.Font_Style
	x := c.getRoot().curr.X
	y := c.getRoot().curr.Y
	setXCount := c.getRoot().curr.setXCount
	fontSubset := c.getRoot().curr.Font_ISubset

	cache := cacheContentText{
		fontSubset:     fontSubset,
		rectangle:      nil,
		textColor:      textColor,
		grayFill:       grayFill,
		fontCountIndex: fontCountIndex,
		fontSize:       fontSize,
		fontStyle:      fontStyle,
		setXCount:      setXCount,
		x:              x,
		y:              y,
		pageheight:     c.getRoot().curr.pageSize.H,
		contentType:    ContentTypeText,
		lineWidth:      c.getRoot().curr.lineWidth,
	}

	var err error
	c.getRoot().curr.X, c.getRoot().curr.Y, err = c.listCache.appendContentText(cache, text)
	if err != nil {
		return err
	}

	return nil
}

//AppendStreamSubsetFont add stream of text
func (c *ContentObj) AppendStreamSubsetFont(rectangle *Rect, text string, cellOpt CellOption) error {

	textColor := c.getRoot().curr.textColor()
	grayFill := c.getRoot().curr.grayFill
	fontCountIndex := c.getRoot().curr.Font_FontCount + 1
	fontSize := c.getRoot().curr.Font_Size
	fontStyle := c.getRoot().curr.Font_Style
	x := c.getRoot().curr.X
	y := c.getRoot().curr.Y
	setXCount := c.getRoot().curr.setXCount
	fontSubset := c.getRoot().curr.Font_ISubset

	cache := cacheContentText{
		fontSubset:     fontSubset,
		rectangle:      rectangle,
		textColor:      textColor,
		grayFill:       grayFill,
		fontCountIndex: fontCountIndex,
		fontSize:       fontSize,
		fontStyle:      fontStyle,
		setXCount:      setXCount,
		x:              x,
		y:              y,
		pageheight:     c.getRoot().curr.pageSize.H,
		contentType:    ContentTypeCell,
		cellOpt:        cellOpt,
		lineWidth:      c.getRoot().curr.lineWidth,
	}
	var err error
	c.getRoot().curr.X, c.getRoot().curr.Y, err = c.listCache.appendContentText(cache, text)
	if err != nil {
		return err
	}
	return nil
}

// AppendStreamClosePath : appends a path closing
func (c *ContentObj) AppendStreamClosePath() {
	var cache cacheContentClosePath
	c.listCache.append(&cache)
}

// AppendStreamCurve : appends a curve
func (c *ContentObj) AppendStreamCurve(cx, cy, x, y float64) {
	var cache cacheContentCurve
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x = x
	cache.y = y
	cache.cx = cx
	cache.cy = cy
	c.listCache.append(&cache)
}

// AppendStreamDrawPath : appends a draw path with style string
func (c *ContentObj) AppendStreamDrawPath(styleStr string) {
	var cache cacheContentDrawPath
	cache.styleStr = styleStr
	c.listCache.append(&cache)
}

// AppendStreamArcTo : appends an arc to
func (c *ContentObj) AppendStreamArcTo(x, y, rx, ry, degRotate, degStart, degEnd float64, styleStr string, path bool) {
	var cache cacheContentArcTo
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.currentX = &c.getRoot().curr.X
	cache.currentY = &c.getRoot().curr.Y
	cache.x = x
	cache.y = y
	cache.rx = rx
	cache.ry = ry
	cache.degRotate = degRotate
	cache.degStart = degStart
	cache.degEnd = degEnd
	cache.styleStr = styleStr
	cache.path = path
	c.listCache.append(&cache)
}

// AppendStreamPoint appends a point
func (c *ContentObj) AppendStreamPoint(x, y float64) {
	var cache cacheContentPoint
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x = x
	cache.y = y
	c.listCache.append(&cache)
}

// AppendStreamLineTo appends a line to
func (c *ContentObj) AppendStreamLineTo(x, y float64) {
	var cache cacheContentLineTo
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x = x
	cache.y = y
	c.listCache.append(&cache)
}

//AppendStreamLine append line
func (c *ContentObj) AppendStreamLine(x1 float64, y1 float64, x2 float64, y2 float64) {
	//h := c.getRoot().config.PageSize.H
	//c.stream.WriteString(fmt.Sprintf("%0.2f %0.2f m %0.2f %0.2f l s\n", x1, h-y1, x2, h-y2))
	var cache cacheContentLine
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x1 = x1
	cache.y1 = y1
	cache.x2 = x2
	cache.y2 = y2
	c.listCache.append(&cache)
}

//AppendStreamRectangle : draw rectangle from lower-left corner (x, y) with specif width/height
func (c *ContentObj) AppendStreamRectangle(x float64, y float64, wdth float64, hght float64, style string) {
	var cache cacheContentRectangle
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x = x
	cache.y = y
	cache.width = wdth
	cache.height = hght
	cache.style = style
	c.listCache.append(&cache)
}

//AppendStreamOval append oval
func (c *ContentObj) AppendStreamOval(x1 float64, y1 float64, x2 float64, y2 float64) {
	var cache cacheContentOval
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x1 = x1
	cache.y1 = y1
	cache.x2 = x2
	cache.y2 = y2
	c.listCache.append(&cache)
}

// AppendStreamCurveBezierCubic : draw a curve from the current spot to the x and y position
func (c *ContentObj) AppendStreamCurveBezierCubic(cx0, cy0, cx1, cy1, x, y float64) {
	var cache cacheContentCurveBezierCubic
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.cx0 = cx0
	cache.cy0 = cy0
	cache.cx1 = cx1
	cache.cy1 = cy1
	cache.x = x
	cache.y = y
	c.listCache.append(&cache)
}

// AppendStreamTransformBegin starts a transformation in the current content
func (c *ContentObj) AppendStreamTransformBegin() {
	var cache cacheContentTransformBegin
	c.listCache.append(&cache)
}

// AppendStreamTransformEnd stops a transformation in the current content
func (c *ContentObj) AppendStreamTransformEnd() {
	var cache cacheContentTransformEnd
	c.listCache.append(&cache)
}

// AppendStreamTransform applies the transformation matrix to the current content
func (c *ContentObj) AppendStreamTransform(matrix TransformMatrix) {
	var cache cacheContentTransform
	cache.matrix = matrix
	c.listCache.append(&cache)
}

//AppendStreamClipBegin starts a new clip in the content
func (c *ContentObj) AppendStreamClipBegin() {
	var cache cacheContentClipBegin
	c.listCache.append(&cache)
}

//AppendStreamClipEnd ends a clipping path in the content
func (c *ContentObj) AppendStreamClipEnd() {
	var cache cacheContentClipEnd
	c.listCache.append(&cache)
}

//AppendStreamClipClipEllipse creates an Elliptical clipping path
func (c *ContentObj) AppendStreamClipPolygon(points []Point, style string) {
	var cache cacheContentClipPolygon
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.points = points
	cache.style = style

	c.listCache.append(new(cacheContentClipBegin))
	c.listCache.append(&cache)
}

//AppendStreamClipClipEllipse creates an Elliptical clipping path
func (c *ContentObj) AppendStreamClipEllipse(x, y, rx, ry float64, style string) {
	var cache cacheContentClipEllipse
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x = x
	cache.y = y
	cache.rx = rx
	cache.ry = ry
	cache.style = style

	c.listCache.append(new(cacheContentClipBegin))
	c.listCache.append(&cache)
}

//AppendStreamClipRoundedRect creates a rounded rectangle clipping path
func (c *ContentObj) AppendStreamClipRoundedRect(x, y, w, h, r float64, style string) {
	var cache cacheContentClipRoundedRect
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x = x
	cache.y = y
	cache.w = w
	cache.h = h
	cache.r = r
	cache.style = style

	c.listCache.append(new(cacheContentClipBegin))
	c.listCache.append(&cache)
}

//AppendStreamClipText creates a clipping path using the text
func (c *ContentObj) AppendStreamClipText(x, y float64, txtStr string, style int) {
	var cache cacheContentClipText
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x = x
	cache.y = y
	cache.txtStr = txtStr
	cache.style = style

	c.listCache.append(new(cacheContentClipBegin))
	c.listCache.append(&cache)
}

//AppendStreamClipRect starts a rectangle clipping path
func (c *ContentObj) AppendStreamClipRect(x, y, w, h float64, style string) {
	var cache cacheContentClipRect
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.x = x
	cache.y = y
	cache.w = w
	cache.h = h
	cache.style = style

	c.listCache.append(new(cacheContentClipBegin))
	c.listCache.append(&cache)
}

//AppendStreamSetLineWidth : set line width
func (c *ContentObj) AppendStreamSetLineWidth(w float64) {
	var cache cacheContentLineWidth
	cache.width = w
	c.listCache.append(&cache)
}

//AppendStreamSetLineType : Set linetype [solid, dashed, dotted]
func (c *ContentObj) AppendStreamSetLineType(t string) {
	var cache cacheContentLineType
	cache.lineType = t
	c.listCache.append(&cache)

}

//AppendStreamSetGrayFill  set the grayscale fills
func (c *ContentObj) AppendStreamSetGrayFill(w float64) {
	w = fixRange10(w)
	var cache cacheContentGray
	cache.grayType = grayTypeFill
	cache.scale = w
	c.listCache.append(&cache)
}

//AppendStreamSetGrayStroke  set the grayscale stroke
func (c *ContentObj) AppendStreamSetGrayStroke(w float64) {
	w = fixRange10(w)
	var cache cacheContentGray
	cache.grayType = grayTypeStroke
	cache.scale = w
	c.listCache.append(&cache)
}

//AppendStreamSetRGBColorStroke  set the color stroke using RGB
func (c *ContentObj) AppendStreamSetRGBColorStroke(r uint8, g uint8, b uint8) {
	var cache cacheContentColor
	cache.colorType = colorTypeStrokeRGB
	cache.r = r
	cache.g = g
	cache.b = b
	c.listCache.append(&cache)
}

//AppendStreamSetRGBColorFill  set the color fill using RB
func (c *ContentObj) AppendStreamSetRGBColorFill(r uint8, g uint8, b uint8) {
	var cache cacheContentColor
	cache.colorType = colorTypeFillRGB
	cache.r = r
	cache.g = g
	cache.b = b
	c.listCache.append(&cache)
}

//AppendStreamSetCMYKColorStroke  set the color stroke using CMYK
func (c *ContentObj) AppendStreamSetCMYKColorStroke(cy uint8, m uint8, y uint8, k uint8) {
	var cache cacheContentColor
	cache.colorType = colorTypeStrokeCMYK
	cache.c = cy
	cache.m = m
	cache.y = y
	cache.k = k
	c.listCache.append(&cache)
}

//AppendStreamSetCMYKColorFill  set the color fill using CMYK
func (c *ContentObj) AppendStreamSetCMYKColorFill(cy uint8, m uint8, y uint8, k uint8) {
	var cache cacheContentColor
	cache.colorType = colorTypeFillCMYK
	cache.c = cy
	cache.m = m
	cache.y = y
	cache.k = k
	c.listCache.append(&cache)
}

//AppendStreamImage append image
func (c *ContentObj) AppendStreamImage(index int, x float64, y float64, rect *Rect) {
	//fmt.Printf("index = %d",index)
	h := c.getRoot().curr.pageSize.H
	var cache cacheContentImage
	cache.h = h
	cache.x = x
	cache.y = y
	cache.rect = *rect
	cache.index = index
	c.listCache.append(&cache)
	//c.stream.WriteString(fmt.Sprintf("q %0.2f 0 0 %0.2f %0.2f %0.2f cm /I%d Do Q\n", rect.W, rect.H, x, h-(y+rect.H), index+1))
}

func (c *ContentObj) appendRotate(angle, x, y float64) {
	var cache cacheContentRotate
	cache.isReset = false
	cache.pageHeight = c.getRoot().curr.pageSize.H
	cache.angle = angle
	cache.x = x
	cache.y = y
	c.listCache.append(&cache)
}

func (c *ContentObj) appendRotateReset() {
	var cache cacheContentRotate
	cache.isReset = true
	c.listCache.append(&cache)
}

//ContentObj_CalTextHeight calculate height of text
func ContentObj_CalTextHeight(fontsize int) float64 {
	return (float64(fontsize) * 0.7)
}

// When setting colour and grayscales the value has to be between 0.00 and 1.00
// This function takes a float64 and returns 0.0 if it is less than 0.0 and 1.0 if it
// is more than 1.0
func fixRange10(val float64) float64 {
	if val < 0.0 {
		return 0.0
	}
	if val > 1.0 {
		return 1.0
	}
	return val
}

func convertTTFUnit2PDFUnit(n int, upem int) int {
	var ret int
	if n < 0 {
		rest1 := n % upem
		storrest := 1000 * rest1
		//ledd2 := (storrest != 0 ? rest1 / storrest : 0);
		ledd2 := 0
		if storrest != 0 {
			ledd2 = rest1 / storrest
		} else {
			ledd2 = 0
		}
		ret = -((-1000*n)/upem - int(ledd2))
	} else {
		ret = (n/upem)*1000 + ((n%upem)*1000)/upem
	}
	return ret
}
