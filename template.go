package gofpdf

/*
 * Copyright (c) 2015 Kurt Jung (Gmail: kurt.w.jung),
 *   Marcus Downing, Jan Slabon (Setasign)
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

import (
	"bytes"
	"crypto/sha1"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/ISeeMe/gofpdf/bp"
	"github.com/ISeeMe/gofpdf/geh"
)

// Template is an object that can be written to, then used and re-used any number of times within a document.
type Template interface {
	ID() string
	Size() (Point, Rect)
	Bytes() []byte
	Images() []*ImageObj
	Fonts() []*SubsetFontObj
	Templates() []Template
	NumPages() int
	FromPage(int) (Template, error)
	FromPages() []Template
	Serialize() ([]byte, error)
	gob.GobDecoder
	gob.GobEncoder
}

func NewTemplateFont(sf *SubsetFontObj) *TemplateFont {
	return &TemplateFont{
		subsetFont: sf,
	}
}

type TemplateFont struct {
	subsetFont *SubsetFontObj
}

// GobEncode encodes the receiving template into a byte buffer. Use GobDecode
// to decode the byte buffer back to a template.
func (t *TemplateFont) GobEncode() ([]byte, error) {
	return geh.EncodeMany(t.subsetFont)
}

// GobDecode decodes the specified byte buffer into the receiving template.
func (t *TemplateFont) GobDecode(buf []byte) error {
	return geh.DecodeMany(buf, &t.subsetFont)
}

func NewTemplateImage(img *ImageObj) *TemplateImage {
	return &TemplateImage{
		image: img,
	}
}

type TemplateImage struct {
	image *ImageObj
}

// GobEncode encodes the receiving template into a byte buffer. Use GobDecode
// to decode the byte buffer back to a template.
func (t *TemplateImage) GobEncode() ([]byte, error) {
	return geh.EncodeMany(t.image)
}

// GobDecode decodes the specified byte buffer into the receiving template.
func (t *TemplateImage) GobDecode(buf []byte) error {
	return geh.DecodeMany(buf, &t.image)
}

type TplFunc func(*Fpdf) error

// newTpl creates a template, copying graphics settings from a template if one is given
func newTpl(corner Point, opts []PdfOption, fn TplFunc, copyFrom *Fpdf) (Template, error) {
	gp, err := New(opts...)
	if err != nil {
		return nil, err
	}
	gp.AddPage()

	if copyFrom != nil {
		gp.loadParamsFromFpdf(copyFrom)
	}

	if err = fn(gp); err != nil {
		return nil, err
	}

	return gp.Template(corner)
}

func (gp *Fpdf) Template(corner Point) (*FpdfTpl, error) {
	var err error
	pages := gp.getAllPages()
	numPages := len(pages)
	bytes := make([][]byte, numPages)
	sizes := make([]Rect, numPages)

	for x := 0; x < numPages; x++ {
		page := pages[x]
		content := page.getContent()
		bytes[x], err = content.bytes(page.indexOfContentObj)
		if err != nil {
			return nil, err
		}

		if pb := page.pageOption.GetBoundary(PageBoundaryMedia); pb != nil {
			sizes[x] = gp.GetBoundarySize(PageBoundaryMedia)
		}
	}

	fpdf := &FpdfTpl{
		corner: corner,
		size:   sizes,
		bytes:  bytes,
		page:   len(bytes) - 1,
	}

	fpdf.fonts = gp.getAllSubsetFonts()
	fpdf.images = gp.getAllImages()
	fpdf.templates, err = gp.getTemplates()
	return fpdf, err
}

// FpdfTpl is a concrete implementation of the Template interface.
type FpdfTpl struct {
	corner    Point
	size      []Rect
	bytes     [][]byte
	fonts     []*SubsetFontObj
	images    []*ImageObj
	templates []Template
	page      int
}

// ID returns the global template identifier
func (t *FpdfTpl) ID() string {
	return fmt.Sprintf("%x", sha1.Sum(t.Bytes()))
}

// Size gives the bounding dimensions of this template
func (t *FpdfTpl) Size() (corner Point, size Rect) {
	return t.corner, t.size[t.page]
}

// Bytes returns the actual template data, not including resources
func (t *FpdfTpl) Bytes() []byte {
	return t.bytes[t.page]
}

// FromPage creates a new template from a specific Page
func (t *FpdfTpl) FromPage(page int) (Template, error) {
	// pages start at 1
	page--

	if page < 0 {
		return nil, errors.New("Pages start at 1 No template will have a page 0")
	}

	if page > t.NumPages() {
		return nil, fmt.Errorf("The template does not have a page %d", page)
	}
	// if it is already pointing to the correct page
	// there is no need to create a new template
	if t.page == page {
		return t, nil
	}

	t2 := *t
	t2.page = page
	return &t2, nil
}

// FromPages creates a template slice with all the pages within a template.
func (t *FpdfTpl) FromPages() []Template {
	p := make([]Template, t.NumPages())
	for x := 0; x < t.NumPages(); x++ {
		// the only error is when accessing a
		// non existing template... that can't happen
		// here
		p[x], _ = t.FromPage(x)
	}

	return p
}

// Images returns a list of the images used in this template
func (t *FpdfTpl) Images() []*ImageObj {
	return t.images
}

// Fonts returns a list of the fonts used in the template
func (t *FpdfTpl) Fonts() []*SubsetFontObj {
	return t.fonts
}

// Templates returns a list of templates used in this template
func (t *FpdfTpl) Templates() []Template {
	return t.templates
}

// NumPages returns the number of available pages within the template. Look at FromPage and FromPages on access to that content.
func (t *FpdfTpl) NumPages() int {
	// the first page is empty to
	// make the pages begin at one
	return len(t.bytes)
}

// Serialize turns a template into a byte string for later deserialization
func (t *FpdfTpl) Serialize() ([]byte, error) {
	b := bp.GetBuffer()
	enc := gob.NewEncoder(b)
	err := enc.Encode(t)
	return b.Bytes(), err
}

// DeserializeTemplate creaties a template from a previously serialized
// template
func DeserializeTemplate(b []byte) (Template, error) {
	tpl := new(FpdfTpl)
	dec := gob.NewDecoder(bytes.NewReader(b))
	err := dec.Decode(tpl)
	if err != nil {
	}
	return tpl, err
}

// GobEncode encodes the receiving template into a byte buffer. Use GobDecode
// to decode the byte buffer back to a template.
func (t *FpdfTpl) GobEncode() ([]byte, error) {
	return geh.EncodeMany(t.templates, t.images, t.fonts, t.corner, t.size, t.bytes, t.page)
}

// GobDecode decodes the specified byte buffer into the receiving template.
func (t *FpdfTpl) GobDecode(buf []byte) error {
	tpls := make([]*FpdfTpl, 0)

	if err := geh.DecodeMany(buf, &tpls, &t.images, &t.fonts, &t.corner, &t.size, &t.bytes, &t.page); err != nil {
		return err
	}

	t.templates = make([]Template, len(tpls))
	for x := 0; x < len(t.templates); x++ {
		t.templates[x] = Template(tpls[x])
	}

	return nil
}

func (gp *Fpdf) loadParamsFromFpdf(f *Fpdf) {
	gp.SetNoCompression()

	gp.curr.X = f.curr.X
	gp.curr.Y = f.curr.Y
	gp.curr.unit = f.curr.unit
	gp.SetLineWidth(f.curr.lineWidth)
	gp.SetLineCapStyle(f.curr.capStyle)
	gp.SetLineJoinStyle(f.curr.joinStyle)

	gp.loadFontsFromFpdf(f)
	gp.curr.Font_Size = f.curr.Font_Size
	gp.curr.Font_Style = f.curr.Font_Style
	gp.curr.Font_Type = f.curr.Font_Type
}

func (gp *Fpdf) loadFontsFromFpdf(f *Fpdf) {
	fonts := gp.pdfObjs.allSubsetFonts()
	max := len(fonts)
	for x := 0; x < max; x++ {
		ssf := gp.loadFontFromFpdf(f, fonts[x])
		// set the current font
		if ssf.procsetIdentifier() == f.curr.Font_ISubset.procsetIdentifier() {
			gp.curr.Font_ISubset = ssf
		}
	}
}

func (gp *Fpdf) loadFontFromFpdf(f *Fpdf, orig *SubsetFontObj) *SubsetFontObj {
	subfont := orig.copy()
	cidfont := new(CIDFontObj)
	unicodemap := new(UnicodeMap)
	subfontdesc := new(SubfontDescriptorObj)
	pdfdic := new(PdfDictionaryObj)

	*cidfont = *f.pdfObjs.at(orig.indexObjCIDFont).(*CIDFontObj)
	*unicodemap = *f.pdfObjs.at(orig.indexObjUnicodeMap).(*UnicodeMap)
	*subfontdesc = *f.pdfObjs.at(cidfont.indexObjSubfontDescriptor).(*SubfontDescriptorObj)
	*pdfdic = *f.pdfObjs.at(subfontdesc.indexObjPdfDictionary).(*PdfDictionaryObj)

	unicodemap.SetPtrToSubsetFontObj(subfont)
	unicodeindex := gp.addObj(unicodemap)

	pdfdic.SetPtrToSubsetFontObj(subfont)
	pdfdicindex := gp.addObj(pdfdic)

	subfontdesc.SetPtrToSubsetFontObj(subfont)
	subfontdesc.SetIndexObjPdfDictionary(pdfdicindex)
	subfontdescindex := gp.addObj(subfontdesc)

	cidfont.SetPtrToSubsetFontObj(subfont)
	cidfont.SetIndexObjSubfontDescriptor(subfontdescindex)
	cidindex := gp.addObj(cidfont)

	subfont.SetIndexObjCIDFont(cidindex)
	subfont.SetIndexObjUnicodeMap(unicodeindex)

	gp.addProcsetObj(subfont)

	return subfont
}
