package gofpdf

import (
	"fmt"
	"log"
	"testing"
)

func TestBezierSpline(t *testing.T) {
	// Single segment
	fmt.Println("Testing single curve")
	p0, p1, p2, p3 := Point{1.0, 1.0}, Point{2.0, 2.0}, Point{3.0, 1.0}, Point{4.0, 2.0}
	pts := Points{p0, p1, p2, p3}
	bs := NewBezierSpline(pts)
	ExpectOutput(t, "bs", bs, "[{0 0 3 1 4 -6 3 1 3.268288235437555}]")

	o := make([]Point, 5)
	ExpectOutput(t, "sample", bs[0].Curve(o), "[{1 1} {1.75 1.4375} {2.5 1.5} {3.25 1.5625} {4 2}]")
	q := make([]float64, 5)
	ExpectOutput(t, "usample", bs[0].SampleByArcLength(q), "[0 0.23334661802430987 0.5 0.7666533819756902 1]")
	ExpectOutput(t, "t0.5", bs[0].At(0.5), "{2.5 1.5}")
	ExpectOutput(t, "d0.5", bs[0].Tangent(0.5), "{3 0}")

	// Multi-segment
	fmt.Println("Testing multiple curves")
	p0, p1, p2, p3, p4, p5, p6 := Point{1.0, 7.0}, Point{1.5, 8.0}, Point{1.5, 7.0}, Point{2.0, 8.0}, Point{2.5, 9.0}, Point{2.5, 8.0}, Point{3.0, 9.0}
	pts = Points{p0, p1, p2, p3, p4, p5, p6}
	bs = NewBezierSpline(pts)
	ExpectOutput(t, "bs", bs, "[{1 -1.5 1.5 1 4 -6 3 7 1.499989386099213} {1 -1.5 1.5 2 4 -6 3 8 1.4999893860992164}]")
	ExpectOutput(t, "length", bs.Length(), "2.9999787721984292") // convergence depends on bezierSplineCardinality
	ExpectOutput(t, "usample", bs.SampleByArcLength(5), "[[0 0.5] [0 0.5 1]]")

	p0, p1, p2, p3, p4, p5, p6, p7, p8, p9 := Point{1.0, 7.0}, Point{1.5, 8.0}, Point{1.5, 7.0}, Point{2.0, 8.0}, Point{2.5, 9.0}, Point{2.5, 8.0}, Point{3.0, 9.0}, Point{3.5, 9.5}, Point{4.0, 9.5}, Point{4.5, 10.0}
	pts = Points{p0, p1, p2, p3, p4, p5, p6, p7, p8, p9}
	bs = NewBezierSpline(pts)
	ExpectOutput(t, "bs", bs, "[{1 -1.5 1.5 1 4 -6 3 7 1.499989386099213} {1 -1.5 1.5 2 4 -6 3 8 1.4999893860992164} {0 0 1.5 3 1 -1.5 1.5 9 1.8120008506711385}]")
	ExpectOutput(t, "length", bs.Length(), "4.8119796228695675") // convergence depends on bezierSplineCardinality
	ExpectOutput(t, "usample", bs.SampleByArcLength(17), "[[0 0.12838267906952094 0.32880290143620794 0.6711970985637917 0.8716173209304788] [0 0.1283826790695211 0.3288029014362082 0.6711970985637918 0.871617320930479] [0 0.15317000199322645 0.3219339454960186 0.5 0.6780660545039813 0.8468299980067736 1]]")
}

func ExpectOutput(t *testing.T, name string, a interface{}, b string) {
	if fmt.Sprintf("%v", a) != b {
		t.Errorf("Error for %v: Expected %v, got %v\n", name, b, a)
		return
	}
	log.Printf("Passed %v: %v\n", name, b)
}