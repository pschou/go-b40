package b40

import (
	"reflect"
	"strings"
	"unsafe"
)

type Encoding struct {
	// Mapping to use for B40 lookups
	byteToB40Lookup [3 * 256]uint16
	b40ToByteLookup [3 * 40 * 1600]byte
	byteToB40Map    [256]byte
	b40ToByteMap    [40]byte
}

var (
	keep24 = func() uint32 {
		dat := [8]byte{0xff, 0xff, 0xff, 0}
		return *(*uint32)(unsafe.Pointer(&dat))
	}()
	mask24 = func() uint32 {
		dat := [8]byte{0, 0, 0, 0xff}
		return *(*uint32)(unsafe.Pointer(&dat))
	}()
	mask16 = func() uint32 {
		dat := [8]byte{0, 0, 0xff, 0xff}
		return *(*uint32)(unsafe.Pointer(&dat))
	}()
	mask8 = func() uint32 {
		dat := [8]byte{0, 0, 0, 0xff}
		return *(*uint32)(unsafe.Pointer(&dat))
	}()

	// The standard is inspired by the limitations imposed by RFC1034 and RFC2396
	Standard = NewFoldedEncoding("\x00-.1234567890:abcdefghijklmnopqrstuvwxyz")
)

// Create an encoding with the lookup table with a different b40 conversion after setting the B40ToByteMap value.
func NewEncoding(B40ToByteMap string) (e *Encoding) {
	e = &Encoding{}
	copy(e.b40ToByteMap[:], s2b(B40ToByteMap))
	for i, c := range B40ToByteMap {
		e.byteToB40Map[c] = byte(i)
		e.byteToB40Lookup[c] |= uint16(i) * 1600
		e.byteToB40Lookup[256+c] |= uint16(i) * 40
		e.byteToB40Lookup[512+c] |= uint16(i)
	}
	for i, j := 0, 0; j < 40*1600; i, j = i+3, j+1 {
		e.b40ToByteLookup[i], e.b40ToByteLookup[i+1], e.b40ToByteLookup[i+2] =
			B40ToByteMap[j/1600], B40ToByteMap[(j/40)%40], B40ToByteMap[j%40]
	}
	return
}

// Create an encoding the lookup table with a different b40 conversion after setting the B40ToByteMap value.
func NewFoldedEncoding(B40ToByteMap string) (e *Encoding) {
	e = &Encoding{}
	copy(e.b40ToByteMap[:], s2b(B40ToByteMap))
	for i, c := range strings.ToLower(B40ToByteMap) {
		e.byteToB40Map[c] = byte(i)
		e.byteToB40Lookup[c] |= uint16(i) * 1600
		e.byteToB40Lookup[256+c] |= uint16(i) * 40
		e.byteToB40Lookup[512+c] |= uint16(i)
	}
	for i, c := range strings.ToUpper(B40ToByteMap) {
		e.byteToB40Map[c] = byte(i)
		e.byteToB40Lookup[c] |= uint16(i) * 1600
		e.byteToB40Lookup[256+c] |= uint16(i) * 40
		e.byteToB40Lookup[512+c] |= uint16(i)
	}
	for i, j := 0, 0; j < 40*1600; i, j = i+3, j+1 {
		e.b40ToByteLookup[i], e.b40ToByteLookup[i+1], e.b40ToByteLookup[i+2] =
			B40ToByteMap[j/1600], B40ToByteMap[(j/40)%40], B40ToByteMap[j%40]
	}
	return
}

// Converting one byte at a time.
func (e *Encoding) B40ToByte(out, in []byte) {
	for i, c := range in {
		out[i] = e.b40ToByteMap[c]
	}
}

// Converting one byte at a time.
func (e *Encoding) ByteToB40(out, in []byte) {
	for i, c := range in {
		out[i] = e.byteToB40Map[c]
	}
}

// Compress a string into a b40 binary representation.
func (e *Encoding) CompressString(src string) (dst []byte) {
	ilen := len(src)
	olen := (ilen + 2) / 3 * 2
	dst = make([]byte, olen)
	return e.comp(dst, s2b(src), uintptr(olen), uintptr(ilen))
}

