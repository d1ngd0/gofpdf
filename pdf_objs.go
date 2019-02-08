package gofpdf

import (
	"fmt"
	"io"
)

type pdfObjs struct {
	objs       []IObj
	typeMap    map[string][]int
	procsetMap map[string]int
	getRoot    func() *Fpdf
}

func (p *pdfObjs) write(writer *countingWriter) ([]int, error) {
	max := len(p.objs)
	linelens := make([]int, max)

	for x := 0; x < max; x++ {
		objID := x + 1
		linelens[x] = writer.offset
		pdfObj := p.objs[x]
		fmt.Fprintf(writer, "%d 0 obj\n", objID)
		if err := pdfObj.write(writer, objID); err != nil {
			return linelens, err
		}
		io.WriteString(writer, "endobj\n\n")
	}

	return linelens, nil
}

func newPdfObjs(funcGetRoot func() *Fpdf) *pdfObjs {
	p := &pdfObjs{
		objs:       make([]IObj, 0),
		typeMap:    make(map[string][]int),
		procsetMap: make(map[string]int),
		getRoot:    funcGetRoot,
	}
	return p
}

func (p *pdfObjs) addObj(iobj IObj) int {
	index := len(p.objs)
	p.objs = append(p.objs, iobj)
	p.typeMap[iobj.getType()] = append(p.typeMap[iobj.getType()], index)
	return index
}

func (p *pdfObjs) addProcsetObj(iobj ProcsetIObj) (string, int) {
	var index int
	pid := iobj.procsetIdentifier()

	if i, ok := p.procsetMap[pid]; !ok {
		procset := p.procset()
		index = p.addObj(iobj)
		p.procsetMap[pid] = index
		if subfont, ok := iobj.(*SubsetFontObj); ok {
			procset.Realtes = append(procset.Realtes,
				RelateFont{
					Family:     subfont.Family,
					IndexOfObj: index,
					IdOfObj:    pid,
					Style:      subfont.ttfFontOption.Style &^ Underline,
				})
		} else {
			procset.RealteXobjs = append(procset.RealteXobjs,
				RealteXobject{
					IndexOfObj: index,
					IdOfObj:    pid,
				})
		}
	} else {
		index = i
	}

	return pid, index
}

func (p *pdfObjs) hasProcsetID(pid string) (int, bool) {
	id, ok := p.procsetMap[pid]
	return id, ok
}

func (p *pdfObjs) hasProcsetObj(iobj ProcsetIObj) (string, int, bool) {
	pid := iobj.procsetIdentifier()
	id, e := p.hasProcsetID(pid)
	return pid, id, e
}

func (p *pdfObjs) indexOfFirst(t string) int {
	if len(p.typeMap[t]) <= 0 {
		return -1
	}

	return p.typeMap[t][0]
}

func (p *pdfObjs) indexOfLatest(t string) int {
	if len(p.typeMap[t]) <= 0 {
		return -1
	}

	return p.typeMap[t][(len(p.typeMap[t]) - 1)]
}

func (p *pdfObjs) at(index int) IObj {
	if len(p.objs) <= index || index < 0 {
		return nil
	}

	return p.objs[index]
}

func (p *pdfObjs) allOf(t string) []IObj {
	obs, ok := p.typeMap[t]
	if !ok {
		return make([]IObj, 0)
	}

	l := len(obs)
	iobjs := make([]IObj, l)
	for x := 0; x < l; x++ {
		iobjs[x] = p.objs[obs[x]]
	}

	return iobjs
}

func (p *pdfObjs) allPages() []*PageObj {
	obs, ok := p.typeMap[pageType]
	if !ok {
		return make([]*PageObj, 0)
	}

	l := len(obs)
	iobjs := make([]*PageObj, l)
	for x := 0; x < l; x++ {
		iobjs[x] = p.objs[obs[x]].(*PageObj)
	}

	return iobjs
}

func (p *pdfObjs) allImages() []*ImageObj {
	obs, ok := p.typeMap[imageType]
	if !ok {
		return make([]*ImageObj, 0)
	}

	l := len(obs)
	iobjs := make([]*ImageObj, l)
	for x := 0; x < l; x++ {
		iobjs[x] = p.objs[obs[x]].(*ImageObj)
	}

	return iobjs
}

func (p *pdfObjs) allSubsetFonts() []*SubsetFontObj {
	obs, ok := p.typeMap[subsetFontType]
	if !ok {
		return make([]*SubsetFontObj, 0)
	}

	l := len(obs)
	iobjs := make([]*SubsetFontObj, l)
	for x := 0; x < l; x++ {
		iobjs[x] = p.objs[obs[x]].(*SubsetFontObj)
	}

	return iobjs
}

func (p *pdfObjs) procset() *ProcSetObj {
	index := p.indexOfFirst(procSetType)
	var procset *ProcSetObj
	if index < 0 {
		procset = new(ProcSetObj)
		procset.init(p.getRoot)
		p.addObj(procset)
	} else {
		procset = p.objs[index].(*ProcSetObj)
	}
	return procset
}

func (p *pdfObjs) currentPage() *PageObj {
	index := p.indexOfLatest(pageType)
	return p.getPage(index)
}

func (p *pdfObjs) getPages() *PagesObj {
	index := p.indexOfFirst(pagesType)
	if pages, ok := p.objs[index].(*PagesObj); ok {
		return pages
	}

	return nil
}

func (p *pdfObjs) getPage(index int) *PageObj {
	if len(p.objs) <= index || index < 0 {
		return nil
	}

	if page, ok := p.objs[index].(*PageObj); ok {
		return page
	}

	return nil
}

func (p *pdfObjs) currentContent() *ContentObj {
	page := p.currentPage()
	return p.getPageContent(page)
}

func (p *pdfObjs) getPageContent(page *PageObj) *ContentObj {
	var content *ContentObj
	if page == nil {
		return content
	}

	if page.indexOfContentObj < 0 {
		content = new(ContentObj)
		content.init(p.getRoot)
		page.setIndexOfContentObj(p.addObj(content))
	} else {
		content = p.getContent(page.indexOfContentObj)
	}

	return content

}

func (p *pdfObjs) getContent(index int) *ContentObj {
	if len(p.objs) <= index || index < 0 {
		return nil
	}

	if content, ok := p.objs[index].(*ContentObj); ok {
		return content
	}

	return nil
}

func (p *pdfObjs) getSubsetFontObjByFamilyAndStyle(family string, style int) *SubsetFontObj {
	rf := p.procset().Realtes.getFamilyAndStyle(family, style)
	if rf == nil {
		return nil
	}

	return p.subsetFontObjAt(rf.IndexOfObj)
}

func (p *pdfObjs) subsetFontObjAt(index int) *SubsetFontObj {
	if len(p.objs) <= index || index < 0 {
		return nil
	}

	if sf, ok := p.objs[index].(*SubsetFontObj); ok {
		return sf
	}

	return nil
}
