package gofpdf

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/ISeeMe/gofpdf/geh"
)

const imageType = "Image"

//ImageObj image object
type ImageObj struct {
	//imagepath string

	rawImgReader  *bytes.Reader
	imginfo       imgInfo
	pdfProtection *PDFProtection
	procsetid     string
	imageid       string
	getRoot       func() *Fpdf
}

func (i *ImageObj) Serialize() ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err := enc.Encode(i)
	return b.Bytes(), err
}

func DeserializeImage(b []byte) (*ImageObj, error) {
	var i ImageObj
	dec := gob.NewDecoder(bytes.NewReader(b))
	err := dec.Decode(&i)
	return &i, err
}

func (s *ImageObj) GobEncode() ([]byte, error) {
	s.rawImgReader.Seek(0, 0)
	b, err := ioutil.ReadAll(s.rawImgReader)
	if err != nil {
		return make([]byte, 0), err
	}

	return geh.EncodeMany(b, s.imginfo, s.procsetid, s.imageid)
}

func (s *ImageObj) GobDecode(buf []byte) error {
	var b []byte
	if err := geh.DecodeMany(buf, &b, &s.imginfo, &s.procsetid, &s.imageid); err != nil {
		return err
	}

	s.rawImgReader = bytes.NewReader(b)
	return nil
}

func NewImageObj(img ImageHolder) (*ImageObj, error) {
	imgobj := new(ImageObj)
	imgobj.imageid = img.ID()
	imgobj.procsetid = fmt.Sprintf("I%s", imgobj.imageid)
	if err := imgobj.SetImage(img); err != nil {
		return imgobj, err
	}

	err := imgobj.parse()
	return imgobj, err
}

func (i *ImageObj) setProtection(p *PDFProtection) {
	i.pdfProtection = p
}

func (i *ImageObj) protection() *PDFProtection {
	return i.pdfProtection
}

func (i *ImageObj) write(w io.Writer, objID int) error {

	err := writeImgProp(w, i.imginfo)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "/Length %d\n>>\n", len(i.imginfo.data)) // /Length 62303>>\n
	io.WriteString(w, "stream\n")
	if i.protection() != nil {
		tmp, err := rc4Cip(i.protection().objectkey(objID), i.imginfo.data)
		if err != nil {
			return err
		}
		w.Write(tmp)
		io.WriteString(w, "\n")
	} else {
		w.Write(i.imginfo.data)
	}
	io.WriteString(w, "\nendstream\n")

	return nil
}

func (i *ImageObj) isColspaceIndexed() bool {
	return isColspaceIndexed(i.imginfo)
}

func (i *ImageObj) haveSMask() bool {
	return haveSMask(i.imginfo)
}

func (i *ImageObj) createSMask() (*SMask, error) {
	var smk SMask
	smk.setProtection(i.protection())
	smk.w = i.imginfo.w
	smk.h = i.imginfo.h
	smk.colspace = "DeviceGray"
	smk.bitsPerComponent = "8"
	smk.filter = i.imginfo.filter
	smk.data = i.imginfo.smask
	smk.decodeParms = fmt.Sprintf("/Predictor 15 /Colors 1 /BitsPerComponent 8 /Columns %d", i.imginfo.w)
	return &smk, nil
}

func (i *ImageObj) createDeviceRGB() (*DeviceRGBObj, error) {
	var dRGB DeviceRGBObj
	dRGB.data = i.imginfo.pal
	return &dRGB, nil
}

func (i *ImageObj) getType() string {
	return imageType
}

//SetImagePath set image path
func (i *ImageObj) SetImagePath(path string) error {

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	err = i.SetImage(file)
	if err != nil {
		return err
	}
	return nil
}

//SetImage set image
func (i *ImageObj) SetImage(r io.Reader) error {

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	i.rawImgReader = bytes.NewReader(data)

	return nil
}

//GetRect get rect of img
func (i *ImageObj) GetRect() *Rect {

	rect, err := i.getRect()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	return rect
}

//GetRect get rect of img
func (i *ImageObj) getRect() (*Rect, error) {

	i.rawImgReader.Seek(0, 0)
	m, _, err := image.Decode(i.rawImgReader)
	if err != nil {
		return nil, err
	}

	imageRect := m.Bounds()
	k := 1
	w := -128 //init
	h := -128 //init
	if w < 0 {
		w = -imageRect.Dx() * 72 / w / k
	}
	if h < 0 {
		h = -imageRect.Dy() * 72 / h / k
	}
	if w == 0 {
		w = h * imageRect.Dx() / imageRect.Dy()
	}
	if h == 0 {
		h = w * imageRect.Dy() / imageRect.Dx()
	}

	var rect = new(Rect)
	rect.H = float64(h)
	rect.W = float64(w)

	return rect, nil
}

func (i *ImageObj) parse() error {
	i.rawImgReader.Seek(0, 0)
	imginfo, err := parseImg(i.rawImgReader)
	if err != nil {
		return err
	}
	i.imginfo = imginfo

	return nil
}

//Parse parse img
func (i *ImageObj) Parse() error {
	return i.parse()
}

// hash returns the hash of the image object
func (i *ImageObj) procsetIdentifier() string {
	return i.procsetid
}
