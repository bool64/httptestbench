package httptestbench_test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bool64/httptestbench"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func TestRoundTrip_helloWorld(t *testing.T) {
	res := testing.Benchmark(Benchmark_helloWorld)

	if httptestbench.RaceDetectorEnabled {
		assert.Less(t, res.Extra["B:rcvd/op"], 130.0)
		assert.Less(t, res.Extra["B:sent/op"], 64.0)
		assert.Less(t, res.AllocsPerOp(), int64(27))
		assert.Less(t, res.AllocedBytesPerOp(), int64(6000))
	} else {
		assert.Less(t, res.Extra["B:rcvd/op"], 130.0)
		assert.Less(t, res.Extra["B:sent/op"], 64.0)
		assert.Less(t, res.AllocsPerOp(), int64(18))
		assert.Less(t, res.AllocedBytesPerOp(), int64(1600))
	}
}

func TestRoundTrip_helloWorld_fast(t *testing.T) {
	res := testing.Benchmark(Benchmark_helloWorld_fast)

	if httptestbench.RaceDetectorEnabled {
		assert.Less(t, res.Extra["B:rcvd/op"], 148.0)
		assert.Less(t, res.Extra["B:sent/op"], 64.0)
		assert.Less(t, res.AllocsPerOp(), int64(10))
		assert.Less(t, res.AllocedBytesPerOp(), int64(4000))
	} else {
		assert.Less(t, res.Extra["B:rcvd/op"], 148.0)
		assert.Less(t, res.Extra["B:sent/op"], 64.0)
		assert.Equal(t, res.AllocsPerOp(), int64(1))
		assert.Less(t, res.AllocedBytesPerOp(), int64(40))
	}
}

func Benchmark_helloWorld_fast(b *testing.B) {
	srv := httptestbench.NewServer(func(ctx *fasthttp.RequestCtx) {
		_, err := ctx.Write([]byte("Hello World!"))
		if err != nil {
			b.Fatal(err)
		}
	})

	defer srv.Close()

	httptestbench.RoundTrip(b, 50,
		func(i int, req *fasthttp.Request) {
			req.SetRequestURI(srv.URL)
		},
		func(i int, resp *fasthttp.Response) bool {
			return resp.StatusCode() == http.StatusOK
		},
	)
}

func Benchmark_helloWorld(b *testing.B) {
	srv := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte("Hello World!"))
		if err != nil {
			b.Fatal(err)
		}
	}))
	defer srv.Close()

	httptestbench.RoundTrip(b, 50,
		func(i int, req *fasthttp.Request) {
			req.SetRequestURI(srv.URL)
		},
		func(i int, resp *fasthttp.Response) bool {
			return resp.StatusCode() == http.StatusOK
		},
	)
}

func Benchmark_zeroPort(b *testing.B) {
	listener, err := net.Listen("tcp", ":0") // nolint:gosec
	require.NoError(b, err)

	// Initialize HTTP server.
	srv := http.Server{Handler: http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte("Hello World!"))
		if err != nil {
			b.Fatal(err)
		}
	})}

	go func() {
		require.NoError(b, srv.Serve(listener))
	}()

	defer func() {
		require.NoError(b, srv.Close())
	}()

	baseURL := "http://" + listener.Addr().String() + "/"
	baseURL = strings.Replace(baseURL, "[::]", "127.0.0.1", 1)

	httptestbench.RoundTrip(b, 50,
		func(i int, req *fasthttp.Request) {
			req.SetRequestURI(baseURL)
		},
		func(i int, resp *fasthttp.Response) bool {
			return resp.StatusCode() == http.StatusOK
		},
	)
}

func BenchmarkServeHTTP(b *testing.B) {
	h := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusAccepted)
	})

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(b, err)

	httptestbench.ServeHTTP(b, 50, h, func(i int) *http.Request {
		return req
	}, func(i int, resp *httptest.ResponseRecorder) bool {
		return resp.Code == http.StatusAccepted
	})
}

func TestServeHTTP(t *testing.T) {
	res := testing.Benchmark(BenchmarkServeHTTP)

	if httptestbench.RaceDetectorEnabled {
		assert.LessOrEqual(t, res.AllocsPerOp(), int64(5))
		assert.Less(t, res.AllocedBytesPerOp(), int64(300))
	} else {
		assert.LessOrEqual(t, res.AllocsPerOp(), int64(5))
		assert.Less(t, res.AllocedBytesPerOp(), int64(250))
	}
}
