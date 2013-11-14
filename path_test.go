package main

import (
	"testing"
)

func TestPath(t *testing.T) {
	_, err := CalculatePath(253299, 762749, "finland", 60)

	if err != nil {
		t.Fatal(err)
	}
}

func BenchmarkPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CalculatePath(253299, 762749, "finland", 60)
		if err != nil {
			b.Fatal(err)
		}
	}
}
