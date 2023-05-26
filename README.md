# Base-40

I designed the base40 encoding in 2020 to solve a challenge of representing 3 character in 2 bytes of data while maintaining byte sort order.  It's simple as it fits very well in the uint16 as in 40x40x40 = 64000 while the max uint16 is 65535.

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
```
