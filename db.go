package ip2loc

import (
	"math/big"
	"net"
)

const countryOffset = 4

const (
	ipVersionInvalid ipVersion = iota
	ipVersion4
	ipVersion6
)

var (
	// https://en.wikipedia.org/wiki/IPv6#IPv4-mapped_IPv6_addresses
	ipv4mappedMin = big.NewInt(281470681743360)
	ipv4mappedMax = big.NewInt(281474976710655)

	// Gets set in init()
	v6v4Min = big.NewInt(0)
	v6v4Max = big.NewInt(0)

	// Teredo-tunneled ipv4 over ipv6.
	// Gets set in init()
	teredoMin = big.NewInt(0)
	teredoMax = big.NewInt(0)

	ls32bits = big.NewInt(4294967295)

	v4Max = big.NewInt(4294967295)
	v6Max = big.NewInt(0)
)

type (
	DB struct {
		r *reader

		v4IndexOffset uint32
		v6IndexOffset uint32

		v4DataLen uint32
		v6DataLen uint32

		v4DataOffset uint32
		v6DataOffset uint32

		v4ColumnWidth uint32
		v6ColumnWidth uint32
	}

	ipVersion = uint8

	Result struct {
		CountryCode string
		CountryName string
	}
)

func New(path string) (*DB, error) {
	r, err := newReader(path)
	if err != nil {
		return nil, err
	}

	if r.format != 1 {
		return nil, ErrUnsupportedFormat{
			badFormat: r.format,
		}
	}

	db := DB{
		r: r,
	}

	db.readMeta()

	return &db, nil
}

func (db *DB) Close() {
	if db.r != nil {
		db.r.Close()
	}
}

