package httpstats

import (
	"errors"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/montanaflynn/stats"
)

type (
	statusCountArray   []int64
	statusCountMap     map[int]int64
	percentiledTimeMap map[int]float64
)

const (
	// DefaultBufferSize is buf size sampling response time for percentile and average.
	DefaultBufferSize = 1000
	// DefaultSamplingFactor is factor to sample response time for percentile and average
	DefaultSamplingFactor = 1
)

var (
	httpStatuses = []int{
		http.StatusOK,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
		http.StatusNotImplemented,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout}
	percents       = []int{90, 95, 99}
	rand4HttpStats *rand.Rand
)

// Metrics is a HTTP metrics structure.
// a Metrics measures only one http.Handler.
// Metrics is needed to the number of
// http.Handler which you want to measure.
//
// Metrics has buffer for measuring latest HTTP requests.
// default bufsize is DefaultBufferSize.
// if you want to modify, use NewCapa().
type Metrics struct {
	m           *sync.RWMutex
	count       int64
	statusCount statusCountArray
	requests    stats.Float64Data
	factor      int
	reqIdx      int
}

// RequestData is metrics of HTTP request.
type RequestData struct {
	// Count is sum of all HTTP request counts.
	Count int64 `json:"count"`

	// StatusCount is HTTP request count for each HTTP status.
	StatusCount statusCountMap `json:"status_count"`
}

// ResponseData is metrics of HTTP response.
type ResponseData struct {
	// MaxTime is maximum response time in a period.
	MaxTime float64 `json:"max_time"`

	// MinTime is minimum response time in a period.
	MinTime float64 `json:"min_time"`

	// AverageTime is average HTTP response time of HTTP requests in buffer.
	AverageTime float64 `json:"average_time"`

	// PercentiledTime is HTTP response time percentiles of latest bufsize HTTP requests.
	PercentiledTime percentiledTimeMap `json:"percentiled_time"`
}

// Data is a metrics data.
type Data struct {
	Request  RequestData  `json:"request"`
	Response ResponseData `json:"response"`
}

// statusRecorder is extended http.ResponsWriter.
// statusRecorder enables to access what HTTP status the handler returns.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func init() {
	rand4HttpStats = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// New returns a new statistics structure.
// buffer size of HTTP requests is allocated length of DefaultBufferSize.
func New() (*Metrics, error) {
	return NewCapa(DefaultBufferSize, DefaultSamplingFactor)
}

// NewCapa returns a new statistics structure.
// buffer size of HTTP requests is allocated length of bufsize.
//
// if bufsize is less than 2,
// NewCapa returns error.
func NewCapa(bufsize, factor int) (*Metrics, error) {
	if bufsize < 2 {
		return nil, errors.New("bufsize must be greater than or equal to 2")
	}
	if factor < 1 {
		return nil, errors.New("factor must be greater than 0")
	}

	statusCount := make(statusCountArray, 1000)
	for _, status := range httpStatuses {
		statusCount[status] = 0
	}

	return &Metrics{
		m:           &sync.RWMutex{},
		statusCount: statusCount,
		requests:    make(stats.Float64Data, bufsize),
		factor:      factor,
	}, nil
}

// Data returns statistics Data for http.Handler you set.
func (m *Metrics) Data() (*Data, error) {
	var (
		totalResponseTime float64
		maxTime           float64
		minTime           = math.MaxFloat64
		bufsize           = len(m.requests)
		percentiledTime   = make(percentiledTimeMap)
	)
	m.m.RLock()
	defer m.m.RUnlock()
	for _, v := range m.requests {
		totalResponseTime += v

		if minTime > v {
			minTime = v
		}
		if maxTime < v {
			maxTime = v
		}
	}

	for _, p := range percents {
		var err error
		percentiledTime[p], err = m.requests.Percentile(float64(p))
		if err != nil {
			return nil, err
		}
	}

	statusCount := make(statusCountMap)
	for _, status := range httpStatuses {
		statusCount[status] = atomic.LoadInt64(&m.statusCount[status])
	}

	return &Data{
		Request: RequestData{
			Count:       atomic.LoadInt64(&m.count),
			StatusCount: statusCount,
		},
		Response: ResponseData{
			MaxTime:         maxTime,
			MinTime:         minTime,
			AverageTime:     totalResponseTime / float64(bufsize),
			PercentiledTime: percentiledTime,
		},
	}, nil
}

// WrapHandleFunc wraps http.Handler which you want to measure metrics.
func (m *Metrics) WrapHandleFunc(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		before := time.Now()
		rw := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		h.ServeHTTP(rw, r)
		after := time.Now()
		m.add(rw, after.Sub(before).Seconds())
	})
}

func (m *Metrics) add(r *statusRecorder, t float64) {

	if rand4HttpStats.Intn(m.factor) == 0 {
		m.m.Lock()

		atomic.AddInt64(&m.count, 1)
		atomic.AddInt64(&m.statusCount[r.statusCode], 1)

		m.insertRequestBuffer(t)

		m.m.Unlock()
	} else {
		// Sampling factor was introduced to mitigate heavy lock contention between each goroutine.
		// If sampling factor is greater than 1, RequestData.Count and RequestData.StatusCount are measured in a separated timing.
		atomic.AddInt64(&m.count, 1)
		atomic.AddInt64(&m.statusCount[r.statusCode], 1)
	}
}

func (m *Metrics) insertRequestBuffer(t float64) {
	m.requests[m.reqIdx] = t
	m.reqIdx = (m.reqIdx + 1) % len(m.requests)
}

// WriteHeader is extended http.ResponseWriter's one.
// WriteHeader records HTTP status.
func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
