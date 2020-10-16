package httptestbench_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bool64/httptestbench"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestRoundTrip_helloWorld(t *testing.T) {
	res := testing.Benchmark(Benchmark_helloWorld)

	if httptestbench.RaceDetectorEnabled {
		assert.Less(t, res.Extra["B:rcvd/op"], 130.0)
		assert.Less(t, res.Extra["B:sent/op"], 64.0)
		assert.Less(t, res.AllocsPerOp(), int64(25))
		assert.Less(t, res.AllocedBytesPerOp(), int64(5000))
	} else {
		assert.Less(t, res.Extra["B:rcvd/op"], 130.0)
		assert.Less(t, res.Extra["B:sent/op"], 64.0)
		assert.Less(t, res.AllocsPerOp(), int64(16))
		assert.Less(t, res.AllocedBytesPerOp(), int64(1500))
	}
}

func TestRoundTrip_helloWorld_fast(t *testing.T) {
	res := testing.Benchmark(Benchmark_helloWorld_fast)

	if httptestbench.RaceDetectorEnabled {
		assert.Less(t, res.Extra["B:rcvd/op"], 148.0)
		assert.Less(t, res.Extra["B:sent/op"], 64.0)
		assert.Less(t, res.AllocsPerOp(), int64(10))
		assert.Less(t, res.AllocedBytesPerOp(), int64(2700))
	} else {
		assert.Less(t, res.Extra["B:rcvd/op"], 148.0)
		assert.Less(t, res.Extra["B:sent/op"], 64.0)
		assert.Equal(t, res.AllocsPerOp(), int64(0))
		assert.Less(t, res.AllocedBytesPerOp(), int64(17))
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
