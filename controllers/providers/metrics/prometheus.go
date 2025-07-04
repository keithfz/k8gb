package metrics

/*
Copyright 2021-2025 The k8gb Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Generated by GoLic, for more details see: https://github.com/AbsaOSS/golic
*/

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/k8gb-io/k8gb/controllers/resolver"

	"github.com/k8gb-io/k8gb/controllers/utils"

	externaldns "sigs.k8s.io/external-dns/endpoint"

	k8gbv1beta1 "github.com/k8gb-io/k8gb/api/v1beta1"
	"github.com/prometheus/client_golang/prometheus"
	crm "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	primary   = "primary"
	secondary = "secondary"
)

const (
	K8gbGslbErrorsTotal               = "k8gb_gslb_errors_total"
	K8gbGslbHealthyRecords            = "k8gb_gslb_healthy_records"
	K8gbGslbReconciliationLoopsTotal  = "k8gb_gslb_reconciliation_loops_total"
	K8gbGslbServiceStatusNum          = "k8gb_gslb_service_status_num"
	K8gbGslbStatusCountForFailover    = "k8gb_gslb_status_count_for_failover"
	K8gbGslbStatusCountForRoundrobin  = "k8gb_gslb_status_count_for_roundrobin"
	K8gbGslbStatusCountForGeoIP       = "k8gb_gslb_status_count_for_geoip"
	K8gbInfobloxHeartbeatsTotal       = "k8gb_infoblox_heartbeats_total"
	K8gbInfobloxHeartbeatErrorsTotal  = "k8gb_infoblox_heartbeat_errors_total"
	K8gbInfobloxRequestDuration       = "k8gb_infoblox_request_duration"
	K8gbInfobloxZoneUpdatesTotal      = "k8gb_infoblox_zone_updates_total"
	K8gbInfobloxZoneUpdateErrorsTotal = "k8gb_infoblox_zone_update_errors_total"
	K8gbEndpointStatusNum             = "k8gb_endpoint_status_num"
	K8gbRuntimeInfo                   = "k8gb_runtime_info"
)

// collectors contains list of metrics.
type collectors struct {
	K8gbGslbHealthyRecords            *prometheus.GaugeVec
	K8gbGslbServiceStatusNum          *prometheus.GaugeVec
	K8gbGslbStatusCountForFailover    *prometheus.GaugeVec
	K8gbGslbStatusCountForRoundrobin  *prometheus.GaugeVec
	K8gbGslbStatusCountForGeoip       *prometheus.GaugeVec
	K8gbGslbErrorsTotal               *prometheus.CounterVec
	K8gbGslbReconciliationLoopsTotal  *prometheus.CounterVec
	K8gbInfobloxRequestDuration       *prometheus.HistogramVec
	K8gbInfobloxZoneUpdatesTotal      *prometheus.CounterVec
	K8gbInfobloxZoneUpdateErrorsTotal *prometheus.CounterVec
	K8gbInfobloxHeartbeatsTotal       *prometheus.CounterVec
	K8gbInfobloxHeartbeatErrorsTotal  *prometheus.CounterVec
	K8gbEndpointStatusNum             *prometheus.GaugeVec
	K8gbRuntimeInfo                   *prometheus.GaugeVec
}

type PrometheusMetrics struct {
	once    sync.Once
	config  resolver.Config
	metrics collectors
}

// DNSProviderRequest is a label for histogram metric
type DNSProviderRequest string

const (
	CreateZoneDelegated DNSProviderRequest = "ZoneCreate"
	GetZoneDelegated    DNSProviderRequest = "ZoneRead"
	UpdateZoneDelegated DNSProviderRequest = "ZoneUpdate"
	DeleteZoneDelegated DNSProviderRequest = "ZoneDelete"

	CreateTXTRecord = "TXTRecordCreate"
	GetTXTRecord    = "TXTRecordRead"
	UpdateTXTRecord = "TXTRecordUpdate"
	DeleteTXTRecord = "TXTRecordDelete"
)

var regex = regexp.MustCompile("[A-Z]")

// newPrometheusMetrics creates new prometheus metrics instance
func newPrometheusMetrics(config resolver.Config) (metrics *PrometheusMetrics) {
	metrics = new(PrometheusMetrics)
	metrics.config = config
	metrics.init()
	return
}

func (m *PrometheusMetrics) UpdateIngressHostsPerStatusMetric(gslb *k8gbv1beta1.Gslb, serviceHealth map[string]k8gbv1beta1.HealthStatus) {
	var healthyHostsCount, unhealthyHostsCount, notFoundHostsCount int
	for _, hs := range serviceHealth {
		switch hs {
		case k8gbv1beta1.Healthy:
			healthyHostsCount++
		case k8gbv1beta1.Unhealthy:
			unhealthyHostsCount++
		default:
			notFoundHostsCount++
		}
	}
	m.metrics.K8gbGslbServiceStatusNum.
		With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name, "status": k8gbv1beta1.Healthy.String()}).Set(float64(healthyHostsCount))
	m.metrics.K8gbGslbServiceStatusNum.
		With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name, "status": k8gbv1beta1.Unhealthy.String()}).Set(float64(unhealthyHostsCount))
	m.metrics.K8gbGslbServiceStatusNum.
		With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name, "status": k8gbv1beta1.NotFound.String()}).Set(float64(notFoundHostsCount))
}

