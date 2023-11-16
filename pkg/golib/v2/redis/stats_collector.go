package redis

import (
	"github.com/prometheus/client_golang/prometheus"
)

type dbStatsCollector struct {
	redis *Redis

	maxOpenConnections *prometheus.Desc // maxActive

	openConnections  *prometheus.Desc // activeCount
	inUseConnections *prometheus.Desc // inUseCount
	idleConnections  *prometheus.Desc // idleCount
}

func NewStatsCollector(redis *Redis, dbName string) prometheus.Collector {
	fqName := func(name string) string {
		return "go_redis_" + name
	}
	return &dbStatsCollector{
		redis: redis,
		maxOpenConnections: prometheus.NewDesc(
			fqName("max_open_connections"),
			"Maximum number of open connections to the redis.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		openConnections: prometheus.NewDesc(
			fqName("open_connections"),
			"The number of established connections both in use and idle.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		inUseConnections: prometheus.NewDesc(
			fqName("in_use_connections"),
			"The number of connections currently in use.",
			nil, prometheus.Labels{"db_name": dbName},
		),
		idleConnections: prometheus.NewDesc(
			fqName("idle_connections"),
			"The number of idle connections.",
			nil, prometheus.Labels{"db_name": dbName},
		),
	}
}

// Describe implements Collector.
func (c *dbStatsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.maxOpenConnections
	ch <- c.openConnections
	ch <- c.inUseConnections
	ch <- c.idleConnections
}

// Collect implements Collector.
func (c *dbStatsCollector) Collect(ch chan<- prometheus.Metric) {
	inUseCount, idleCount, activeCount := c.redis.Stats()
	ch <- prometheus.MustNewConstMetric(c.maxOpenConnections, prometheus.GaugeValue, float64(c.redis.pool.MaxActive))
	ch <- prometheus.MustNewConstMetric(c.openConnections, prometheus.GaugeValue, float64(activeCount))
	ch <- prometheus.MustNewConstMetric(c.inUseConnections, prometheus.GaugeValue, float64(inUseCount))
	ch <- prometheus.MustNewConstMetric(c.idleConnections, prometheus.GaugeValue, float64(idleCount))
}
