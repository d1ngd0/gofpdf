package gofpdf

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/ISeeMe/gofpdf/bp"
)

//ImageHolder hold image data
type ImageHolder interface {
	ID() string
	io.Reader
}

//ImageHolderByBytes create ImageHolder by []byte
func ImageHolderByBytes(b []byte) (ImageHolder, error) {
	return newImageBuff(b)
}

//ImageHolderByPath create ImageHolder by image path
func ImageHolderByPath(path string) (ImageHolder, error) {
	return newImageBuffByPath(path)
}

//ImageHolderByReader create ImageHolder by io.Reader
func ImageHolderByReader(r io.Reader) (ImageHolder, error) {
	return newImageBuffByReader(r)
}

//imageBuff image holder (impl ImageHolder)
type imageBuff struct {
	id string
	*bytes.Buffer
}

func newImageBuff(b []byte) (*imageBuff, error) {
	i := imageBuff{
		Buffer: bp.GetBuffer(),
		id:     fmt.Sprintf("%x", sha1.Sum(b)),
	}
	i.Write(b)
	return &i, nil
}

func newImageBuffByPath(path string) (*imageBuff, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return newImageBuff(b)
}

func newImageBuffByReader(r io.Reader) (*imageBuff, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return newImageBuff(b)
}

func newImageBuffByURL(url string) (*imageBuff, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return newImageBuffByReader(resp.Body)
}

func (i *imageBuff) ID() string {
	return i.id
}
