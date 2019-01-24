go-httpstats
============

Usage:
------

`go-httpstats` is a Golang package to report HTTP statistics for each `http.Handler`.
the library reports

- HTTP request count
- HTTP response time
- HTTP request count for each HTTP status
- HTTP response time for each HTTP status
- 90, 95, 99 percentiles of HTTP response time
- average HTTP  response time

Example
------

\_example/main.go

```go:_example/main.go
func main() {
	mw := stats.New()
	handler := mw.WrapHandleFunc(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("hello world\n"))
		}))
	statsHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, err := mw.Data()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		if err := json.NewEncoder(w).Encode(d); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	})
	http.Handle("/", handler)
	http.Handle("/stats", statsHandler)
	if err := http.ListenAndServe(":9999", nil); err != nil {
		fmt.Println(err)
	}

}

```

request to \_example/main.go.  
```bash
$ curl -s http://localhost:9999/stats | jq .
{
  "request": {
    "count": 641344,
    "status_count": {
      "200": 641344,
      "400": 0,
      "401": 0,
      "403": 0,
      "404": 0,
      "500": 0,
      "501": 0,
      "502": 0,
      "503": 0,
      "504": 0
    }
  },
  "response": {
    "max_time": 1.4884e-05,
    "min_time": 4.36e-07,
    "average_time": 7.35283e-07,
    "percentiled_time": {
      "90": 8.02e-07,
      "95": 8.94e-07,
      "99": 1.225e-06
    }
  }
}
```

License:
--------
MIT License

Author:
-------
Copyright Mercari, Inc.
