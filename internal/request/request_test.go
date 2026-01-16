package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n
	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		chunk  int
		assert func(t *testing.T, r *Request, err error)
	}{
		{
			name:  "Good GET Request line",
			data:  "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n",
			chunk: 3,
			assert: func(t *testing.T, r *Request, err error) {
				require.NoError(t, err)
				assert.Equal(t, "GET", r.RequestLine.Method)
				assert.Equal(t, "/", r.RequestLine.RequestTarget)
				assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
			},
		},
		{
			name:  "Good GET Request line with path",
			data:  "GET /coffee HTTP/1.1\r\nHost: localhost\r\n\r\n",
			chunk: 10,
			assert: func(t *testing.T, r *Request, err error) {
				require.NoError(t, err)
				assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
			},
		},
		{
			name:  "Bad method",
			data:  "GIT /coffee HTTP/1.1\r\n",
			chunk: 2,
			assert: func(t *testing.T, r *Request, err error) {
				require.Error(t, err)
			},
		},
		{
			name:  "Unsupported HTTP version",
			data:  "POST /coffee HTTP/1.2\r\nHost: localhost\r\n\r\n",
			chunk: 5,
			assert: func(t *testing.T, r *Request, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &chunkReader{
				data:            tt.data,
				numBytesPerRead: tt.chunk,
			}
			r, err := RequestFromReader(reader)
			tt.assert(t, r, err)
		})
	}
}

func TestRequestHeadersParse(t *testing.T) {
	tests := []struct {
		name   string
		data   string
		chunk  int
		assert func(t *testing.T, r *Request, err error)
	}{
		{
			name:  "Standard Headers",
			data:  "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			chunk: 3,
			assert: func(t *testing.T, r *Request, err error) {
				require.NoError(t, err)
				assert.Equal(t, "localhost:42069", r.Headers["host"])
				assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
				assert.Equal(t, "*/*", r.Headers["accept"])
			},
		},
		{
			name:  "Empty Headers",
			data:  "GET / HTTP/1.1\r\n\r\n",
			chunk: 2,
			assert: func(t *testing.T, r *Request, err error) {
				require.NoError(t, err)
				assert.Len(t, r.Headers, 0)
			},
		},
		{
			name:  "Malformed Header",
			data:  "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
			chunk: 3,
			assert: func(t *testing.T, r *Request, err error) {
				require.Error(t, err)
			},
		},
		{
			name:  "Duplicate Headers",
			data:  "GET / HTTP/1.1\r\nHost: a\r\nHost: b\r\n\r\n",
			chunk: 3,
			assert: func(t *testing.T, r *Request, err error) {
				require.NoError(t, err)
				assert.Equal(t, "a, b", r.Headers["host"])
			},
		},
		{
			name:  "Case Insensitive Headers",
			data:  "GET / HTTP/1.1\r\nHOST: localhost\r\nUsEr-AgEnT: curl\r\n\r\n",
			chunk: 3,
			assert: func(t *testing.T, r *Request, err error) {
				require.NoError(t, err)
				assert.Equal(t, "localhost", r.Headers["host"])
				assert.Equal(t, "curl", r.Headers["user-agent"])
			},
		},
		{
			name:  "Missing End of Headers",
			data:  "GET / HTTP/1.1\r\nHost: localhost",
			chunk: 3,
			assert: func(t *testing.T, r *Request, err error) {
				require.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := &chunkReader{
				data:            tt.data,
				numBytesPerRead: tt.chunk,
			}
			r, err := RequestFromReader(reader)
			tt.assert(t, r, err)
		})
	}
}

func TestRequestBodyParsing(t *testing.T) {
	// Test: Standard Body (valid)
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Empty Body, 0 reported content length (valid)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 0\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Empty(t, r.Body)

	// Test: Empty Body, no reported content length (valid)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Empty(t, r.Body)

	// Test: Body shorter than reported content length (should error)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: No Content-Length but Body Exists (should not error)
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"hello world!",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
}
