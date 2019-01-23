package gofpdf

import (
	"fmt"
	"io"
	"strings"
)

type cacheContentDrawPath struct {
	styleStr string
}

func (c *cacheContentDrawPath) write(w io.Writer, protection *PDFProtection) error {
	var opStr string

	switch strings.ToUpper(c.styleStr) {
	case "", "D":
		// Stroke the path.
		opStr = "S"
	case "F":
		// fill the path, using the nonzero winding number rule
		opStr = "f"
	case "F*":
		// fill the path, using the even-odd rule
		opStr = "f*"
	case "FD", "DF":
		// fill and then stroke the path, using the nonzero winding number rule
		opStr = "B"
	case "FD*", "DF*":
		// fill and then stroke the path, using the even-odd rule
		opStr = "B*"
	default:
		opStr = c.styleStr
	}

	fmt.Fprintf(w, "%s\n", opStr)

	return nil
}
