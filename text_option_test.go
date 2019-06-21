package gofpdf

import (
	"testing"
)

func TestTextOpt(t *testing.T) {
	expectRenderMode(t, &TextOption{}, 0)
	expectRenderMode(t, &TextOption{Stroke: true, NoFill: true}, 1)
	expectRenderMode(t, &TextOption{Stroke: true}, 2)
	expectRenderMode(t, &TextOption{NoFill: true}, 3)
	expectRenderMode(t, &TextOption{Clip: true}, 4)
	expectRenderMode(t, &TextOption{Clip: true, Stroke: true, NoFill: true}, 5)
	expectRenderMode(t, &TextOption{Clip: true, Stroke: true}, 6)
	expectRenderMode(t, &TextOption{Clip: true, NoFill: true}, 7)

}

func expectRenderMode(t *testing.T, to *TextOption, rm int) {
	if to.GetRenderMode() != rm {
		t.Errorf("Render Mode was %d, Expecting %d", to.GetRenderMode(), rm)
	}
}
