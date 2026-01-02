package blosc_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	blosc "github.com/seerai/go-blosc"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip(t *testing.T) {

	ints := make([]int64, 10000)
	for i := 0; i < len(ints); i++ {
		ints[i] = int64(112)
	}

	// convert to byte slice
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, ints)
	require.NoError(t, err)

	c, err := blosc.NewContext(
		blosc.WithCompressor("lz4"),
		blosc.WithCompressionLevel(9),
		blosc.WithShuffleString("bitshuffle"),
		blosc.WithShuffle(blosc.BitShuffle),
		blosc.WithBlockSize(100),
		blosc.WithTypeSize(8),
		blosc.WithNThreads(2),
	)
	require.NoError(t, err)

	cmp, err := c.Compress(buf.Bytes())
	require.NoError(t, err)
	dec, err := c.Decompress(cmp)
	require.NoError(t, err)
	require.Equal(t, len(buf.Bytes()), len(dec))

	for _, cName := range blosc.ValidCompressors() {
		c, err := blosc.NewContext(
			blosc.WithCompressor(cName),
		)
		require.NoError(t, err)

		cmp, err := c.Compress(buf.Bytes())
		require.NoError(t, err)
		dec, err := c.Decompress(cmp)
		require.NoError(t, err)
		require.Equal(t, len(buf.Bytes()), len(dec))
	}

	// for each shuffle type
	for _, shuffle := range blosc.ValidShuffles() {
		c, err := blosc.NewContext(
			blosc.WithShuffle(shuffle),
		)
		require.NoError(t, err)

		cmp, err := c.Compress(buf.Bytes())
		require.NoError(t, err)
		dec, err := c.Decompress(cmp)
		require.NoError(t, err)
		require.Equal(t, len(buf.Bytes()), len(dec))
	}

	// Invalid shuffle string
	_, err = blosc.NewContext(
		blosc.WithShuffleString("invalidshuffle"),
	)
	require.Error(t, err)

	// Invalid shuffle type
	_, err = blosc.NewContext(
		blosc.WithShuffle(blosc.ShuffleType(100)),
	)
	require.Error(t, err)

	// Invalid compressor
	_, err = blosc.NewContext(
		blosc.WithCompressor("invalidcompressor"),
	)
	require.Error(t, err)
}

func TestShuffleTypes(t *testing.T) {
	shuffles := blosc.ValidShuffles()
	require.Equal(t, 3, len(shuffles))
	require.Contains(t, shuffles, blosc.NoShuffle)
	require.Contains(t, shuffles, blosc.ByteShuffle)
	require.Contains(t, shuffles, blosc.BitShuffle)

	// test string conversion
	require.Equal(t, blosc.NoShuffleStr, blosc.NoShuffle.String())
	require.Equal(t, blosc.ByteShuffleStr, blosc.ByteShuffle.String())
	require.Equal(t, blosc.BitShuffleStr, blosc.BitShuffle.String())

	// test FromString
	var s blosc.ShuffleType
	s.FromString("nosHUffle")
	require.Equal(t, blosc.NoShuffle, s)
	s.FromString("SHUffle")
	require.Equal(t, blosc.ByteShuffle, s)
	s.FromString("bitSHUffle")
	require.Equal(t, blosc.BitShuffle, s)

	s = blosc.ShuffleType(101)
	require.Equal(t, "invalid", s.String())
}

func TestRoundTripDefaults(t *testing.T) {

	ints := make([]int64, 10000)
	for i := 0; i < len(ints); i++ {
		ints[i] = int64(112)
	}

	// convert to byte slice
	buf := &bytes.Buffer{}
	err := binary.Write(buf, binary.LittleEndian, ints)
	require.NoError(t, err)

	c, err := blosc.NewContext()
	require.NoError(t, err)

	cmp, err := c.Compress(buf.Bytes())
	require.NoError(t, err)
	dec, err := c.Decompress(cmp)
	require.NoError(t, err)
	require.Equal(t, len(buf.Bytes()), len(dec))
}

func TestEmpty(t *testing.T) {
	var b []byte
	c, err := blosc.NewContext()
	require.NoError(t, err)
	cmp, err := c.Compress(b)
	require.NoError(t, err)
	dec, err := c.Decompress(cmp)
	require.NoError(t, err)
	require.Equal(t, 0, len(dec))
}
