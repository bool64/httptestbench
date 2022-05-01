# HTTP test benchmark helper for Go

[![Build Status](https://github.com/bool64/httptestbench/workflows/test/badge.svg)](https://github.com/bool64/httptestbench/actions?query=branch%3Amaster+workflow%3Atest)
[![Coverage Status](https://codecov.io/gh/bool64/httptestbench/branch/master/graph/badge.svg)](https://codecov.io/gh/bool64/httptestbench)
[![GoDevDoc](https://img.shields.io/badge/dev-doc-00ADD8?logo=go)](https://pkg.go.dev/github.com/bool64/httptestbench)
![Code lines](https://sloc.xyz/github/bool64/httptestbench/?category=code)
![Comments](https://sloc.xyz/github/bool64/httptestbench/?category=comments)

This module provides benchmark helper to send concurrent http requests and assert responses.

There are a few custom metrics added to benchmark result:
* `50%:ms`, `90%:ms`, `99%:ms`, `99.9%:ms` are percentile values of round trip latency in milliseconds,
* `B:rcvd/op` and `B:sent/op` are average bytes received and sent per request,
* `rps` is requests per second rate.

The client is using [`github.com/valyala/fasthttp`](https://github.com/valyala/fasthttp) for better performance 
and lower impact on benchmark results.

## Example

```go
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
```

Sample result:

```
Benchmark_helloWorld-12    	  142648	      8230 ns/op	         0.3469 50%:ms	         0.6917 90%:ms	         1.428 99%:ms	         2.721 99.9%:ms	       129.0 B:rcvd/op	        63.00 B:sent/op	    121496 rps	    1551 B/op	      17 allocs/op
```

Performance expectations can also be locked with a test: 
```go
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
```