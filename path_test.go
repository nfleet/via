package main

import (
	"testing"
)

func BenchmarkFinlandPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CalculatePath(253299, 762749, "finland", 60)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGermanyPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := CalculatePath(54184, 3165075, "germany", 60)
		if err != nil {
			b.Fatal(err)
		}
	}
}
