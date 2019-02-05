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
	"io"
	"io/ioutil"
)

// Template is an object that can be written to, then used and re-used any number of times within a document.
type Template interface {
	ID() string
	Size() (Point, Rect)
	Bytes() []byte
	Images() []ImageHolder
	Fonts() []*TemplateFont
	Templates() []Template
	NumPages() int
	FromPage(int) (Template, error)
	FromPages() []Template
	Serialize() ([]byte, error)
	gob.GobDecoder
	gob.GobEncoder
}

func NewTemplateFont(id, family string, option TtfOption, b []byte) *TemplateFont {
	return &TemplateFont{
		id:     id,
		family: family,
		option: option,
		b:      b,
	}
}

type TemplateFont struct {
	id         string
	procsetId  string
	family     string
	option     TtfOption
	b          []byte
	bbuffer    io.Reader
	characters string
}

func (t *TemplateFont) Read(p []byte) (n int, err error) {
	if t.bbuffer == nil {
		t.bbuffer = bytes.NewBuffer(t.b)
	}

	return t.bbuffer.Read(p)
}

// GobEncode encodes the receiving template into a byte buffer. Use GobDecode
// to decode the byte buffer back to a template.
func (t *TemplateFont) GobEncode() ([]byte, error) {
	var err error
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	encodeMe := []interface{}{t.id, t.procsetId, t.family, t.option, t.b, t.characters}

	for x := 0; x < len(encodeMe); x++ {
		if err == nil {
			err = encoder.Encode(encodeMe[x])
		}
	}

	return w.Bytes(), err
}

// GobDecode decodes the specified byte buffer into the receiving template.
func (t *TemplateFont) GobDecode(buf []byte) error {
	var err error
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	decodeMe := []interface{}{&t.id, &t.procsetId, &t.family, &t.option, &t.b, &t.characters}

	for x := 0; x < len(decodeMe); x++ {
		if err == nil {
			err = decoder.Decode(decodeMe[x])
		}
	}

	return err
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

	contents := gp.getAllContent()
	bytes := make([][]byte, len(contents))
	sizes := make([]Rect, len(contents))
	x := 0

	for index, content := range contents {
		bytes[x], err = content.bytes(index)
		if err != nil {
			return nil, err
		}

		page := gp.pdfObjs[content.pageIndex].(*PageObj)
		if pb := page.pageOption.GetBoundary(PageBoundaryMedia); pb != nil {
			gp.curr.IndexOfPageObj = content.pageIndex
			sizes[x] = gp.GetBoundarySize(PageBoundaryMedia)
		}
		x++
	}

	fpdf := &FpdfTpl{
		corner: corner,
		size:   sizes,
		bytes:  bytes,
		page:   len(bytes) - 1,
	}

	if fpdf.fonts, err = gp.getTemplateFonts(); err != nil {
		return nil, err
	}

	if fpdf.images, err = gp.getImageHolders(); err != nil {
		return nil, err
	}

	fpdf.templates, err = gp.getTemplates()
	return fpdf, err
}

// FpdfTpl is a concrete implementation of the Template interface.
type FpdfTpl struct {
	corner    Point
	size      []Rect
	bytes     [][]byte
	fonts     []*TemplateFont
	images    []ImageHolder
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
func (t *FpdfTpl) Images() []ImageHolder {
	return t.images
}

// Fonts returns a list of the fonts used in the template
func (t *FpdfTpl) Fonts() []*TemplateFont {
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
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(t)

	return b.Bytes(), err
}

// DeserializeTemplate creaties a template from a previously serialized
// template
func DeserializeTemplate(b []byte) (Template, error) {
	tpl := new(FpdfTpl)
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	err := dec.Decode(tpl)
	return tpl, err
}

// GobEncode encodes the receiving template into a byte buffer. Use GobDecode
// to decode the byte buffer back to a template.
func (t *FpdfTpl) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)

	err := encoder.Encode(t.templates)

	if err == nil {
		err = encoder.Encode(len(t.images))
	}

	for x := 0; x < len(t.images); x++ {
		if err == nil {
			var b []byte
			b, err = ioutil.ReadAll(t.images[x])

			if err == nil {
				err = encoder.Encode(b)
			}
		}
	}

	if err == nil {
		err = encoder.Encode(t.fonts)
	}
	if err == nil {
		err = encoder.Encode(t.corner)
	}
	if err == nil {
		err = encoder.Encode(t.size)
	}
	if err == nil {
		err = encoder.Encode(t.bytes)
	}
	if err == nil {
		err = encoder.Encode(t.page)
	}

	return w.Bytes(), err
}