// Compress a slice to a b40 binary representation from a byte format.
//
// Note: The DST and SRC can be the same to prevent an additional malloc call.
// The returned slice is truncated to the proper length.
func (e *Encoding) Compress(dst, src []byte) []byte {
	ilen := len(src)
	olen := (ilen + 2) / 3 * 2
	return e.comp(dst, src, uintptr(olen), uintptr(ilen))
}

func (e *Encoding) comp(dst, src []byte, olen, ilen uintptr) []byte {
	ip := (*reflect.StringHeader)(unsafe.Pointer(&src)).Data
	op := (*reflect.SliceHeader)(unsafe.Pointer(&dst)).Data
	ipstop := ip + ilen - 2
	var c int
	for ip < ipstop {
		c = int(e.byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip)))]) +
			int(e.byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip + 1)))|0x0100]) +
			int(e.byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip + 2)))|0x0200])
		*(*uint16)(unsafe.Pointer(op)) = bswap16(uint16(c))
		op += 2
		ip += 3
	}
	switch ip - ipstop {
	case 0:
		c = int(e.byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip)))]) +
			int(e.byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip + 1)))|0x0100])
		*(*uint16)(unsafe.Pointer(op)) = bswap16(uint16(c))
	case 1:
		c = int(e.byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip)))])
		*(*uint16)(unsafe.Pointer(op)) = bswap16(uint16(c))
	}
	return dst[:olen]
}

func b2s(value []byte) string {
	return *(*string)(unsafe.Pointer(&value))
}

func s2b(value string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&value))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

// Decompress a string from a b40 binary representation.
func (e *Encoding) DecompressToString(src []byte) string {
	ilen := len(src)
	olen := (ilen + 1) / 2 * 3
	dst := make([]byte, olen+7)
	return b2s(e.decomp(dst, src, uintptr(olen), uintptr(ilen)))
}

// Decompress a slice from a b40 binary representation to the original byte format.
//
// Note: The DST and SRC can be the same to prevent an additional malloc call.
// The returned slice is truncated to the proper length.
func (e *Encoding) Decompress(dst, src []byte, srcLen int) []byte {
	olen := (srcLen + 1) / 2 * 3
	return e.decomp(dst, src, uintptr(olen), uintptr(srcLen))
}

func (e *Encoding) decomp(dst, src []byte, olen, ilen uintptr) []byte {
	ipstart := (*reflect.SliceHeader)(unsafe.Pointer(&src)).Data
	op := (*reflect.SliceHeader)(unsafe.Pointer(&dst)).Data + (ilen+1)/2*3 - 3
	ip := ipstart + ilen - 2
	blp := uintptr(unsafe.Pointer(&e.b40ToByteLookup))

	i := blp + uintptr(bswap16(*(*uint16)(unsafe.Pointer(ip))))*3
	*(*byte)(unsafe.Pointer(op)) = *(*byte)(unsafe.Pointer(i))
	a := *(*byte)(unsafe.Pointer(i + 1))
	*(*byte)(unsafe.Pointer(op + 1)) = a
	b := *(*byte)(unsafe.Pointer(i + 2))
	*(*byte)(unsafe.Pointer(op + 2)) = b

	for ip -= 2; ip >= ipstart; ip -= 2 {
		op -= 3
		i = blp + uintptr(bswap16(*(*uint16)(unsafe.Pointer(ip))))*3
		*(*byte)(unsafe.Pointer(op)) = *(*byte)(unsafe.Pointer(i))
		*(*byte)(unsafe.Pointer(op + 1)) = *(*byte)(unsafe.Pointer(i + 1))
		*(*byte)(unsafe.Pointer(op + 2)) = *(*byte)(unsafe.Pointer(i + 2))
	}

	if a == 0 {
		if b == 0 {
			return dst[:olen-2]
		}
		return dst[:olen-1]
	}
	return dst[:olen]
}
