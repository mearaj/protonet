package service

import (
	"encoding/binary"
	"github.com/jfreymuth/pulse"
)

type AudioBuffer struct {
	buffer []byte
	pos    int
}

func (rw *AudioBuffer) Read(p []byte) (n int, err error) {
	if rw.pos >= len(rw.buffer)-1 {
		return rw.pos, pulse.EndOfData
	}
	pos := rw.pos
	for i := 0; i < len(p); i++ {
		if i+pos >= len(rw.buffer) {
			return i, pulse.EndOfData
		}
		p[i] = rw.buffer[i+pos]
		rw.pos++
	}
	return len(p), nil
}

func (rw *AudioBuffer) SetPos(pos int) {
	rw.pos = pos
}

func (rw *AudioBuffer) GetBuffer() []byte {
	return rw.buffer
}

func (rw *AudioBuffer) Write(p []byte) (n int, err error) {
	minCap := rw.pos + len(p)
	if minCap > cap(rw.buffer) { // Make sure buf has enough capacity:
		buf2 := make([]byte, len(rw.buffer), minCap+len(p)) // add some extra
		copy(buf2, rw.buffer)
		rw.buffer = buf2
	}
	if minCap > len(rw.buffer) {
		rw.buffer = rw.buffer[:minCap]
	}
	copy(rw.buffer[rw.pos:], p)
	rw.pos += len(p)
	return len(p), nil
}

func (rw *AudioBuffer) WriteFloat(p []float32) (int, error) {
	return len(p), binary.Write(rw, binary.LittleEndian, p)
}
func (rw *AudioBuffer) ReadFloat(out []float32) (int, error) {
	return len(out), binary.Read(rw, binary.LittleEndian, out)
}
