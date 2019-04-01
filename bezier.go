package gofpdf

import "math"

const bezierSampleCardinality = 10000

type BezierCurve struct {
	Cx1, Cx2, Cx3, Cx4, Cy1, Cy2, Cy3, Cy4, Length float64
}

type BezierSpline []BezierCurve

type BezierSplineSample [][]float64

type BezierPoint struct {
	pt        Point
	normaldir float64
}

func NewBezierCurve(x0, y0, cx0, cy0, cx1, cy1, x1, y1 float64) BezierCurve {
	Cx1, Cx2, Cx3, Cx4 := Coefficients(x0, cx0, cx1, x1)
	Cy1, Cy2, Cy3, Cy4 := Coefficients(y0, cy0, cy1, y1)
	bc := BezierCurve{Cx1, Cx2, Cx3, Cx4, Cy1, Cy2, Cy3, Cy4, 0.0}
	bc.Length = CurveLength(bc)
	return bc
}

func NewBezierSpline(cp []Point) BezierSpline {
	var bs []BezierCurve
	// Consume groups of 4 points to create curve segments
	for len(cp) >= 4 {
		x0, y0 := cp[0].XY()
		cx0, cy0 := cp[1].XY()
		cx1, cy1 := cp[2].XY()
		x1, y1 := cp[3].XY()
		bs = append(bs, NewBezierCurve(x0, y0, cx0, cy0, cx1, cy1, x1, y1))
		cp = cp[3:] // Each curve's tail is also the previous curve's tip
	}
	return bs
}

func Coefficients(p0, p1, p2, p3 float64) (C1, C2, C3, C4 float64) {
	C1 = (p3 - (3.0 * p2) + (3.0 * p1) - p0)
	C2 = ((3.0 * p2) - (6.0 * p1) + (3.0 * p0))
	C3 = ((3.0 * p1) - (3.0 * p0))
	C4 = p0
	return
}

func (bc BezierCurve) At(t float64) Point {
	x := bc.Cx1*t*t*t + bc.Cx2*t*t + bc.Cx3*t + bc.Cx4
	y := bc.Cy1*t*t*t + bc.Cy2*t*t + bc.Cy3*t + bc.Cy4
	return Point{x, y}
}

func (bc BezierCurve) Curve(p []Point) []Point {
	// Returns a uniform sample with respect to the parameter t
	for i, nf, l := 0, float64(len(p)-1), len(p); i < l; i++ {
		p[i] = bc.At(float64(i) / nf)
	}
	return p
}

func CurveLength(bc BezierCurve) float64 {
	n := bezierSampleCardinality
	d := 0.0
	curve := make([]Point, n)
	// Approximate the curve by a polyline
	bc.Curve(curve)
	for len(curve) > 1 {
		d += Distance(curve[0], curve[1])
		curve = curve[1:]
	}
	return d
}

func Distance(p0, p1 Point) float64 {
	return math.Sqrt(math.Pow(p1.Y-p0.Y, 2) + math.Pow(p1.X-p0.X, 2))
}

func (bc BezierCurve) SampleByArcLength(sample []float64) []float64 {
	n := len(sample)
	d := 0.0
	curve := make([]Point, n)
	// Approximate the curve by a polyline
	bc.Curve(curve)
	polyline := curve
	distances := make([]float64, n-1)
	for len(curve) > 1 {
		distances[n-len(curve)] = Distance(curve[0], curve[1])
		d += distances[n-len(curve)]
		curve = curve[1:]
	}
	dd := d / float64(n-1)
	// Walk the polyline with even steps
	stride := dd
	sample[0] = 0.0
	i := 1
	for len(polyline) > 1 {
		if distances[0] >= stride {
			frac := stride / distances[0]
			t0 := float64(n-len(polyline)) / float64(n-1)
			t1 := float64(n-len(polyline)+1) / float64(n-1)
			t := t0 + frac*(t1-t0)
			sample[i] = t
			i++
			distances[0] -= stride
			stride = dd
		} else {
			stride -= distances[0]
			polyline = polyline[1:]
			distances = distances[1:]
		}
	}
	for i < len(sample) {
		sample[i] = 1.0
		i++
	}
	return sample
}

func (bc BezierCurve) Tangent(t float64) Point {
	dx := bc.Dx(t)
	dy := bc.Dy(t)
	return Point{dx, dy}
}

func (bc BezierCurve) NormalDegrees(t float64) float64 {
	tan := bc.Tangent(t)
	normal := Point{tan.Y, -1 * tan.X}
	return (math.Atan2(normal.Y, normal.X) * -180.0 / math.Pi) - 90.0
}

func (bc BezierCurve) Dx(t float64) float64 {
	return 3.0*bc.Cx1*t*t + 2.0*bc.Cx2*t + bc.Cx3
}

func (bc BezierCurve) Dy(t float64) float64 {
	return 3.0*bc.Cy1*t*t + 2.0*bc.Cy2*t + bc.Cy3
}

func (bs BezierSpline) SampleByArcLength(n int) BezierSplineSample {
	totalLength := bs.Length()
	clens := make([]int, len(bs))
	for i, bc := range bs {
		if i == len(bs)-1 {
			clens[i] = n
			break
		}
		// Extra point here is the endpoint which will be removed
		clens[i] = int((bc.Length/totalLength)*float64(n)) + 1
		n -= clens[i] - 1
		totalLength -= bc.Length
	}
	csamples := make([][]float64, len(bs))
	for i, cn := range clens {
		curve := make([]float64, cn)
		curve = bs[i].SampleByArcLength(curve)
		if i < len(clens)-1 && len(curve) > 0 {
			// Omit the final point of each but the last curve
			curve = curve[:len(curve)-1]
		}
		csamples[i] = curve
	}
	return csamples
}

func (bss BezierSplineSample) At(k int) (int, float64) {
	for i, c := range bss {
		if k < len(c) {
			return i, c[k]
		}
		k -= len(c)
	}
	return len(bss) - 1, 1.0
}

func (bs BezierSpline) Length() float64 {
	length := 0.0
	for _, bc := range bs {
		length += bc.Length
	}
	return length
}