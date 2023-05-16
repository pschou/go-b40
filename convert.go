package b40

import (
	"reflect"
	"unsafe"
)

var (
	byteToB40Lookup [3 * 256]uint16
	b40ToByteLookup [40 * 1600]uint32
	byteToB40Map    [256]byte
	//mask24          = func() uint {
	//	dat := [8]byte{0xff, 0xff, 0xff, 0, 0, 0, 0, 0}
	//	return *(*uint)(unsafe.Pointer(&dat))
	//}()
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

// Reset the lookup table with a different b40 conversion
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

func B40ToByte(dat []byte) []byte {
	for i, c := range dat {
		dat[i] = B40ToByteMap[c]
	}
	return dat
}

func ByteToB40(dat []byte) []byte {
	for i, c := range dat {
		dat[i] = byteToB40Map[c]
	}
	return dat
}

/*func B40ToByte(dat []byte) []byte {
	ip := (*reflect.StringHeader)(unsafe.Pointer(&dat)).Data
	m := &B40ToByteMap
	ipstop := ip + uintptr(len(dat))
	for ip < ipstop {
		*(*byte)(unsafe.Pointer(ip)) = *(*byte)(unsafe.Pointer(m + uintptr(*(*byte)(unsafe.Pointer(ip)))))
		ip++
	}
	return src
}

func ByteToB40(src []byte) []byte {
	ip := (*reflect.StringHeader)(unsafe.Pointer(&src)).Data
	ipstop := ip + uintptr(len(src))
	for ip < ipstop {
		*(*byte)(unsafe.Pointer(ip)) = byteToB40Map[*(*byte)(unsafe.Pointer(ip))]
		ip++
	}
	return src
}*/

func CompressString(src string) (dst []byte) {
	ilen := len(src)
	olen := (ilen + 2) / 3 * 2
	dst = make([]byte, olen)
	ip := (*reflect.StringHeader)(unsafe.Pointer(&src)).Data
	op := (*reflect.SliceHeader)(unsafe.Pointer(&dst)).Data
	ipstop := ip + uintptr(ilen) - 2
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
	return
}

func b2s(value []byte) string {
	return *(*string)(unsafe.Pointer(&value))
}

func DecompressToString(src []byte) string {
	ilen := len(src)
	olen := (ilen + 1) / 2 * 3
	dst := make([]byte, olen+7)
	ip := (*reflect.SliceHeader)(unsafe.Pointer(&src)).Data
	op := (*reflect.StringHeader)(unsafe.Pointer(&dst)).Data
	ipstop := ip + uintptr(ilen) - 1
	blp := uintptr(unsafe.Pointer(&b40ToByteLookup))

	var k uint32
	//kp := uintptr(unsafe.Pointer(&k))
	for ip < ipstop {
		//k = *(*uint)(unsafe.Pointer(blp + uintptr(bswap16(*(*uint16)(unsafe.Pointer(ip))))<<2))
		//*(*byte)(unsafe.Pointer(op)) = *(*byte)(unsafe.Pointer(kp))
		//*(*byte)(unsafe.Pointer(op + 1)) = *(*byte)(unsafe.Pointer(kp + 1))
		//*(*byte)(unsafe.Pointer(op + 2)) = *(*byte)(unsafe.Pointer(kp + 2))
		//k = b40ToByteLookup[bswap16(*(*uint16)(unsafe.Pointer(ip)))]
		k = *(*uint32)(unsafe.Pointer(blp + uintptr(bswap16(*(*uint16)(unsafe.Pointer(ip))))<<2))
		*(*uint32)(unsafe.Pointer(op)) = k
		op += 3
		ip += 2
	}
	switch {
	case k&mask16 == 0:
		return b2s(dst[:olen-2])
	case k&mask8 == 0:
		return b2s(dst[:olen-1])
	}
	return b2s(dst[:olen])
}