func (db *DB) Query(addr string) (*Result, error) {
	ver, ip := parseIP(addr)
	if ver == ipVersionInvalid {
		return nil, ErrInvalidIP{}
	}

	idx := db.ipIndex(ver, ip)
	offset, high, colsize, max := db.scanner(ver)

	var err error
	var low uint32

	if idx > 0 {
		low, err = db.r.readUint32(idx - 1)
		if err != nil {
			return nil, err
		}

		high, err = db.r.readUint32(idx + 3)
		if err != nil {
			return nil, err
		}
	}

	if ip.Cmp(max) >= 0 {
		ip.Sub(ip, big.NewInt(1))
	}

	for low <= high {
		mid := ((low + high) >> 1)
		rowoffset := offset + (mid * colsize)
		rowoffset2 := rowoffset + colsize

		var ipfrom *big.Int
		var ipto *big.Int

		if ver == ipVersion4 {
			v, err := db.r.readUint32(rowoffset - 1)
			if err != nil {
				return nil, err
			}

			ipfrom = big.NewInt(int64(v))

			v, err = db.r.readUint32(rowoffset2 - 1)
			if err != nil {
				return nil, err
			}
			ipto = big.NewInt(int64(v))
		} else if ver == ipVersion6 {
			ipfrom, err = db.r.readUint128(rowoffset - 1)
			if err != nil {
				return nil, err
			}

			ipto, err = db.r.readUint128(rowoffset2 - 1)
			if err != nil {
				return nil, err
			}
		}

		if ip.Cmp(ipfrom) >= 0 && ip.Cmp(ipto) < 0 {
			if ver == ipVersion6 {
				rowoffset += 12
			}

			countryIndex, err := db.r.readUint32(rowoffset + countryOffset - 1)
			if err != nil {
				return nil, err
			}

			ccCode, err := db.r.readString(countryIndex)
			if err != nil {
				return nil, err
			}

			ccName, err := db.r.readString(countryIndex + 3)
			if err != nil {
				return nil, err
			}

			return &Result{
				CountryCode: ccCode,
				CountryName: ccName,
			}, nil
		} else if ip.Cmp(ipfrom) < 0 {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return nil, ErrNoResults{}
}

func (db *DB) Version() (string, error) {
	return db.r.readVersion()
}

func (db *DB) ipIndex(ver ipVersion, ipint *big.Int) uint32 {
	if ver == ipVersionInvalid {
		return 0
	}

	if ver == ipVersion4 {
		if db.v4IndexOffset == 0 {
			return 0
		}

		idx := big.NewInt(0)
		idx.Rsh(ipint, 16)
		idx.Lsh(idx, 3)

		return uint32(idx.Add(idx, big.NewInt(int64(db.v4IndexOffset))).Uint64())
	}

	if ver == ipVersion6 {
		if db.v6IndexOffset == 0 {
			return 0
		}

		idx := big.NewInt(0)
		idx.Rsh(ipint, 112)
		idx.Lsh(idx, 3)

		return uint32(idx.Add(idx, big.NewInt(int64(db.v6IndexOffset))).Uint64())
	}

	return 0
}

func (db *DB) readMeta() (err error) {
	col, err := db.r.readUint8(1)
	if err != nil {
		return
	}
	db.v4ColumnWidth = uint32(col << 2)              // all columns are 4 bytes
	db.v6ColumnWidth = uint32(16 + ((col - 1) << 2)) // one column is 16 bytes, rest are 4 bytes

	db.v4DataLen, err = db.r.readUint32(5)
	if err != nil {
		return
	}

	db.v4DataOffset, err = db.r.readUint32(9)
	if err != nil {
		return
	}

	db.v6DataLen, err = db.r.readUint32(13)
	if err != nil {
		return
	}

	db.v6DataOffset, err = db.r.readUint32(17)
	if err != nil {
		return
	}

	db.v4IndexOffset, err = db.r.readUint32(21)
	if err != nil {
		return
	}

	db.v6IndexOffset, err = db.r.readUint32(25)
	if err != nil {
		return
	}

	return
}

func (db *DB) scanner(ver ipVersion) (uint32, uint32, uint32, *big.Int) {
	if ver == ipVersion4 {
		return db.v4DataOffset, db.v4DataLen, db.v4ColumnWidth, v4Max
	}

	if ver == ipVersion6 {
		return db.v6DataOffset, db.v6DataLen, db.v6ColumnWidth, v6Max
	}

	return 0, 0, 0, nil
}

func parseIP(addr string) (ipVersion, *big.Int) {
	ip := net.ParseIP(addr)
	if ip == nil {
		return ipVersionInvalid, nil
	}

	ipint := big.NewInt(0)

	if v4 := ip.To4(); v4 != nil {
		ipint.SetBytes(ip.To4())

		return ipVersion4, ipint
	}

	v6 := ip.To16()
	if v6 == nil {
		return ipVersionInvalid, nil
	}

	ipint.SetBytes(ip.To16())

	// Check for v4-mapped v6
	if ipint.Cmp(ipv4mappedMin) >= 0 && ipint.Cmp(ipv4mappedMax) <= 0 {
		return ipVersion4, ipint.Sub(ipint, ipv4mappedMin)
	}

	// Check for v4-embedded v6
	if ipint.Cmp(v6v4Min) >= 0 && ipint.Cmp(v6v4Max) <= 0 {
		ipint.Rsh(ipint, 80)
		ipint.And(ipint, ls32bits)
		return ipVersion4, ipint
	}

	// Check for v6-Teredo-tunneled v4
	if ipint.Cmp(teredoMin) >= 0 && ipint.Cmp(teredoMax) <= 0 {
		ipint.Not(ipint)
		ipint.And(ipint, ls32bits)
		return ipVersion4, ipint
	}

	// Looks like a normal v6
	return ipVersion6, ipint
}

func init() {
	v6v4Min.SetString("42545680458834377588178886921629466624", 10)
	v6v4Max.SetString("42550872755692912415807417417958686719", 10)

	teredoMin.SetString("42540488161975842760550356425300246528", 10)
	teredoMax.SetString("42540488241204005274814694018844196863", 10)

	v6Max.SetString("340282366920938463463374607431768211455", 10)
}
