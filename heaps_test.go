package sketchy

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestPointHeap_InsertPop(t *testing.T) {
	assert := assert.New(t)
	maxHeap := NewMaxPointHeap()
	minHeap := NewMinPointHeap()
	metrics := []float64{5, 3, 2, 7, 20}
	for _, m := range metrics {
		p := MetricPoint{
			Metric: m,
			Index:  0,
			Point:  Point{X: 0, Y: 0},
		}
		maxHeap.Push(p)
		minHeap.Push(p)
	}
	n := len(metrics)
	sort.Float64s(metrics)
	for i, m := range metrics {
		assert.Equal(m, minHeap.Pop().Metric)
		assert.Equal(metrics[n-i-1], maxHeap.Pop().Metric)
	}
}

func TestPointHeap_Report(t *testing.T) {
	assert := assert.New(t)
	maxHeap := NewMaxPointHeap()
	minHeap := NewMinPointHeap()
	metrics := []float64{5, 3, 2, 7, 20}
	for _, m := range metrics {
		p := MetricPoint{
			Metric: m,
			Index:  0,
			Point:  Point{X: 0, Y: 0},
		}
		maxHeap.Push(p)
		minHeap.Push(p)
	}
	sort.Float64s(metrics)
	minHeapReport := minHeap.Report()
	maxHeapReportReversed := maxHeap.ReportReversed()
	for i := range minHeapReport {
		assert.Equal(metrics[i], minHeapReport[i].Metric)
		assert.Equal(minHeapReport[i].Metric, maxHeapReportReversed[i].Metric)
	}
}
