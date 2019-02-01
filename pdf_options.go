package gofpdf

import "fmt"

// A pdf option is passed into the New method to set up the initial state of the pdf being generated.
type PdfOption interface {
	apply(pdf *Fpdf) error
}

// unitPdfOption implements PdfOption and will set the unit type for the pdf
type unitPdfOption struct {
	unit int
}

func (u *unitPdfOption) apply(gp *Fpdf) error {
	if u.unit >= unit_max {
		return fmt.Errorf("unit value %d is not valid", u.unit)
	}
	gp.curr.unit = u.unit
	return nil
}

// PdfOptionUnit Creates an option that sets the pdf unit
func PdfOptionUnit(unit int) PdfOption {
	return &unitPdfOption{
		unit: unit,
	}
}

type pageBoundaryPdfOption struct {
	t          int
	x, y, w, h float64
}

func (p *pageBoundaryPdfOption) apply(gp *Fpdf) error {
	pb, err := gp.NewPageBoundary(p.t, p.x, p.y, p.w, p.h)
	if err != nil {
		return err
	}

	gp.curr.pageOption.AddPageBoundary(pb)
	return nil
}

// PdfOptionPageBoundary creates a PdfOption that sets a page boundary. Options that can be set are MediaBox, BleedBox, TrimBox, CropBox and ArtBox
func PdfOptionPageBoundary(t int, x, y, w, h float64) PdfOption {
	return &pageBoundaryPdfOption{
		t: t,
		x: x,
		y: y,
		w: w,
		h: h,
	}
}

// PdfOptionPageSize creates a PdfOption that sets the size of the document. This also sets the size of the MediaBox
func PdfOptionPageSize(w, h float64) PdfOption {
	return PdfOptionPageBoundary(PageBoundaryMedia, 0, 0, w, h)
}

// PdfOptionMediaBox creates a PdfOption that sets the documents MediaBox
func PdfOptionMediaBox(x, y, w, h float64) PdfOption {
	return PdfOptionPageBoundary(PageBoundaryMedia, x, y, w, h)
}

// PdfOptionBleedBox creates a PdfOption that sets the documents BleedBox
func PdfOptionBleedBox(x, y, w, h float64) PdfOption {
	return PdfOptionPageBoundary(PageBoundaryBleed, x, y, w, h)
}

// PdfOptionTrimBox creates a PdfOption that sets the documents TrimBox
func PdfOptionTrimBox(x, y, w, h float64) PdfOption {
	return PdfOptionPageBoundary(PageBoundaryTrim, x, y, w, h)
}

// PdfOptionCropBox creates a PdfOption that sets the documents CropBox
func PdfOptionCropBox(x, y, w, h float64) PdfOption {
	return PdfOptionPageBoundary(PageBoundaryCrop, x, y, w, h)
}

// PdfOptionArtBox creates a PdfOption that sets the documents ArtBox
func PdfOptionArtBox(x, y, w, h float64) PdfOption {
	return PdfOptionPageBoundary(PageBoundaryArt, x, y, w, h)
}

type marginsPdfOption struct {
	l, t, r, b float64
}

func (m *marginsPdfOption) apply(gp *Fpdf) error {
	gp.SetMargins(m.l, m.t, m.r, m.b)
	return nil
}

// PdfOptionMargin creates a new PdfOption that sets the margins of the pdf
func PdfOptionMargin(l, t, r, b float64) PdfOption {
	return &marginsPdfOption{l: l, t: t, r: r, b: b}
}

type compressPdfOption struct {
	level int
}

func (c *compressPdfOption) apply(gp *Fpdf) error {
	gp.SetCompressLevel(c.level)
	return nil
}

// PdfOptionCompress creates a PdfOption that sets the compression level of the document
func PdfOptionCompress(level int) PdfOption {
	return &compressPdfOption{level: level}
}

// PdfOptionNoCompress creates a PdfOption that sets the compression level to none.
func PdfOptionNoCompress() PdfOption {
	return &compressPdfOption{level: 0}
}

type protectionPdfOption struct {
	permissions int
	userpass    []byte
	ownerpass   []byte
}

func (p *protectionPdfOption) apply(gp *Fpdf) error {
	pro := new(PDFProtection)
	err := pro.SetProtection(p.permissions, p.userpass, p.ownerpass)

	if err != nil {
		return err
	}

	gp.pdfProtection = pro
	return nil
}

// PdfOptionProtection creates a PdfOption that set the protection for the document
func PdfOptionProtection(permissions int, userpass string, ownerpass string) PdfOption {
	return &protectionPdfOption{
		permissions: permissions,
		userpass:    []byte(userpass),
		ownerpass:   []byte(ownerpass),
	}
}
