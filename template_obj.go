package gofpdf

import (
	"fmt"
	"io"
)

//TemplateObj Template object
type TemplateObj struct {
	//imagepath string

	pdfProtection *PDFProtection
	getRoot       func() *Fpdf
	images        []ImageHolder
	fonts         []*TemplateFont
	templates     []Template
	x, y          float64
	w, h          float64
	b             []byte
	id            string
}

func newTemplateObj(template Template, p *PDFProtection, funcGetRoot func() *Fpdf) *TemplateObj {
	tpl := new(TemplateObj)
	tpl.fonts = template.Fonts()
	tpl.templates = template.Templates()
	tpl.images = template.Images()
	point, size := template.Size()
	tpl.x, tpl.y = point.X, point.Y
	tpl.w, tpl.h = size.W, size.H
	tpl.b = template.Bytes()
	tpl.id = template.ID()
	tpl.fonts = template.Fonts()
	tpl.getRoot = funcGetRoot
	tpl.pdfProtection = p
	return tpl
}

func (tpl *TemplateObj) setProtection(p *PDFProtection) {
	tpl.pdfProtection = p
}

func (tpl *TemplateObj) protection() *PDFProtection {
	return tpl.pdfProtection
}

func (tpl *TemplateObj) write(w io.Writer, objID int) error {
	io.WriteString(w, "<</Type /XObject\n")
	io.WriteString(w, "/Subtype /Form\n")
	io.WriteString(w, "/Formtype 1\n")
	fmt.Fprintf(w, "/BBox [%.2f %.2f %.2f %.2f]\n", tpl.x, tpl.y, (tpl.x + tpl.w), (tpl.y + tpl.h))
	if tpl.x != 0 || tpl.y != 0 {
		fmt.Fprintf(w, "/Matrix [1 0 0 1 %.5f %.5f]\n", -tpl.x*2, tpl.y*2)
	}

	// Template's resource dictionary
	io.WriteString(w, "/Resources \n")
	io.WriteString(w, "<</ProcSet [/PDF /Text /ImageB /ImageC /ImageI]\n")

	io.WriteString(w, "/Font <<\n")
	for x := 0; x < len(tpl.fonts); x++ {
		id := fmt.Sprintf("F%s", tpl.fonts[x].id)
		fmt.Fprintf(w, "/%s %d 0 R\n", id, tpl.getProcsetIndex(id, true)+1)
	}
	io.WriteString(w, ">>\n")

	if len(tpl.images) > 0 { //|| len(tTemplates) > 0 {
		io.WriteString(w, "/XObject <<\n")
		for x := 0; x < len(tpl.images); x++ {
			id := fmt.Sprintf("I%s", tpl.images[x].ID())
			fmt.Fprintf(w, "/%s %d 0 R\n", id, tpl.getProcsetIndex(id, false)+1)
		}
		for x := 0; x < len(tpl.templates); x++ {
			id := fmt.Sprintf("TPL%s", tpl.templates[x].ID())
			fmt.Fprintf(w, "/TPL%s %d 0 R\n", id, tpl.getProcsetIndex(id, false)+1)
		}
		io.WriteString(w, ">>\n")
	}

	io.WriteString(w, ">>\n")

	fmt.Fprintf(w, "/Length %d\n>>\n", len(tpl.b))
	io.WriteString(w, "stream\n")
	if tpl.protection() != nil {
		tmp, err := rc4Cip(tpl.protection().objectkey(objID), tpl.b)
		if err != nil {
			return err
		}
		w.Write(tmp)
		io.WriteString(w, "\n")
	} else {
		w.Write(tpl.b)
	}
	io.WriteString(w, "\nendstream\n")

	return nil
}

func (tpl *TemplateObj) getType() string {
	return "Template"
}

func (tpl *TemplateObj) ToTemplate() Template {
	return &FpdfTpl{
		corner:    Point{X: tpl.x, Y: tpl.y},
		size:      &Rect{W: tpl.w, H: tpl.h},
		bytes:     [][]byte{tpl.b},
		fonts:     tpl.fonts,
		images:    tpl.images,
		templates: tpl.templates,
		page:      0,
	}
}

func (tpl *TemplateObj) procsetIdentifier() string {
	return fmt.Sprintf("TPL%s", tpl.id)
}

func (tpl *TemplateObj) getProcsetIndex(id string, isFont bool) int {
	return tpl.getRoot().getProcsetIndex(id, isFont)
}
