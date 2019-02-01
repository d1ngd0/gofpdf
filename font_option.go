package gofpdf

import (
	"strings"
)

const (
	//Regular - font style regular
	Regular = 0 //000000
	//Italic - font style italic
	Italic = 1 //000001
	//Bold - font style bold
	Bold = 2 //000010
	//Underline - font style underline
	Underline = 4 //000100
)

func GetConvertedStyle(fontStyle string) int {
	return getConvertedStyle(fontStyle)
}

func getConvertedStyle(fontStyle string) (style int) {
	fontStyle = strings.ToUpper(fontStyle)
	if strings.Contains(fontStyle, "B") {
		style = style | Bold
	}
	if strings.Contains(fontStyle, "I") {
		style = style | Italic
	}
	if strings.Contains(fontStyle, "U") {
		style = style | Underline
	}
	return
}
