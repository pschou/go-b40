package b40

import (
	"encoding/base64"
	"fmt"
	"testing"

	cbase64 "github.com/cristalhq/base64"
)

func ExampleCompressString() {
	dat := "helloworld"

	fmt.Printf("In: %s\n", dat)
	db := CompressString(dat)
	fmt.Printf("Compress: %v\n", db)
	b := DecompressToString(db)
	fmt.Printf("Decompress: %v\n", b)
	b2 := []byte(b)
	ByteToB40(b2, b2) // output, input
	fmt.Printf("b40: %v\n", b2)
	B40ToByte(b2, b2) // output, input
	fmt.Printf("byte: %v\n", b2)
	// Output:
	// In: helloworld
	// Compress: [134 41 160 196 179 241 106 64]
	// Decompress: helloworld
	// b40: [21 18 25 25 28 36 28 31 25 17]
	// byte: [104 101 108 108 111 119 111 114 108 100]
}

func BenchmarkBase40encode(b *testing.B) {
	dat := "helloworld"
	for n := 0; n < b.N; n++ {
		CompressString(dat)
	}
}
func BenchmarkBase40decode(b *testing.B) {
	isB40 := []byte{134, 41, 160, 196, 179, 241, 106, 64}
	for n := 0; n < b.N; n++ {
		DecompressToString(isB40)
	}
}
func BenchmarkCBase64decode(b *testing.B) {
	isBase64 := `VmFsaWQgc3RyaW5nCg==`
	for n := 0; n < b.N; n++ {
		cbase64.StdEncoding.DecodeString(isBase64)
	}
}

func BenchmarkCBase64encode(b *testing.B) {
	dat := []byte("helloworld")
	for n := 0; n < b.N; n++ {
		cbase64.StdEncoding.EncodeToString(dat)
	}
}
func BenchmarkBase64decode(b *testing.B) {
	isBase64 := `VmFsaWQgc3RyaW5nCg==`
	for n := 0; n < b.N; n++ {
		base64.StdEncoding.DecodeString(isBase64)
	}
}

func BenchmarkBase64encode(b *testing.B) {
	dat := []byte("helloworld")
	for n := 0; n < b.N; n++ {
		base64.StdEncoding.EncodeToString(dat)
	}
}
