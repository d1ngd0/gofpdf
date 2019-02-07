package gofpdf

import "github.com/jung-kurt/gofpdf/geh"

type imgInfo struct {
	w, h int
	//src              string
	formatName       string
	colspace         string
	bitsPerComponent string
	filter           string
	decodeParms      string
	trns             []byte
	smask            []byte
	smarkObjID       int
	pal              []byte
	deviceRGBObjID   int
	data             []byte
}

func (s imgInfo) GobEncode() ([]byte, error) {
	return geh.EncodeMany(s.w, s.h, s.formatName, s.colspace, s.bitsPerComponent, s.filter, s.decodeParms, s.trns, s.smask, s.pal, s.data)
}

func (s *imgInfo) GobDecode(buf []byte) error {
	return geh.DecodeMany(buf, &s.w, &s.h, &s.formatName, &s.colspace, &s.bitsPerComponent, &s.filter, &s.decodeParms, &s.trns, &s.smask, &s.pal, &s.data)
}
