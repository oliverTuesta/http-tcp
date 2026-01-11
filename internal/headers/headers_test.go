package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {

	t.Run("Valid empty header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("\r\n")

		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, 2, n)
		assert.True(t, done)
	})

	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")

		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 23, n)
		assert.False(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("     Host:    localhost:42069     \r\n\r\n")

		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, 36, n)
		assert.False(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()

		data := []byte("Host: localhost\r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(t, "localhost", headers["host"])
		assert.Equal(t, len(data), n)

		data = []byte("Connection: keep-alive\r\n\r\n")
		n, done, err = headers.Parse(data)

		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(t, "keep-alive", headers["connection"])
		assert.Equal(t, 24, n) // "Connection: keep-alive\r\n"
	})

	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("\r\n")

		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.True(t, done)
		assert.Equal(t, 2, n)
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("       Host : localhost:42069       \r\n\r\n")

		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid character in header key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("H©st: localhost:42069\r\n\r\n")

		n, done, err := headers.Parse(data)

		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Multiple values for same header", func(t *testing.T) {
		headers := NewHeaders()

		data := []byte("Set-Person: lane-loves-go\r\n")
		n, done, err := headers.Parse(data)

		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(t, "lane-loves-go", headers["set-person"])
		assert.Equal(t, len(data), n)

		data = []byte("Set-Person: prime-loves-zig\r\n")
		n, done, err = headers.Parse(data)

		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(
			t,
			"lane-loves-go, prime-loves-zig",
			headers["set-person"],
		)
		assert.Equal(t, len(data), n)

		data = []byte("Set-Person: tj-loves-ocaml\r\n")
		n, done, err = headers.Parse(data)

		require.NoError(t, err)
		assert.False(t, done)
		assert.Equal(
			t,
			"lane-loves-go, prime-loves-zig, tj-loves-ocaml",
			headers["set-person"],
		)
		assert.Equal(t, len(data), n)

		data = []byte("\r\n")
		n, done, err = headers.Parse(data)

		require.NoError(t, err)
		assert.True(t, done)
		assert.Equal(t, 2, n)
	})

}

