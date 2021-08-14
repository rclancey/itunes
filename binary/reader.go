package binary

import (
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"io"

	"github.com/pkg/errors"
)

type ecb struct {
	block cipher.Block
	blockSize int
}

func NewECBDecrypter(b cipher.Block) cipher.BlockMode {
	return &ecb{block: b, blockSize: b.BlockSize()}
}

func (x *ecb) BlockSize() int { return x.blockSize }

func (x *ecb) CryptBlocks(dst, src []byte) {
	if len(src) % x.blockSize != 0 {
		panic("crypto/cipher: input not full blocks")
	}
	if len(dst) < len(src) {
		panic("crypto/cipher: output smaller than input")
	}
	for len(src) > 0 {
		x.block.Decrypt(dst, src[:x.blockSize])
		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

type DecryptReader struct {
	rawReader io.ReadCloser
	mode cipher.BlockMode
	buffer []byte
	rawPosition int
	encryptedSize int
}

func NewDecryptReader(r io.ReadCloser, encryptedSize int) (*DecryptReader, error) {
	cipher, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	mode := NewECBDecrypter(cipher)
	return &DecryptReader{
		rawReader: r,
		mode: mode,
		buffer: []byte{},
		rawPosition: 0,
		encryptedSize: encryptedSize,
	}, nil
}

func (r *DecryptReader) Read(dst []byte) (int, error) {
	if r.rawReader == nil {
		return 0, errors.New("file already closed")
	}
	if len(r.buffer) > len(dst) {
		copy(dst, r.buffer)
		r.buffer = r.buffer[len(dst):]
		return len(dst), nil
	}
	offset := 0
	if len(r.buffer) > 0 {
		offset += len(r.buffer)
		copy(dst, r.buffer)
		r.buffer = []byte{}
	}
	if r.encryptedSize >= 0 && r.rawPosition >= r.encryptedSize {
		bytesRead, err := r.rawReader.Read(dst[offset:])
		r.rawPosition += bytesRead
		return bytesRead + offset, errors.WithStack(err)
	}
	size := len(dst) - offset
	overhang := size % r.mode.BlockSize()
	if overhang > 0 {
		size += r.mode.BlockSize() - overhang
		if r.encryptedSize >= 0 && r.rawPosition + size > r.encryptedSize {
			size = r.encryptedSize - r.rawPosition
		}
	}
	buf := make([]byte, size)
	bytesRead, err := r.rawReader.Read(buf)
	r.rawPosition += bytesRead
	if err != nil {
		return offset + bytesRead, errors.WithStack(err)
	}
	r.mode.CryptBlocks(buf, buf)
	copy(dst[offset:], buf)
	if len(buf) + offset < len(dst) {
		r.buffer = []byte{}
		return len(buf) + offset, nil
	}
	r.buffer = buf[len(dst) - offset:]
	return len(dst), nil
}

func (r *DecryptReader) Close() error {
	if r.rawReader == nil {
		return nil
	}
	r.buffer = nil
	r.rawReader.Close()
	r.rawReader = nil
	return nil
}

type PayloadReader struct {
	cryptReader io.ReadCloser
	zlibReader io.ReadCloser
}

func NewPayloadReader(r io.ReadCloser, encryptedSize int) (*PayloadReader, error) {
	cr, err := NewDecryptReader(r, encryptedSize)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if encryptedSize == -2 {
		return &PayloadReader{cr, nil}, nil
	}
	zr, err := zlib.NewReader(cr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &PayloadReader{cr, zr}, nil
}

func (r *PayloadReader) Read(dst []byte) (int, error) {
	if r.zlibReader == nil {
		return r.cryptReader.Read(dst)
	}
	return r.zlibReader.Read(dst)
}

func (r *PayloadReader) Close() error {
	if r.zlibReader != nil {
		err := r.zlibReader.Close()
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return r.cryptReader.Close()
}
