# Base40 Encoder / Decoder for Golang

An optimized base 40 encoder/decoder module.

## Performance
Here we compare the encoder/decoder against the built in base64 module and the optimized one by cristalhq:
```
$ go test -bench=.
goos: linux
goarch: amd64
cpu: Intel(R) Xeon(R) CPU           X5650  @ 2.67GHz
BenchmarkBase40encode-12        16502744                63.13 ns/op
BenchmarkBase40decode-12        16755488                64.54 ns/op
BenchmarkCBase64decode-12       12064394                88.12 ns/op
BenchmarkCBase64encode-12       14831389                75.77 ns/op
BenchmarkBase64decode-12         8561384               119.8 ns/op
BenchmarkBase64encode-12         9217584               131.5 ns/op
PASS
ok      _/home/schou/git/go-b40 7.190s
```

## Usage
```golang
  dat := "helloworld"

  fmt.Printf("In: %s\n", dat)
  // In: helloworld

  db := CompressString(dat)
  fmt.Printf("Compress: %v\n", db)
  // Compress: [134 41 160 196 179 241 106 64]

  b := DecompressToString(db)
  fmt.Printf("Decompress: %v\n", b)
  // Decompress: helloworld

  b2 := []byte(b)
  ByteToB40(b2)
  fmt.Printf("b40: %v\n", b2)
  // b40: [21 18 25 25 28 36 28 31 25 17]

  B40ToByte(b2)
  fmt.Printf("byte: %v\n", b2)
  // byte: [104 101 108 108 111 119 111 114 108 100]
```
