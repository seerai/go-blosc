package blosc_test

import (
	"testing"
	"unsafe"

	blosc "github.com/seerai/go-blosc"
	"github.com/stretchr/testify/require"
)

func TestRoundTrip(t *testing.T) {

	buf := make([]int64, 10000)
	for i := 0; i < len(buf); i++ {
		buf[i] = int64(112)
	}

	err := blosc.SetCompressor("lz4")
	require.NoError(t, err)

	cmp, err := blosc.Compress(1, true, buf)
	require.NoError(t, err)
	dec := blosc.Decompress(cmp)
	require.Equal(t, len(buf)*int(unsafe.Sizeof(buf[0])), len(dec))

	obuf := (*(*[1 << 32]int64)(unsafe.Pointer(&dec[0])))[:len(dec)/int(unsafe.Sizeof(buf[0]))]
	for i, o := range obuf {
		require.Equal(t, buf[i], o)
	}

}

func TestEmpty(t *testing.T) {
	var b []byte
	err := blosc.SetCompressor("lz4")
	require.NoError(t, err)

	cmp, err := blosc.Compress(1, true, b)
	require.NoError(t, err)
	dec := blosc.Decompress(cmp)
	require.Equal(t, 0, len(dec))
}
