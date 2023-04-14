package audio

//
//import (
//	"encoding/binary"
//	"github.com/mearaj/audio/main/orig"
//	"io"
//	"sync"
//)
//
//type VirtualFile interface {
//	io.Reader
//	io.ReaderAt
//	Bytes() []byte
//	Reset()
//	io.Writer
//	io.WriterAt
//	io.Seeker
//	Clear()
//}
//
//type virtualFile struct {
//	bytesMutex sync.RWMutex
//	posMutex   sync.RWMutex
//	pos        int
//	bytes      []byte
//	io.SectionReader
//}
//
//func NewVirtualFile(b []byte) VirtualFile {
//	return &virtualFile{
//		bytesMutex: sync.RWMutex{},
//		bytes:      b,
//	}
//}
//func NewVF(b []byte) VirtualFile {
//	return NewVirtualFile(b)
//}
//
//func (vf *virtualFile) setPos(pos int) {
//	vf.posMutex.Lock()
//	defer vf.posMutex.Unlock()
//	vf.pos = pos
//}
//
//func (vf *virtualFile) setBytes(b []byte) {
//	vf.bytesMutex.Lock()
//	defer vf.bytesMutex.Unlock()
//	vf.bytes = b
//}
//
//func (vf *virtualFile) Bytes() []byte {
//	vf.bytesMutex.RLock()
//	defer vf.bytesMutex.RUnlock()
//	return vf.bytes
//}
//
//func (vf *virtualFile) Pos() int {
//	vf.posMutex.RLock()
//	defer vf.posMutex.RUnlock()
//	return vf.pos
//}
//
//// Read satisfies io.Reader interface
//func (vf *virtualFile) Read(b []byte) (int, error) {
//	pos := vf.Pos()
//	n, err := vf.readAt(b, int64(pos))
//	pos += n
//	vf.setPos(pos)
//	return n, err
//}
//
//// ReadAt satisfies io.ReaderAt
//func (vf *virtualFile) ReadAt(b []byte, offset int64) (int, error) {
//	return vf.readAt(b, offset)
//}
//
//func (vf *virtualFile) readAt(b []byte, offset int64) (int, error) {
//	if offset < 0 {
//		return 0, orig.ErrInvalidOffset
//	}
//	if offset > int64(len(vf.Bytes())) {
//		return 0, io.EOF
//	}
//	n := copy(b, vf.Bytes()[offset:])
//	if n < len(b) {
//		return n, io.EOF
//	}
//	return len(b), nil
//}
//
//func (vf *virtualFile) Write(b []byte) (int, error) {
//	pos := vf.Pos()
//	n, err := vf.writeAt(b, int64(pos))
//	pos += n
//	vf.setPos(pos)
//	return n, err
//}
//
//func (vf *virtualFile) WriteAt(b []byte, offset int64) (int, error) {
//	return vf.writeAt(b, offset)
//}
//
//func (vf *virtualFile) writeAt(b []byte, offset int64) (int, error) {
//	if offset < 0 {
//		return 0, orig.ErrInvalidOffset
//	}
//	if offset > int64(len(vf.Bytes())) {
//		err := vf.resize(offset)
//		if err != nil {
//			return 0, err
//		}
//	}
//	n := copy(vf.Bytes()[offset:], b)
//	bytes := append(vf.Bytes(), b[n:]...)
//	vf.setBytes(bytes)
//	return len(b), nil
//}
//
//// Seek behavior is similar to io.Seeker,
//// allowed whence value is io.SeekStart, io.SeekCurrent and io.SeekEnd
//func (vf *virtualFile) Seek(offset int64, whence int) (int64, error) {
//	var absolute int64
//	switch whence {
//	case io.SeekStart:
//		absolute = offset
//	case io.SeekCurrent:
//		absolute = int64(vf.Pos()) + offset
//	case io.SeekEnd:
//		absolute = int64(len(vf.Bytes())) + offset
//	default:
//		return 0, orig.ErrInvalidWhence
//	}
//	if absolute < 0 {
//		return 0, orig.ErrInvalidOffset
//	}
//	vf.setPos(int(absolute))
//	return absolute, nil
//}
//
//func (vf *virtualFile) Truncate(n int64) error {
//	return vf.resize(n)
//}
//
//// resize resizes based on comparison with len(bf.bytes)
//func (vf *virtualFile) resize(n int64) error {
//	switch {
//	case n < 0:
//		return orig.ErrInvalidOffset
//	case n <= int64(len(vf.Bytes())):
//		vf.setBytes(vf.Bytes()[:n])
//		return nil
//	default:
//		bytes := append(vf.Bytes(), make([]byte, int(n)-len(vf.Bytes()))...)
//		vf.setBytes(bytes)
//		return nil
//	}
//}
//
//func (vf *virtualFile) WriteFloat(p []float32) (int, error) {
//	return len(p), binary.Write(vf, binary.LittleEndian, p)
//}
//
//func (vf *virtualFile) ReadFloat(out []float32) (int, error) {
//	return len(out), binary.Read(vf, binary.LittleEndian, out)
//}
//
//// Reset the position to 0
//func (vf *virtualFile) Reset() {
//	vf.setPos(0)
//}
//
//// Clear clears the bytes and sets pos to 0
//func (vf *virtualFile) Clear() {
//	vf.Reset()
//	vf.setBytes([]byte{})
//}
//
//// Close is noop
//func (vf *virtualFile) Close() error {
//	return nil
//}
