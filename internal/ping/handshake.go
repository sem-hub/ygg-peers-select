package pinger

import (
	"crypto/ed25519"
	"encoding/binary"
	"fmt"
	"io"

	"golang.org/x/crypto/blake2b"
)

const (
	protocolMajor = 0
	protocolMinor = 5
)

func metadataHash(key, password []byte) ([]byte, error) {
	h, err := blake2b.New512(password)
	if err != nil {
		return nil, fmt.Errorf("initialize handshake hash: %w", err)
	}
	_, _ = h.Write(key)
	return h.Sum(nil), nil
}

func encodeMetadata(publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey, password []byte, priority byte) ([]byte, error) {
	msg := []byte{'m', 'e', 't', 'a', 0, 0}
	appendField := func(id uint16, value []byte) {
		msg = binary.BigEndian.AppendUint16(msg, id)
		msg = binary.BigEndian.AppendUint16(msg, uint16(len(value)))
		msg = append(msg, value...)
	}
	appendField(0, []byte{0, protocolMajor})
	appendField(1, []byte{0, protocolMinor})
	appendField(2, publicKey)
	appendField(3, []byte{priority})
	hash, err := metadataHash(publicKey, password)
	if err != nil {
		return nil, err
	}
	msg = append(msg, ed25519.Sign(privateKey, hash)...)
	binary.BigEndian.PutUint16(msg[4:6], uint16(len(msg)-6))
	return msg, nil
}

func exchangeMetadata(conn io.ReadWriter, metadata, password []byte) error {
	n, err := conn.Write(metadata)
	if err != nil {
		return fmt.Errorf("write metadata: %w", err)
	}
	if n != len(metadata) {
		return io.ErrShortWrite
	}
	header := make([]byte, 6)
	if _, err := io.ReadFull(conn, header); err != nil {
		return fmt.Errorf("read metadata header: %w", err)
	}
	if string(header[:4]) != "meta" {
		return fmt.Errorf("invalid metadata preamble")
	}
	length := int(binary.BigEndian.Uint16(header[4:]))
	if length < ed25519.SignatureSize {
		return fmt.Errorf("invalid metadata length")
	}
	body := make([]byte, length)
	if _, err := io.ReadFull(conn, body); err != nil {
		return fmt.Errorf("read metadata: %w", err)
	}
	fields, signature := body[:length-ed25519.SignatureSize], body[length-ed25519.SignatureSize:]
	var major, minor uint16
	var key ed25519.PublicKey
	seen := [4]bool{}
	for len(fields) > 0 {
		if len(fields) < 4 {
			return fmt.Errorf("invalid metadata field header")
		}
		id, size := binary.BigEndian.Uint16(fields), int(binary.BigEndian.Uint16(fields[2:]))
		fields = fields[4:]
		if size > len(fields) {
			return fmt.Errorf("invalid metadata field length")
		}
		value := fields[:size]
		fields = fields[size:]
		if id <= 3 {
			if seen[id] {
				return fmt.Errorf("duplicate metadata field %d", id)
			}
			seen[id] = true
		}
		switch id {
		case 0, 1:
			if len(value) != 2 {
				return fmt.Errorf("invalid protocol version field")
			}
			if id == 0 {
				major = binary.BigEndian.Uint16(value)
			} else {
				minor = binary.BigEndian.Uint16(value)
			}
		case 2:
			if len(value) != ed25519.PublicKeySize {
				return fmt.Errorf("invalid peer public key")
			}
			key = value
		case 3:
			if len(value) != 1 {
				return fmt.Errorf("invalid priority field")
			}
		}
	}
	if !seen[0] || !seen[1] || !seen[2] || major != protocolMajor || minor != protocolMinor {
		return fmt.Errorf("incompatible Yggdrasil metadata version %d.%d", major, minor)
	}
	hash, err := metadataHash(key, password)
	if err != nil {
		return err
	}
	if !ed25519.Verify(key, hash, signature) {
		return fmt.Errorf("invalid metadata signature or password")
	}
	return nil
}
