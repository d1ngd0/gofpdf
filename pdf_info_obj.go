package gofpdf

import (
	"fmt"
	"io"
	"time"
)

//PdfInfo Document Information Dictionary
type PdfInfo struct {
	Title    string //The document’s title
	Author   string //The name of the person who created the document.
	Subject  string //The subject of the document.
	Creator  string // If the document was converted to PDF from another format, the name of the application original document from which it was converted.
	Producer string //If the document was converted to PDF from another format, the name of the application (for example, Acrobat Distiller) that converted it to PDF.
	Keywords string // The keywords associated with the document
}

// writeInfo writes the pdf info object to a writer in the pdf spec
func (info *PdfInfo) write(w io.Writer) {
	io.WriteString(w, "/Info <<\n")

	if info.Author != "" {
		fmt.Fprintf(w, "/Author <FEFF%s>\n", encodeUtf8(info.Author))
	}

	if info.Keywords != "" {
		fmt.Fprintf(w, "/Keywords <FEFF%s>\n", encodeUtf8(info.Keywords))
	}

	if info.Title != "" {
		fmt.Fprintf(w, "/Title <FEFF%s>\n", encodeUtf8(info.Title))
	}

	if info.Subject != "" {
		fmt.Fprintf(w, "/Subject <FEFF%s>\n", encodeUtf8(info.Subject))
	}

	if info.Creator != "" {
		fmt.Fprintf(w, "/Creator <FEFF%s>\n", encodeUtf8(info.Creator))
	}

	if info.Producer != "" {
		fmt.Fprintf(w, "/Producer <FEFF%s>\n", encodeUtf8(info.Producer))
	}

	fmt.Fprintf(w, "/CreationDate(D:%s)>>\n", infodate(time.Now()))
	io.WriteString(w, " >>\n")
}

//SetInfo set Document Information Dictionary
func (gp *Fpdf) SetInfo(info PdfInfo) {
	gp.info = &info
}

//GetInfo get Document Information Dictionary
func (gp *Fpdf) GetInfo() *PdfInfo {
	if gp.info == nil {
		gp.info = &PdfInfo{}
	}
	return gp.info
}

// SetTitle defines the title of the document.
func (gp *Fpdf) SetTitle(titleStr string) {
	gp.GetInfo().Title = titleStr
}

// SetSubject defines the subject of the document.
func (gp *Fpdf) SetSubject(subjectStr string) {
	gp.GetInfo().Subject = subjectStr
}

// SetAuthor defines the author of the document.
func (gp *Fpdf) SetAuthor(authorStr string) {
	gp.GetInfo().Author = authorStr
}

// SetKeywords defines the keywords of the document. keywordStr is a
// space-delimited string, for example "invoice August".
func (gp *Fpdf) SetKeywords(keywordsStr string) {
	gp.GetInfo().Keywords = keywordsStr
}

// SetCreator defines the creator of the document.
func (gp *Fpdf) SetCreator(creatorStr string) {
	gp.GetInfo().Creator = creatorStr
}

// SetProdcer defines the producer of the document.
func (gp *Fpdf) SetProducer(producerStr string) {
	gp.GetInfo().Producer = producerStr
}
