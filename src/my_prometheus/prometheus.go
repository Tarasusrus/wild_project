package my_prometheus

import "github.com/prometheus/client_golang/prometheus"

// Инициализация метрик Prometheus
var (
	TotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Количество GET запросов.",
		},
		[]string{"path"},
	)
	CacheResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_cache_response_time_seconds",
			Help: "Гистограмма времени ответа кэша.",
		},
		[]string{"path"},
	)
	DbResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_db_response_time_seconds",
			Help: "Гистограмма времени ответа базы данных.",
		},
		[]string{"path"},
	)
	OverallResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_overall_response_time_seconds",
			Help: "Гистограмма общего времени ответа обработчика.",
		},
		[]string{"path"},
	)
)

func init() {
	// Регистрация метрик в Prometheus
	prometheus.MustRegister(TotalRequests)
	prometheus.MustRegister(CacheResponseTime)
	prometheus.MustRegister(DbResponseTime)
	prometheus.MustRegister(OverallResponseTime)
}
