package sketchy

// PointHeap is a min/max heap for Points using inter-point distance as the metric
type PointHeap struct {
	size      int
	points    []MetricPoint
	isMinHeap bool
}

func NewMaxPointHeap() *PointHeap {
	return &PointHeap{
		size:      0,
		points:    []MetricPoint{},
		isMinHeap: false,
	}
}

func NewMinPointHeap() *PointHeap {
	return &PointHeap{
		size:      0,
		points:    []MetricPoint{},
		isMinHeap: true,
	}
}

func (m *PointHeap) Len() int {
	return m.size
}

func (m *PointHeap) Push(p MetricPoint) {
	m.points = append(m.points, p)
	index := m.size
	m.size++
	if m.isMinHeap {
		for m.points[index].Metric < m.points[m.parent(index)].Metric {
			m.swap(index, m.parent(index))
			index = m.parent(index)
		}
	} else {
		for m.points[index].Metric > m.points[m.parent(index)].Metric {
			m.swap(index, m.parent(index))
			index = m.parent(index)
		}
	}
}

func (m *PointHeap) Peek() MetricPoint {
	return m.points[0]
}

func (m *PointHeap) Pop() MetricPoint {
	if m.size == 0 {
		panic("can't pop empty heap")
	}
	p := m.points[0]
	m.points = m.points[1:m.size]
	m.size--
	m.heapify(0)
	return p
}

func (m *PointHeap) Report() []MetricPoint {
	n := m.size
	result := make([]MetricPoint, n)
	for i := 0; i < n; i++ {
		result[i] = m.Pop()
	}
	return result
}

func (m *PointHeap) ReportReversed() []MetricPoint {
	n := m.size
	result := make([]MetricPoint, n)
	for i := 0; i < n; i++ {
		result[n-i-1] = m.Pop()
	}
	return result
}

func (m *PointHeap) parent(i int) int {
	return (i - 1) / 2
}

func (m *PointHeap) left(i int) int {
	return 2*i + 1
}

func (m *PointHeap) right(i int) int {
	return 2*i + 2
}

func (m *PointHeap) swap(i, j int) {
	m.points[i], m.points[j] = m.points[j], m.points[i]
}

func (m *PointHeap) isLeaf(i int) bool {
	if i > (m.size/2) && i <= m.size {
		return true
	}
	return false
}

func (m *PointHeap) heapify(i int) {
	if m.size <= 1 {
		return
	}
	if m.isLeaf(i) {
		return
	}
	l := m.left(i)
	r := m.right(i)
	if m.isMinHeap {
		var min int
		if l < m.size && m.points[l].Metric < m.points[i].Metric {
			min = l
		} else {
			min = i
		}
		if r < m.size && m.points[r].Metric < m.points[min].Metric {
			min = r
		}
		if min != i {
			m.swap(i, min)
			m.heapify(min)
		}
	} else {
		var max int
		if l < m.size && m.points[l].Metric > m.points[i].Metric {
			max = l
		} else {
			max = i
		}
		if r < m.size && m.points[r].Metric > m.points[max].Metric {
			max = r
		}
		if max != i {
			m.swap(i, max)
			m.heapify(max)
		}
	}
}
