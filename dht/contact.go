package dht

import (
	"bytes"
	"net"

	"github.com/lbryio/lbry.go/errors"
	"github.com/lbryio/reflector.go/dht/bits"

	"github.com/lyoshenka/bencode"
)

// TODO: if routing table is ever empty (aka the node is isolated), it should re-bootstrap

// TODO: use a tree with bucket splitting instead of a fixed bucket list. include jack's optimization (see link in commit mesg)
// https://github.com/lbryio/lbry/pull/1211/commits/341b27b6d21ac027671d42458826d02735aaae41

// Contact is a type representation of another node that a specific node is in communication with.
type Contact struct {
	ID   bits.Bitmap
	IP   net.IP
	Port int
}

// Equals returns T/F if two contacts are the same.
func (c Contact) Equals(other Contact, checkID bool) bool {
	return c.IP.Equal(other.IP) && c.Port == other.Port && (!checkID || c.ID == other.ID)
}

// Addr returns the UPD Address of the contact.
func (c Contact) Addr() *net.UDPAddr {
	return &net.UDPAddr{IP: c.IP, Port: c.Port}
}

// String returns the concatenated short hex encoded string of its ID + @ + string represention of its UPD Address.
func (c Contact) String() string {
	return c.ID.HexShort() + "@" + c.Addr().String()
}

// MarshalCompact returns the compact byte slice representation of a contact.
func (c Contact) MarshalCompact() ([]byte, error) {
	if c.IP.To4() == nil {
		return nil, errors.Err("ip not set")
	}
	if c.Port < 0 || c.Port > 65535 {
		return nil, errors.Err("invalid port")
	}

	var buf bytes.Buffer
	buf.Write(c.IP.To4())
	buf.WriteByte(byte(c.Port >> 8))
	buf.WriteByte(byte(c.Port))
	buf.Write(c.ID[:])

	if buf.Len() != compactNodeInfoLength {
		return nil, errors.Err("i dont know how this happened")
	}

	return buf.Bytes(), nil
}

// UnmarshalCompact unmarshals the compact byte slice representation of a contact.
func (c *Contact) UnmarshalCompact(b []byte) error {
	if len(b) != compactNodeInfoLength {
		return errors.Err("invalid compact length")
	}
	c.IP = net.IPv4(b[0], b[1], b[2], b[3]).To4()
	c.Port = int(uint16(b[5]) | uint16(b[4])<<8)
	c.ID = bits.FromBytesP(b[6:])
	return nil
}

// MarshalBencode returns the serialized byte slice representation of a contact.
func (c Contact) MarshalBencode() ([]byte, error) {
	return bencode.EncodeBytes([]interface{}{c.ID, c.IP.String(), c.Port})
}

// UnmarshalBencode unmarshals the serialized byte slice into the appropriate fields of the contact.
func (c *Contact) UnmarshalBencode(b []byte) error {
	var raw []bencode.RawMessage
	err := bencode.DecodeBytes(b, &raw)
	if err != nil {
		return err
	}

	if len(raw) != 3 {
		return errors.Err("contact must have 3 elements; got %d", len(raw))
	}

	err = bencode.DecodeBytes(raw[0], &c.ID)
	if err != nil {
		return err
	}

	var ipStr string
	err = bencode.DecodeBytes(raw[1], &ipStr)
	if err != nil {
		return err
	}
	c.IP = net.ParseIP(ipStr).To4()
	if c.IP == nil {
		return errors.Err("invalid IP")
	}

	err = bencode.DecodeBytes(raw[2], &c.Port)
	if err != nil {
		return err
	}

	return nil
}

type sortedContact struct {
	contact             Contact
	xorDistanceToTarget bits.Bitmap
}

type byXorDistance []sortedContact

func (a byXorDistance) Len() int      { return len(a) }
func (a byXorDistance) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byXorDistance) Less(i, j int) bool {
	return a[i].xorDistanceToTarget.Cmp(a[j].xorDistanceToTarget) < 0
}