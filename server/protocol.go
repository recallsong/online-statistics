package server

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
)

var ErrorProtocol = errors.New("protocol error")
var ErrorPackageTooLarge = errors.New("package too large")

const (
	version    = 1
	headerSize = 8

	connectPkgType = 1
	connAckPkgType = 2
	pingPkgType    = 3
	pongPkgType    = 4
)

type pkgHeader struct {
	Version uint16
	Type    uint8
	Flags   uint8
	Length  uint32
	Data    []byte
}

type connectPkg struct {
	Id     string `json:"id"`
	Token  string `json:"token"`
	Domain string `json:"domain"`
}
type connAckPkg struct {
	Code      int16 `json:"code"`
	Keepalive int64 `json:"keepalive"`
}

func readPkg(rd io.Reader, buffer []byte) (error, *pkgHeader) {
	// read header
	reads := 0
	for {
		n, err := rd.Read(buffer[reads:8])
		if err != nil {
			return err, nil
		}
		reads += n
		if reads >= 8 {
			break
		}
	}
	pkg := &pkgHeader{}
	pkg.Version = binary.BigEndian.Uint16(buffer[0:2])
	pkg.Type = uint8(buffer[2])
	pkg.Flags = uint8(buffer[3])
	pkg.Length = binary.BigEndian.Uint32(buffer[4:8])
	if pkg.Version > version {
		return ErrorProtocol, pkg
	}
	if int(pkg.Length) > len(buffer) {
		return ErrorPackageTooLarge, pkg
	}
	var total int = int(pkg.Length)
	if total <= 0 {
		pkg.Data = nil
		return nil, pkg
	}
	// read data
	reads = 0
	for {
		n, err := rd.Read(buffer[reads:total])
		if err != nil {
			return err, nil
		}
		reads += n
		if reads >= total {
			break
		}
	}
	pkg.Data = buffer[0:total]
	return nil, pkg
}

func writePkg(wt io.Writer, buffer []byte, typ uint8, data []byte) error {
	if len(data) > len(buffer)-headerSize {
		return ErrorPackageTooLarge
	}
	binary.BigEndian.PutUint16(buffer, version)
	buffer[2] = typ
	buffer[3] = 0
	if len(data) > 0 {
		binary.BigEndian.PutUint32(buffer[4:], uint32(len(data)))
		copy(buffer[8:], data)
	} else {
		binary.BigEndian.PutUint32(buffer[4:], 0)
	}
	var total int = headerSize + len(data)
	writes := 0
	for {
		n, err := wt.Write(buffer[writes:total])
		if err != nil {
			return err
		}
		writes += n
		if writes >= total {
			break
		}
	}
	return nil
}

func readConnPackage(rd io.Reader, buffer []byte, pkg *connectPkg) error {
	err, header := readPkg(rd, buffer)
	if err != nil {
		return err
	}
	if header.Type != connectPkgType {
		return ErrorProtocol
	}
	err = json.Unmarshal(header.Data, pkg)
	if err != nil {
		return err
	}
	return nil
}

func writeConnAckPackage(wt io.Writer, buffer []byte, pkg *connAckPkg) error {
	data, err := json.Marshal(pkg)
	if err != nil {
		return err
	}
	return writePkg(wt, buffer, connAckPkgType, data)
}

func readPingPackage(rd io.Reader, buffer []byte) error {
	err, header := readPkg(rd, buffer)
	if err != nil {
		return err
	}
	if header.Type != pingPkgType {
		return ErrorProtocol
	}
	return nil
}

func writePongPackage(wt io.Writer, buffer []byte) error {
	return writePkg(wt, buffer, pongPkgType, nil)
}
