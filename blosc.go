// Package blosc wraps blosc for compressing numbers.
package blosc

/*
#cgo LDFLAGS: -lpthread -lblosc
#include "blosc.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"unsafe"
)

type ShuffleType uint

const (
	NoShuffle         ShuffleType = C.BLOSC_NOSHUFFLE
	ByteShuffle       ShuffleType = C.BLOSC_SHUFFLE
	BitShuffle        ShuffleType = C.BLOSC_BITSHUFFLE
	BitShuffleInvalid ShuffleType = 100
	// Zarr Version 3 shuffle names
	NoShuffleStr   = "noshuffle"
	ByteShuffleStr = "shuffle"
	BitShuffleStr  = "bitshuffle"
	InvalidStr     = "invalid"
)

type CNameType string

const (
	BloscLZ CNameType = "blosclz"
	LZ4     CNameType = "lz4"
	LZ4HC   CNameType = "lz4hc"
	ZLIB    CNameType = "zlib"
	ZSTD    CNameType = "zstd"
)

func ValidCompressors() []CNameType {
	return []CNameType{
		BloscLZ,
		LZ4,
		LZ4HC,
		ZLIB,
		ZSTD,
	}
}

func ValidShuffles() []ShuffleType {
	return []ShuffleType{
		NoShuffle,
		ByteShuffle,
		BitShuffle,
	}
}

func (s ShuffleType) String() string {
	switch s {
	case NoShuffle:
		return NoShuffleStr
	case ByteShuffle:
		return ByteShuffleStr
	case BitShuffle:
		return BitShuffleStr
	default:
		return InvalidStr
	}
}

func (s *ShuffleType) FromString(str string) {
	str = strings.ToLower(str)
	switch str {
	case "noshuffle":
		*s = NoShuffle
	case "shuffle":
		*s = ByteShuffle
	case "bitshuffle":
		*s = BitShuffle
	default:
		*s = BitShuffleInvalid
	}
}

var nThreads = runtime.NumCPU()

type Context struct {
	CName     CNameType
	CLevel    uint
	Shuffle   ShuffleType
	BlockSize uint
	TypeSize  uint
	nThreads  uint
}

type ContextOption func(*Context) error

// NewContext creates a new blosc context with the given parameters.
func WithCompressor(name CNameType) ContextOption {
	return func(c *Context) error {
		if !slices.Contains(ValidCompressors(), name) {
			validCompressors := make([]string, len(ValidCompressors()))
			for i, v := range ValidCompressors() {
				validCompressors[i] = string(v)
			}
			return fmt.Errorf("invalid compressor name, valid names are: %s", strings.Join(validCompressors, ", "))
		}
		c.CName = name

		return nil
	}
}

func WithCompressionLevel(level uint) ContextOption {
	return func(c *Context) error {
		c.CLevel = level
		return nil
	}
}

func WithShuffle(shuffle ShuffleType) ContextOption {
	return func(c *Context) error {
		if shuffle != NoShuffle && shuffle != ByteShuffle && shuffle != BitShuffle {
			return errors.New("invalid shuffle type")
		}
		c.Shuffle = shuffle
		return nil
	}
}

func WithShuffleString(shuffle string) ContextOption {
	return func(c *Context) error {
		var s ShuffleType
		s.FromString(shuffle)
		if s > BitShuffle {
			return fmt.Errorf("invalid shuffle type string, must be '%s', '%s', or '%s'", NoShuffleStr, ByteShuffleStr, BitShuffleStr)
		}
		c.Shuffle = s
		return nil
	}
}

func WithBlockSize(blocksize uint) ContextOption {
	return func(c *Context) error {
		c.BlockSize = blocksize
		return nil
	}
}

func WithTypeSize(typesize uint) ContextOption {
	return func(c *Context) error {
		c.TypeSize = typesize
		return nil
	}
}

func WithNThreads(threads uint) ContextOption {
	return func(c *Context) error {
		c.nThreads = threads
		return nil
	}
}

func NewContext(opts ...ContextOption) (*Context, error) {
	c := &Context{
		CName:     LZ4,
		CLevel:    5,
		Shuffle:   NoShuffle,
		BlockSize: 0,
		TypeSize:  1,
		nThreads:  uint(nThreads),
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Compress takes a slice of numbers and compresses according to level and shuffle.
func (c *Context) Compress(slice []byte) ([]byte, error) {
	if len(slice) == 0 {
		return []byte{}, nil
	}

	sliceLen := len(slice)
	size := 1
	if c.TypeSize > 0 {
		size = int(c.TypeSize)
	}
	ptr := unsafe.Pointer(&slice[0])

	level := c.CLevel

	compressed := make([]byte, sliceLen*size+C.BLOSC_MAX_OVERHEAD)

	// c str for cname
	cname := C.CString(string(c.CName))
	defer C.free(unsafe.Pointer(cname))

	csize := C.blosc_compress_ctx(C.int(level), C.int(c.Shuffle), C.size_t(size),
		C.size_t(sliceLen),
		ptr,
		unsafe.Pointer(&compressed[0]),
		C.size_t(len(compressed)),
		cname,
		C.size_t(c.BlockSize),
		C.int(nThreads),
	)
	if csize < 0 {
		return nil, fmt.Errorf("blosc compression error while using compressor %s: %d", c.CName, int(csize))
	}

	return compressed[:csize], nil
}

// Decompress takes a byte of compressed data and returns the uncompressed data.
func (c *Context) Decompress(compressed []byte) []byte {
	if len(compressed) == 0 {
		return []byte{}
	}

	nbytes := C.size_t(0)
	cbytes := C.size_t(0)
	blksz := C.size_t(0)

	C.blosc_cbuffer_sizes(unsafe.Pointer(&compressed[0]), &nbytes, &cbytes, &blksz)

	data := make([]byte, int(nbytes))
	C.blosc_decompress_ctx(unsafe.Pointer(&compressed[0]), unsafe.Pointer(&data[0]), nbytes, C.int(nThreads))
	return data
}