func (m *PrometheusMetrics) UpdateHealthyRecordsMetric(gslb *k8gbv1beta1.Gslb, healthyRecords map[string][]string) {
	var hrsCount int
	for _, hrs := range healthyRecords {
		hrsCount += len(hrs)
	}
	m.metrics.K8gbGslbHealthyRecords.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name}).Set(float64(hrsCount))
}

func (m *PrometheusMetrics) UpdateEndpointStatus(ep *externaldns.DNSEndpoint) {
	for _, e := range ep.Spec.Endpoints {
		m.metrics.K8gbEndpointStatusNum.With(prometheus.Labels{"namespace": ep.Namespace, "name": ep.Name, "dns_name": e.DNSName}).
			Set(float64(e.Targets.Len()))
	}
}

func (m *PrometheusMetrics) UpdateFailoverStatus(gslb *k8gbv1beta1.Gslb, isPrimary bool, healthy k8gbv1beta1.HealthStatus, targets []string) {
	t := secondary
	if isPrimary {
		t = primary
	}
	m.updateRuntimeStatus(gslb, m.metrics.K8gbGslbStatusCountForFailover, healthy, targets, "_"+t)
}

func (m *PrometheusMetrics) UpdateRoundrobinStatus(gslb *k8gbv1beta1.Gslb, healthy k8gbv1beta1.HealthStatus, targets []string) {
	m.updateRuntimeStatus(gslb, m.metrics.K8gbGslbStatusCountForRoundrobin, healthy, targets, "")
}

func (m *PrometheusMetrics) UpdateGeoIPStatus(gslb *k8gbv1beta1.Gslb, healthy k8gbv1beta1.HealthStatus, targets []string) {
	m.updateRuntimeStatus(gslb, m.metrics.K8gbGslbStatusCountForGeoip, healthy, targets, "")
}

func (m *PrometheusMetrics) IncrementError(gslb *k8gbv1beta1.Gslb) {
	m.metrics.K8gbGslbErrorsTotal.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name}).Inc()
}

func (m *PrometheusMetrics) IncrementReconciliation(gslb *k8gbv1beta1.Gslb) {
	m.metrics.K8gbGslbReconciliationLoopsTotal.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name}).Inc()
}

func (m *PrometheusMetrics) InfobloxIncrementZoneUpdate(gslb *k8gbv1beta1.Gslb) {
	m.metrics.K8gbInfobloxZoneUpdatesTotal.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name}).Inc()
}

func (m *PrometheusMetrics) InfobloxIncrementZoneUpdateError(gslb *k8gbv1beta1.Gslb) {
	m.metrics.K8gbInfobloxZoneUpdateErrorsTotal.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name}).Inc()
}

func (m *PrometheusMetrics) InfobloxIncrementHeartbeat(gslb *k8gbv1beta1.Gslb) {
	m.metrics.K8gbInfobloxHeartbeatsTotal.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name}).Inc()
}

func (m *PrometheusMetrics) InfobloxIncrementHeartbeatError(gslb *k8gbv1beta1.Gslb) {
	m.metrics.K8gbInfobloxHeartbeatErrorsTotal.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name}).Inc()
}

func (m *PrometheusMetrics) InfobloxObserveRequestDuration(start time.Time, request DNSProviderRequest, success bool) {
	duration := time.Since(start).Seconds()
	m.metrics.K8gbInfobloxRequestDuration.With(prometheus.Labels{"request": string(request), "success": fmt.Sprintf("%t", success)}).Observe(duration)
}

func (m *PrometheusMetrics) SetRuntimeInfo(version, commit string) {
	firstN := func(value string, n int) string {
		if len(value) < n {
			return value
		}
		return value[:n]
	}

	m.metrics.K8gbRuntimeInfo.With(
		prometheus.Labels{"namespace": m.config.K8gbNamespace, "go_version": runtime.Version(), "arch": runtime.GOARCH,
			"os": runtime.GOOS, "k8gb_version": version, "git_sha": firstN(commit, 7)}).Set(1)
}

// Register prometheus metrics. Read register documentation, but shortly:
// You can register metric with given name only once
func (m *PrometheusMetrics) Register() (err error) {
	m.once.Do(func() {
		for _, r := range m.registry() {
			if err = crm.Registry.Register(r); err != nil {
				return
			}
		}
	})
	if err != nil {
		return fmt.Errorf("can't register prometheus metrics: %s", err)
	}
	return
}

// Unregister prometheus metrics
func (m *PrometheusMetrics) Unregister() {
	for _, r := range m.registry() {
		crm.Registry.Unregister(r)
	}
}

