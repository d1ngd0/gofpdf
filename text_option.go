package gofpdf

// TextOption Text Rendering Options
type TextOption struct {
	CharacterSpacing float64 // character spacing for text
	WordSpacing      float64 // word spacing for text
	Rise             float64 // sub/super scripting of fonts
	NoFill           bool    // render the filled text
	Stroke           bool    // render the stroke of the text
	Clip             bool    // Use Text as a clipping path used in conjuction with fill and stroke to determine render mode
}

func (c *TextOption) GetRenderMode() int {
	mode := 0
	if c.NoFill && c.Stroke {
		mode = 1
	} else if c.Stroke && !c.NoFill {
		mode = 2
	} else if c.NoFill && !c.Stroke {
		mode = 3
	}
	if c.Clip {
		mode += 4
	}
	return mode
}
