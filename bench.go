package httptestbench

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/valyala/fasthttp"
)

type countingDialer struct {
	dial fasthttp.DialFunc
	sent uint64
	rcvd uint64

	mu    sync.Mutex
	conns []net.Conn
}

type countingConn struct {
	net.Conn
	dialer *countingDialer
}

func (c countingConn) Write(data []byte) (int, error) {
	n, err := c.Conn.Write(data)
	atomic.AddUint64(&c.dialer.sent, uint64(n))

	return n, err
}

func (c countingConn) Read(data []byte) (int, error) {
	n, err := c.Conn.Read(data)
	atomic.AddUint64(&c.dialer.rcvd, uint64(n))

	return n, err
}

func (c *countingDialer) Dial(addr string) (net.Conn, error) {
	conn, err := c.dial(addr)

	if err == nil {
		c.mu.Lock()
		c.conns = append(c.conns, conn)
		c.mu.Unlock()
	}

	return countingConn{Conn: conn, dialer: c}, err
}

func (c *countingDialer) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, conn := range c.conns {
		err := conn.Close()
		if err != nil {
			log.Printf("failed to close connection: %v", err)
		}
	}

	c.conns = nil
}

var (
	dialer = countingDialer{
		dial: (&fasthttp.TCPDialer{Concurrency: 1000}).Dial,
	}

	client = fasthttp.Client{
		Dial:                          dialer.Dial,
		DisableHeaderNamesNormalizing: true,
		DisablePathNormalizing:        true,
		MaxIdleConnDuration:           time.Second / 10,
	}
)

// RoundTrip sends b.N requests concurrently and asserts valid response.
func RoundTrip(
	b *testing.B,
	concurrency int,
	setupRequest func(i int, req *fasthttp.Request),
	responseIsValid func(i int, resp *fasthttp.Response) bool,
) {
	b.Helper()

	atomic.StoreUint64(&dialer.rcvd, 0)
	atomic.StoreUint64(&dialer.sent, 0)

	concurrentBench(b, concurrency, func(i int) {
		req := fasthttp.AcquireRequest()
		resp := fasthttp.AcquireResponse()
		defer func() {
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
		}()

		setupRequest(i, req)

		err := client.Do(req, resp)
		if err != nil {
			b.Fatal(err.Error())
		}

		if !responseIsValid(i, resp) {
			failIteration(i, resp.StatusCode(), resp.Body())
		}
	})

	b.ReportMetric(float64(atomic.LoadUint64(&dialer.sent))/float64(b.N), "B:sent/op")
	b.ReportMetric(float64(atomic.LoadUint64(&dialer.rcvd))/float64(b.N), "B:rcvd/op")

	client.CloseIdleConnections()
}

func concurrentBench(b *testing.B, concurrency int, iterate func(i int)) {
	semaphore := make(chan bool, concurrency)
	start := time.Now()

	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		i := i
		semaphore <- true

		go func() {
			defer func() {
				<-semaphore
			}()

			iterate(i)
		}()
	}

	for i := 0; i < cap(semaphore); i++ {
		semaphore <- true
	}

	b.StopTimer()

	elapsed := time.Since(start)

	b.ReportMetric(float64(b.N)/elapsed.Seconds(), "rps")
}

func failIteration(i int, code int, body []byte) {
	panic(fmt.Sprintf("iteration: %d, unexpected result status: %d, body: %q",
		i, code, string(body)))
}

// ServeHTTP serves b.N requests with http.Handler concurrently and asserts valid response.
func ServeHTTP(
	b *testing.B,
	concurrency int,
	h http.Handler,
	makeRequest func(i int) *http.Request,
	responseIsValid func(i int, resp *httptest.ResponseRecorder) bool,
) {
	b.Helper()

	concurrentBench(b, concurrency, func(i int) {
		req := makeRequest(i)
		rw := httptest.NewRecorder()

		h.ServeHTTP(rw, req)

		if !responseIsValid(i, rw) {
			failIteration(i, rw.Code, rw.Body.Bytes())
		}
	})
}
