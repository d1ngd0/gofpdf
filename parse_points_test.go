package gofpdf

import (
	"testing"
)

func TestParsePoints(t *testing.T) {
	_, err := ParsePoints("(1.98611, 6.90889) (1.98611, 6.90889) (2.73611, 6.71888) (3.21611, 6.85888) (3.69611, 6.99888) (4.35611, 6.85468) (4.35611, 6.85468)")
	if err != nil {
		t.Error(err)
	}

	_, err = ParsePoints("(1.98611, 6.90889) (1.98611, 6.90889) (2.73611 68) (3.21611, 6.85888) (3.69611, 6.99888) (4.35611, 6.85468) (4.35611, 6.85468)")
	if err == nil {
		t.Error("should have caught missing semicolon")
	}

	_, err = ParsePoints("(1.98611, 6.90889) (1.98611, 6.90889) (2.73611, 68) (3.21611, 6.85888 (3.69611, 6.99888) (4.35611, 6.85468) (4.35611, 6.85468)")
	if err == nil {
		t.Error("should have caught missing closing point")
	}

	_, err = ParsePoints("(1.98611, 6.90889) (1.98611, 6.90889) (2.73611, 6.71888) (3.21611, 6.85888) (3.69611, 6.99888) (4.35611, 6.85468, 4.35611, 6.85468)")
	if err == nil {
		t.Error("should have caught extra numbers")
	}
}

func BenchmarkParsePoints(b *testing.B) {
	for x := 0; x < b.N; x++ {
		_, err := ParsePoints("(1.98611, 6.90889) (1.98611, 6.90889) (2.73611, 6.71888) (3.21611, 6.85888) (3.69611, 6.99888) (4.35611, 6.85468)(4.35611, 6.85468)")
		if err != nil {
			b.Error(err)
		}
	}
}
