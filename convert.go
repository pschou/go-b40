package b40

import (
	"reflect"
	"unsafe"
)

var (
	byteToB40Lookup [3 * 256]uint16
	b40ToByteLookup [40 * 1600]uint32
	byteToB40Map    [256]byte
	keep24          = func() uint32 {
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

	// Mapping to use for B40 lookups
	B40ToByteMap = "\x00-.1234567890:abcdefghijklmnopqrstuvwxyz"
)

// Reset the lookup table with a different b40 conversion after setting the B40ToByteMap value.
func SetMap() {
	for i := range byteToB40Map {
		byteToB40Map[i] = 0
	}
	for i, c := range B40ToByteMap {
		byteToB40Map[c] = byte(i)
		byteToB40Lookup[c] |= uint16(i) * 1600
		byteToB40Lookup[256+c] |= uint16(i) * 40
		byteToB40Lookup[512+c] |= uint16(i)
	}
	for i := 0; i < 40*1600; i++ {
		dat := [4]byte{B40ToByteMap[i/1600], B40ToByteMap[(i/40)%40], B40ToByteMap[i%40], 0}
		b40ToByteLookup[i] = *(*uint32)(unsafe.Pointer(&dat))
	}
}

func init() {
	SetMap()
}

// Converting one byte at a time.
func B40ToByte(out, in []byte) {
	for i, c := range in {
		out[i] = B40ToByteMap[c]
	}
}

// Converting one byte at a time.
func ByteToB40(out, in []byte) {
	for i, c := range in {
		out[i] = byteToB40Map[c]
	}
}

// Compress a string into a b40 binary representation.
func CompressString(src string) (dst []byte) {
	ilen := len(src)
	olen := (ilen + 2) / 3 * 2
	dst = make([]byte, olen)
	return comp(dst, s2b(src), uintptr(olen), uintptr(ilen))
}

// Compress a slice to a b40 binary representation from a byte format.
//
// Note: The DST and SRC can be the same to prevent an additional malloc call.
// The returned slice is truncated to the proper length.
func Compress(dst, src []byte) []byte {
	ilen := len(src)
	olen := (ilen + 2) / 3 * 2
	return comp(dst, src, uintptr(olen), uintptr(ilen))
}

func comp(dst, src []byte, olen, ilen uintptr) []byte {
	ip := (*reflect.StringHeader)(unsafe.Pointer(&src)).Data
	op := (*reflect.SliceHeader)(unsafe.Pointer(&dst)).Data
	ipstop := ip + ilen - 2
	var c int
	for ip < ipstop {
		c = int(byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip)))]) +
			int(byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip + 1)))|0x0100]) +
			int(byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip + 2)))|0x0200])
		*(*uint16)(unsafe.Pointer(op)) = bswap16(uint16(c))
		op += 2
		ip += 3
	}
	switch ip - ipstop {
	case 0:
		c = int(byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip)))]) +
			int(byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip + 1)))|0x0100])
		*(*uint16)(unsafe.Pointer(op)) = bswap16(uint16(c))
	case 1:
		c = int(byteToB40Lookup[int(*(*byte)(unsafe.Pointer(ip)))])
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
func DecompressToString(src []byte) string {
	ilen := len(src)
	olen := (ilen + 1) / 2 * 3
	dst := make([]byte, olen+7)
	return b2s(decomp(dst, src, uintptr(olen), uintptr(ilen)))
}

// Decompress a slice from a b40 binary representation to the original byte format.
//
// Note: The DST and SRC can be the same to prevent an additional malloc call.
// The returned slice is truncated to the proper length.
func Decompress(dst, src []byte, srcLen int) []byte {
	olen := (srcLen + 1) / 2 * 3
	return decomp(dst, src, uintptr(olen), uintptr(srcLen))
}

func decomp(dst, src []byte, olen, ilen uintptr) []byte {
	ipstart := (*reflect.SliceHeader)(unsafe.Pointer(&src)).Data
	op := (*reflect.SliceHeader)(unsafe.Pointer(&dst)).Data + (ilen+1)/2*3 - 3
	ip := ipstart + ilen - 2
	blp := uintptr(unsafe.Pointer(&b40ToByteLookup))

	var j uint32
	j = *(*uint32)(unsafe.Pointer(blp + uintptr(bswap16(*(*uint16)(unsafe.Pointer(ip))))<<2))
	*(*uint32)(unsafe.Pointer(op)) = j
	ip -= 2
	for ip >= ipstart {
		op -= 3
		k := *(*uint32)(unsafe.Pointer(blp + uintptr(bswap16(*(*uint16)(unsafe.Pointer(ip))))<<2))
		*(*uint32)(unsafe.Pointer(op)) = k&keep24 | *(*uint32)(unsafe.Pointer(op)) & ^keep24
		ip -= 2
	}
	switch {
	case j&mask16 == 0:
		return dst[:olen-2]
	case j&mask8 == 0:
		return dst[:olen-1]
	}
	return dst[:olen]
}
