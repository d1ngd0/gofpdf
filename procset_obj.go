package gofpdf

import (
	"fmt"
	"io"
)

const procSetType = "ProcSet"

//ProcSetObj pdf procSet object
type ProcSetObj struct {
	//Font
	Realtes     RelateFonts
	RealteXobjs RealteXobjects
	getRoot     func() *Fpdf
}

func (pr *ProcSetObj) init(funcGetRoot func() *Fpdf) {
	pr.getRoot = funcGetRoot
}

func (pr *ProcSetObj) write(w io.Writer, objID int) error {
	io.WriteString(w, "<<\n")
	io.WriteString(w, "/ProcSet [/PDF /Text /ImageB /ImageC /ImageI]\n")
	io.WriteString(w, "/Font <<\n")
	//me.buffer.WriteString("/F1 9 0 R
	//me.buffer.WriteString("/F2 12 0 R
	//me.buffer.WriteString("/F3 15 0 R
	i := 0
	max := len(pr.Realtes)
	for i < max {
		realte := pr.Realtes[i]
		fmt.Fprintf(w, "      /%s %d 0 R\n", realte.IdOfObj, realte.IndexOfObj+1)
		i++
	}
	io.WriteString(w, ">>\n")
	io.WriteString(w, "/XObject <<\n")
	i = 0
	max = len(pr.RealteXobjs)
	for i < max {
		fmt.Fprintf(w, "/%s %d 0 R\n", pr.RealteXobjs[i].IdOfObj, pr.RealteXobjs[i].IndexOfObj+1)
		pr.getRoot().curr.CountOfL++
		i++
	}
	io.WriteString(w, ">>\n")
	io.WriteString(w, ">>\n")
	return nil
}

func (pr *ProcSetObj) getType() string {
	return "ProcSet"
}

type RelateFonts []RelateFont

func (re RelateFonts) IsContainsFamily(family string) bool {
	i := 0
	max := len(re)
	for i < max {
		if re[i].Family == family {
			return true
		}
		i++
	}
	return false
}

// IsContainsFamilyAndStyle - checks if already exists font with same name and style
func (re RelateFonts) IsContainsFamilyAndStyle(family string, style int) bool {
	rf := re.getFamilyAndStyle(family, style)
	return rf != nil
}

func (re RelateFonts) getFamilyAndStyle(family string, style int) *RelateFont {
	max := len(re)
	for i := 0; i < max; i++ {
		if re[i].Family == family && re[i].Style == style {
			return &re[i]
		}
	}
	return nil
}

func (re RelateFonts) getIndex(id string) int {
	for x := 0; x < len(re); x++ {
		if re[x].IdOfObj == id {
			return re[x].IndexOfObj
		}
	}

	return -1
}

type RelateFont struct {
	Family string
	//etc /F1
	IdOfObj string
	//etc  5 0 R
	IndexOfObj int
	Style      int // Regular|Bold|Italic
}

type RealteXobjects []RealteXobject

type RealteXobject struct {
	IndexOfObj int
	// should start with I when image
	IdOfObj string
}

func (rxo RealteXobjects) getIndex(id string) int {
	for x := 0; x < len(rxo); x++ {
		if rxo[x].IdOfObj == id {
			return rxo[x].IndexOfObj
		}
	}

	return -1
}
