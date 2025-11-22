package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := flag.String("url", "", "URL")
	reqs := flag.Int("requests", 0, "Total de requests")
	conc := flag.Int("concurrency", 0, "Concorrência")
	token := flag.String("token", "", "API Token (opcional)")
	flag.Parse()

	if *url == "" || *reqs <= 0 || *conc <= 0 {
		fmt.Println("Use: --url --requests --concurrency [--token]")
		return
	}

	var total, success int64
	var status sync.Map

	client := &http.Client{Timeout: 5 * time.Second}
	tasks := make(chan struct{}, *reqs)

	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < *conc; i++ {
		wg.Go(func() {
			for range tasks {
				req, err := http.NewRequest("GET", *url, nil)
				if err != nil {
					atomic.AddInt64(&total, 1)
					continue
				}

				if *token != "" {
					req.Header.Set("API_KEY", *token)
				}

				resp, err := client.Do(req)
				current := atomic.AddInt64(&total, 1)

				fmt.Printf("\rProgress: %d/%d (%.0f%%)",
					current, *reqs, float64(current)/float64(*reqs)*100)

				code := -1
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					code = resp.StatusCode

					if code == 200 {
						atomic.AddInt64(&success, 1)
					}
				}

				val, _ := status.LoadOrStore(code, new(int64))
				atomic.AddInt64(val.(*int64), 1)
			}
		})
	}

	for i := 0; i < *reqs; i++ {
		tasks <- struct{}{}
	}
	close(tasks)

	wg.Wait()
	elapsed := time.Since(start)

	fmt.Println("\n\nRESULT:")
	fmt.Printf("Total time: %v\n", elapsed)
	fmt.Printf("Total requests: %d\n", total)
	fmt.Printf("Status 200: %d (%.2f%%)\n", success, float64(success)/float64(total)*100)

	fmt.Println("\nHTTP CODES:")

	status.Range(func(k, v any) bool {
		code := k.(int)
		count := atomic.LoadInt64(v.(*int64))

		label := getStatusLabel(code)
		fmt.Printf("  %s: %d (%.2f%%)\n", label, count, float64(count)/float64(total)*100)

		return true
	})

	fmt.Printf("\nRPS: %.2f req/s\n", float64(total)/elapsed.Seconds())
}

func getStatusLabel(code int) string {
	switch code {
	case 200:
		return "200 OK"
	case 201:
		return "201 Created"
	case 204:
		return "204 No Content"
	case 300:
		return "300 Multiple Choices"
	case 301:
		return "301 Moved Permanently"
	case 302:
		return "302 Found"
	case 304:
		return "304 Not Modified"
	case 400:
		return "400 Bad Request"
	case 401:
		return "401 Unauthorized"
	case 403:
		return "403 Forbidden"
	case 404:
		return "404 Not Found"
	case 429:
		return "429 Too Many Requests"
	case 500:
		return "500 Internal Server Error"
	case 502:
		return "502 Bad Gateway"
	case 503:
		return "503 Service Unavailable"
	case -1:
		return "Erro de Timeout"
	default:
		return fmt.Sprintf("%d Unknown", code)
	}
}