// init instantiates particular metrics
func (m *PrometheusMetrics) init() {

	m.metrics.K8gbRuntimeInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: K8gbRuntimeInfo,
			Help: "K8gb runtime info.",
		},
		[]string{"namespace", "k8gb_version", "go_version", "arch", "os", "git_sha"},
	)

	m.metrics.K8gbEndpointStatusNum = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: K8gbEndpointStatusNum,
			Help: "Number of targets in DNS endpoint.",
		},
		[]string{"namespace", "name", "dns_name"},
	)

	m.metrics.K8gbGslbHealthyRecords = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: K8gbGslbHealthyRecords,
			Help: "Number of healthy records observed by K8GB.",
		},
		[]string{"namespace", "name"},
	)

	m.metrics.K8gbGslbServiceStatusNum = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: K8gbGslbServiceStatusNum,
			Help: "Number of managed hosts observed by K8GB.",
		},
		[]string{"namespace", "name", "status"},
	)

	m.metrics.K8gbGslbErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: K8gbGslbErrorsTotal,
			Help: "Number of errors",
		},
		[]string{"namespace", "name"},
	)

	m.metrics.K8gbGslbReconciliationLoopsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: K8gbGslbReconciliationLoopsTotal,
			Help: "Number of successful reconciliation loops.",
		},
		[]string{"namespace", "name"},
	)

	m.metrics.K8gbGslbStatusCountForFailover = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: K8gbGslbStatusCountForFailover,
			Help: "Gslb status count for Failover strategy.",
		},
		[]string{"namespace", "name", "status"},
	)
	m.metrics.K8gbGslbStatusCountForRoundrobin = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: K8gbGslbStatusCountForRoundrobin,
			Help: "Gslb status count for RoundRobin strategy.",
		},
		[]string{"namespace", "name", "status"},
	)
	m.metrics.K8gbGslbStatusCountForGeoip = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: K8gbGslbStatusCountForGeoIP,
			Help: "Gslb status count for GeoIP strategy.",
		},
		[]string{"namespace", "name", "status"},
	)
	m.metrics.K8gbInfobloxRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    K8gbInfobloxRequestDuration,
			Help:    "How long it took for Infoblox requests to complete, partitioned by request type. Round-trip time of http communication is included.",
			Buckets: prometheus.ExponentialBuckets(.2, 4, 5),
		},
		[]string{"request", "success"},
	)

	m.metrics.K8gbInfobloxZoneUpdatesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: K8gbInfobloxZoneUpdatesTotal,
			Help: "Number of K8GB Infoblox zone updates.",
		},
		[]string{"namespace", "name"},
	)

	m.metrics.K8gbInfobloxZoneUpdateErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: K8gbInfobloxZoneUpdateErrorsTotal,
			Help: "Number of K8GB Infoblox zone update errors.",
		},
		[]string{"namespace", "name"},
	)
	m.metrics.K8gbInfobloxHeartbeatsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: K8gbInfobloxHeartbeatsTotal,
			Help: "Number of K8GB Infoblox heartbeat TXT record updates.",
		},
		[]string{"namespace", "name"},
	)
	m.metrics.K8gbInfobloxHeartbeatErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: K8gbInfobloxHeartbeatErrorsTotal,
			Help: "Number of K8GB Infoblox TXT record errors.",
		},
		[]string{"namespace", "name"},
	)
}

// registry is helper function reading fields from m.metrics structure and builds metrics map
// The key is metric name while value is metric instance
func (m *PrometheusMetrics) registry() (r map[string]prometheus.Collector) {
	r = make(map[string]prometheus.Collector)
	val := reflect.Indirect(reflect.ValueOf(m.metrics))
	for i := 0; i < val.Type().NumField(); i++ {
		n := val.Type().Field(i).Name
		if !val.Field(i).IsNil() {
			var v = val.FieldByName(n).Interface().(prometheus.Collector)
			name := strings.ToLower(strings.Join(utils.SplitAfter(n, regex), "_"))
			r[name] = v
		}
	}
	return
}

func (m *PrometheusMetrics) updateRuntimeStatus(
	gslb *k8gbv1beta1.Gslb,
	vec *prometheus.GaugeVec,
	healthStatus k8gbv1beta1.HealthStatus,
	targets []string,
	tag string) {
	var h, u, n int
	switch healthStatus {
	case k8gbv1beta1.Healthy:
		h = len(targets)
	case k8gbv1beta1.Unhealthy:
		u = len(targets)
	case k8gbv1beta1.NotFound:
		n = len(targets)
	}
	vec.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name, "status": fmt.Sprintf("%s%s", k8gbv1beta1.Healthy, tag)}).
		Set(float64(h))
	vec.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name, "status": fmt.Sprintf("%s%s", k8gbv1beta1.Unhealthy, tag)}).
		Set(float64(u))
	vec.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name, "status": fmt.Sprintf("%s%s", k8gbv1beta1.NotFound, tag)}).
		Set(float64(n))
}
