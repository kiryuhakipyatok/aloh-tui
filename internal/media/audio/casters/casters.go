package casters

import (
	"encoding/binary"
	"math"
)

func MixBytesToInt16(int16s []int16, b []byte) {
	samples := len(b) / 2
	for i := 0; i < samples && i < len(int16s); i++ {
		val := int16(binary.LittleEndian.Uint16(b[i*2 : i*2+2]))
		mixed := int32(int16s[i]) + int32(val)
		if mixed > math.MaxInt16 {
			mixed = math.MaxInt16
		} else if mixed < math.MinInt16 {
			mixed = math.MinInt16
		}
		int16s[i] = int16(mixed)
	}
}

func MixToInt16(int16s []int16, tomix []int16) {
	for i := 0; i < len(tomix) && i < len(int16s); i++ {
		mixed := int32(int16s[i]) + int32(tomix[i])
		if mixed > math.MaxInt16 {
			mixed = math.MaxInt16
		} else if mixed < math.MinInt16 {
			mixed = math.MinInt16
		}
		int16s[i] = int16(mixed)
	}
}

func Int16ToBytes(int16s []int16, b []byte) {
	samples := len(int16s)
	for i := 0; i < samples && i < len(b); i++ {
		binary.LittleEndian.PutUint16(b[i*2:i*2+2], uint16(int16s[i]))
	}
}

func BytesU8ToInt16(int16s []int16, b []byte) {
	samples := len(b)
	for i := 0; i < samples && i < len(int16s); i++ {
		int16s[i] = int16(int8(b[i])) << 8
	}
}

func BytesS16ToInt16(int16s []int16, b []byte) {
	samples := len(b) / 2
	for i := 0; i < samples && i < len(int16s); i++ {
		val := int16(binary.LittleEndian.Uint16(b[i*2 : i*2+2]))
		int16s[i] = val
	}
}

func BytesS24ToInt16(int16s []int16, b []byte) {
	samples := len(b) / 3
	for i := 0; i < samples && i < len(int16s); i++ {
		u := uint32(b[i*3])<<8 | uint32(b[i*3+1])<<16 | uint32(b[i*3+2])<<24
		val := int32(u)
		int16s[i] = int16(val >> 16)
	}
}

func BytesS32ToInt16(int16s []int16, b []byte) {
	samples := len(b) / 4
	for i := 0; i < samples && i < len(b); i++ {
		u := binary.LittleEndian.Uint32(b[i*4 : i*4+4])
		val := int32(u)
		int16s[i] = int16(val >> 16)
	}
}

func BytesF32ToInt16(int16s []int16, b []byte) {
	samples := len(b) / 4
	for i := 0; i < samples && i < len(int16s); i++ {
		bits := binary.LittleEndian.Uint32(b[i*4 : i*4+4])
		f := math.Float32frombits(bits)
		scaled := f * 32767.0
		if scaled > math.MaxInt16 {
			scaled = math.MaxInt16
		} else if scaled < math.MinInt16 {
			scaled = math.MinInt16
		}
		int16s[i] = int16(scaled)
	}
}

func Float32ToInt16(ints16 []int16, floats []float32) {
	for i := 0; i < len(floats) && i < len(ints16); i++ {
		f := floats[i]
		if f > math.MaxInt16 {
			f = math.MaxInt16
		} else if f < math.MinInt16 {
			f = math.MinInt16
		}
		ints16[i] = int16(f)
	}
}

func Int16ToFloat32(ints16 []int16, floats []float32) {
	for i := 0; i < len(floats) && i < len(ints16); i++ {
		intt := ints16[i]
		floats[i] = float32(intt)
	}
}
