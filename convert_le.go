//go:build 386 || amd64 || amd64p32 || arm || arm64 || mipsle || mips64le || mips64p32le || ppc64le || riscv || riscv64 || wasm
// +build 386 amd64 amd64p32 arm arm64 mipsle mips64le mips64p32le ppc64le riscv riscv64 wasm

package b40

import "math/bits"

//go:nosplit
func bswap16(d uint16) uint16 {
	return bits.ReverseBytes16(d)
}
