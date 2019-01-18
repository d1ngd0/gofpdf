package gofpdf

type Margins struct {
	Left, Top, Right, Bottom float64
}

// SetMargins defines the left, top, right and bottom margins. By default, they equal 1 cm. Call this method to change them.
func (gp *Fpdf) SetMargins(left, top, right, bottom float64) {
	gp.UnitsToPointsVar(&left, &top, &right, &bottom)
	gp.margins = Margins{left, top, right, bottom}
}

// SetMarginLeft sets the left margin
func (gp *Fpdf) SetMarginLeft(margin float64) {
	gp.margins.Left = gp.UnitsToPoints(margin)
}

// SetMarginTop sets the top margin
func (gp *Fpdf) SetMarginTop(margin float64) {
	gp.margins.Top = gp.UnitsToPoints(margin)
}

// SetMarginRight sets the right margin
func (gp *Fpdf) SetMarginRight(margin float64) {
	gp.margins.Right = gp.UnitsToPoints(margin)
}

// SetMarginBottom set the bottom margin
func (gp *Fpdf) SetMarginBottom(margin float64) {
	gp.margins.Bottom = gp.UnitsToPoints(margin)
}

// Margins gets the current margins, The margins will be converted back to the documents units. Returned values will be in the following order Left, Top, Right, Bottom
func (gp *Fpdf) Margins() (float64, float64, float64, float64) {
	return gp.PointsToUnits(gp.margins.Left),
		gp.PointsToUnits(gp.margins.Top),
		gp.PointsToUnits(gp.margins.Right),
		gp.PointsToUnits(gp.margins.Bottom)
}

// MarginLeft returns the left margin
func (gp *Fpdf) MarginLeft() float64 {
	return gp.PointsToUnits(gp.margins.Left)
}

// MarginTop returns the top margin
func (gp *Fpdf) MarginTop() float64 {
	return gp.PointsToUnits(gp.margins.Top)
}

// MarginRight returns the right margin
func (gp *Fpdf) MarginRight() float64 {
	return gp.PointsToUnits(gp.margins.Right)
}

// MarginBottom returns the bottom margin
func (gp *Fpdf) MarginBottom() float64 {
	return gp.PointsToUnits(gp.margins.Bottom)
}
