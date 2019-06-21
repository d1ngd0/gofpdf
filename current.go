package gofpdf

//Current current state
type Current struct {
	setXCount int //many times we go func SetX()
	X         float64
	Y         float64

	//font
	IndexOfFontObj int
	CountOfFont    int
	CountOfL       int

	Font_Size      float64
	Font_Style     int // Regular|Bold|Italic|Underline
	Font_FontCount int
	Font_Type      int // CURRENT_FONT_TYPE_IFONT or  CURRENT_FONT_TYPE_SUBSET

	Font_ISubset *SubsetFontObj // Font_Type == CURRENT_FONT_TYPE_SUBSET
	Text_Option  TextOption
	//page
	IndexOfPageObj int

	//img
	CountOfImg int

	//text color
	txtColor Rgb

	//text grayscale
	grayFill float64
	//draw grayscale
	grayStroke float64

	lineWidth float64
	capStyle  int
	joinStyle int

	lheight    float64
	unit       int
	pageOption PageOption
}

func (c *Current) setTextColor(rgb Rgb) {
	c.txtColor = rgb
}

func (c *Current) textColor() Rgb {
	return c.txtColor
}

func (c *Current) setLineHeight(h float64) float64 {
	if h <= 0 {
		c.lheight = h
	}

	return c.lheight
}

//Rgb  rgb color
type Rgb struct {
	r uint8
	g uint8
	b uint8
}

//SetR set red
func (rgb *Rgb) SetR(r uint8) {
	rgb.r = r
}

//SetG set green
func (rgb *Rgb) SetG(g uint8) {
	rgb.g = g
}

//SetB set blue
func (rgb *Rgb) SetB(b uint8) {
	rgb.b = b
}

func (rgb Rgb) equal(obj Rgb) bool {
	if rgb.r == obj.r && rgb.g == obj.g && rgb.b == obj.b {
		return true
	}
	return false
}
