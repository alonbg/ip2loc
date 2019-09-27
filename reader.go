package ip2loc

import (
	"math/big"
	"os"
	"strconv"
)

type reader struct {
	file   *os.File
	format uint8
}

func newReader(path string) (*reader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	r := reader{
		file: f,
	}

	r.format, err = r.readUint8(0)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *reader) Close() {
	if r.file != nil {
		r.file.Close()
	}
}

func (r *reader) readString(idx uint32) (string, error) {
	lenbuf := make([]byte, 1)
	_, err := r.file.ReadAt(lenbuf, int64(idx))
	if err != nil {
		return "", err
	}

	buf := make([]byte, lenbuf[0])
	_, err = r.file.ReadAt(buf, int64(idx+1))
	if err != nil {
		return "", err
	}

	return string(buf[:lenbuf[0]]), nil
}

func (r *reader) readUint8(idx uint32) (uint8, error) {
	buf := make([]byte, 1)

	_, err := r.file.ReadAt(buf, int64(idx))
	if err != nil {
		return 0, err
	}

	return buf[0], nil
}

func (r *reader) readUint32(idx uint32) (uint32, error) {
	buf := make([]byte, 4)

	_, err := r.file.ReadAt(buf, int64(idx))
	if err != nil {
		return 0, err
	}

	return uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24, nil
}

func (r *reader) readUint128(idx uint32) (*big.Int, error) {
	buf := make([]byte, 16)

	_, err := r.file.ReadAt(buf, int64(idx))
	if err != nil {
		return nil, err
	}

	// LE to BE
	for i, j := 0, 15; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	v := big.NewInt(0)
	v.SetBytes(buf)
	return v, nil
}

func (r *reader) readVersion() (string, error) {
	ymd := make([]byte, 3)
	_, err := r.file.ReadAt(ymd, 2)
	if err != nil {
		return "", err
	}

	year := strconv.Itoa(int(ymd[0]))
	month := strconv.Itoa(int(ymd[1]))
	day := strconv.Itoa(int(ymd[2]))

	return year + "-" + month + "-" + day, nil
}
