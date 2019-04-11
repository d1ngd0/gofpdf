package gofpdf

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

const (
	open_pointer   = '('
	close_pointer  = ')'
	seperate_point = ','
)

func ParsePoints(p string) ([]Point, error) {
	s := bufio.NewScanner(bytes.NewReader([]byte(p)))
	s.Split(pointSplitToken)

	pts := make([]Point, 0)
	var currentPoint *Point
	var currentVal *float64

	for s.Scan() {
		switch s.Text() {
		case string(open_pointer):
			if currentPoint != nil {
				return nil, errors.New("unexpected (, expecting , or number")
			}
			currentPoint = &Point{}
			currentVal = &currentPoint.X
		case string(close_pointer):
			if currentPoint == nil {
				return nil, errors.New("unexpected ), expecting (")
			}

			if currentVal != &currentPoint.Y {
				return nil, errors.New("unexpected ), expecting ,")
			}

			pts = append(pts, *currentPoint)
			currentPoint = nil
			currentVal = nil
		case string(seperate_point):
			if currentPoint == nil {
				return nil, errors.New("unexpected ,, expecting (")
			}

			if currentVal != &currentPoint.X {
				return nil, errors.New("unexpected ,, expecting a number or )")
			}

			currentVal = &currentPoint.Y
		default:
			if currentPoint == nil {
				return nil, errors.New("unexpected number, expecting (")
			}

			var err error
			if *currentVal, err = strconv.ParseFloat(s.Text(), 64); err != nil {
				return nil, fmt.Errorf("unexpect \"%s\", expecting number", s.Text())
			}
		}
	}

	return pts, s.Err()
}

func pointSplitToken(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for x, l := 0, len(data); x < l; x++ {
		switch data[x] {
		case ' ', '	':
			// ignore whitespace
		case open_pointer:
			// if the token has a value, we were reading a number
			// so we should see a closing or seperating pointer token
			if len(token) > 0 {
				return 0, nil, errors.New("unexpected (, expecting ) ,")
			}

			// token is empty, this is the first thing we saw,
			// good to go
			return x + 1, []byte{open_pointer}, nil
		case close_pointer:
			// if there is nothing in the token, we can assume all is fine
			if len(token) == 0 {
				return x + 1, []byte{close_pointer}, nil
			}

			// return the token, all is well
			return x, token, nil
		case seperate_point:
			// if there is nothing in the token, we can assume all is fine
			if len(token) == 0 {
				return x + 1, []byte{seperate_point}, nil
			}

			// return the token, all is well
			return x, token, nil
		default:
			token = append(token, data[x])
		}
	}

	if len(token) > 0 && atEOF {
		return 0, nil, errors.New("unexpected eof")
	}

	return 0, nil, nil
}