// GobDecode decodes the specified byte buffer into the receiving template.
func (t *FpdfTpl) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)

	tpls := make([]*FpdfTpl, 0)
	err := decoder.Decode(&tpls)
	t.templates = make([]Template, len(tpls))
	for x := 0; x < len(t.templates); x++ {
		t.templates[x] = Template(tpls[x])
	}

	var numImages int
	if err == nil {
		err = decoder.Decode(&numImages)
	}
	t.images = make([]ImageHolder, numImages)

	if err == nil {
		for x := 0; x < numImages; x++ {
			var b []byte
			err = decoder.Decode(&b)

			if err == nil {
				var ih ImageHolder
				ih, err = ImageHolderByBytes(b)
				t.images[x] = ih
			}
		}
	}

	if err == nil {
		err = decoder.Decode(&t.fonts)
	}
	if err == nil {
		err = decoder.Decode(&t.corner)
	}
	if err == nil {
		err = decoder.Decode(&t.size)
	}
	if err == nil {
		err = decoder.Decode(&t.bytes)
	}
	if err == nil {
		err = decoder.Decode(&t.page)
	}

	return err
}

func (gp *Fpdf) loadParamsFromFpdf(f *Fpdf) {
	gp.SetNoCompression()

	gp.curr.X = f.curr.X
	gp.curr.Y = f.curr.Y
	gp.SetLineWidth(f.curr.lineWidth)
	gp.SetLineCapStyle(f.curr.capStyle)
	gp.SetLineJoinStyle(f.curr.joinStyle)

	gp.loadFontsFromFpdf(f)
	gp.curr.Font_Size = f.curr.Font_Size
	gp.curr.Font_Style = f.curr.Font_Style
	gp.curr.Font_Type = f.curr.Font_Type
}

func (gp *Fpdf) loadFontsFromFpdf(f *Fpdf) {
	for x := 0; x < len(f.pdfObjs); x++ {
		obj := f.pdfObjs[x]
		if subsetFont, ok := obj.(*SubsetFontObj); ok {
			ssf := gp.loadFontFromFpdf(f, subsetFont)
			// set the current font
			if ssf.procsetIdentifier() == f.curr.Font_ISubset.procsetIdentifier() {
				gp.curr.Font_ISubset = ssf
			}
		}
	}
}

func (gp *Fpdf) loadFontFromFpdf(f *Fpdf, orig *SubsetFontObj) *SubsetFontObj {
	subfont := orig.copy()
	cidfont := new(CIDFontObj)
	unicodemap := new(UnicodeMap)
	subfontdesc := new(SubfontDescriptorObj)
	pdfdic := new(PdfDictionaryObj)

	*cidfont = *f.pdfObjs[orig.indexObjCIDFont].(*CIDFontObj)
	*unicodemap = *f.pdfObjs[orig.indexObjUnicodeMap].(*UnicodeMap)
	*subfontdesc = *f.pdfObjs[cidfont.indexObjSubfontDescriptor].(*SubfontDescriptorObj)
	*pdfdic = *f.pdfObjs[subfontdesc.indexObjPdfDictionary].(*PdfDictionaryObj)

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
	index := gp.addObj(subfont) //add หลังสุด
	id := subfont.procsetIdentifier()

	procset := gp.getProcset()
	if !procset.Realtes.IsContainsFamilyAndStyle(subfont.Family, subfont.ttfFontOption.Style&^Underline) {
		procset.Realtes = append(procset.Realtes, RelateFont{Family: subfont.Family, IndexOfObj: index, IdOfObj: id, Style: subfont.ttfFontOption.Style &^ Underline})
	}

	return subfont
}
