package sketchy

import (
	"github.com/tdewolff/canvas"
)

const defaultCapacity = 4

type QuadTree struct {
	capacity int
	points   []IndexPoint
	boundary Rect

	nw *QuadTree
	ne *QuadTree
	se *QuadTree
	sw *QuadTree
}

func NewQuadTree(r Rect) *QuadTree {
	return &QuadTree{
		capacity: defaultCapacity,
		boundary: r,
	}
}

func NewQuadTreeWithCapacity(r Rect, c int) *QuadTree {
	return &QuadTree{
		capacity: c,
		boundary: r,
	}
}

func (q *QuadTree) Insert(p IndexPoint) bool {
	if !q.boundary.ContainsPoint(p.Point) {
		return false
	}

	if len(q.points) < q.capacity && q.ne == nil {
		q.points = append(q.points, p)
		return true
	}

	if q.ne == nil {
		q.subdivide()
	}

	if q.se.Insert(p) {
		return true
	}
	if q.sw.Insert(p) {
		return true
	}
	if q.nw.Insert(p) {
		return true
	}
	if q.ne.Insert(p) {
		return true
	}
	return false
}

func (q *QuadTree) Search(p IndexPoint) *IndexPoint {
	var result *IndexPoint
	if !q.boundary.ContainsPoint(p.Point) {
		return nil
	}
	for _, point := range q.points {
		if point.Point.IsEqual(p.Point) {
			return &point
		}
	}
	if q.ne == nil {
		return nil
	}
	result = q.ne.Search(p)
	if result != nil {
		return result
	}
	result = q.se.Search(p)
	if result != nil {
		return result
	}
	result = q.sw.Search(p)
	if result != nil {
		return result
	}
	result = q.nw.Search(p)
	if result != nil {
		return result
	}
	return nil
}

func (q *QuadTree) UpdateIndex(p IndexPoint, index int) *IndexPoint {
	var result *IndexPoint
	if !q.boundary.ContainsPoint(p.Point) {
		return nil
	}
	for i := range q.points {
		if q.points[i].Point.IsEqual(p.Point) {
			q.points[i].Index = index
			return &q.points[i]
		}
	}
	if q.ne == nil {
		return nil
	}
	result = q.ne.UpdateIndex(p, index)
	if result != nil {
		return result
	}
	result = q.se.UpdateIndex(p, index)
	if result != nil {
		return result
	}
	result = q.sw.UpdateIndex(p, index)
	if result != nil {
		return result
	}
	result = q.nw.UpdateIndex(p, index)
	if result != nil {
		return result
	}
	return nil
}

func (q *QuadTree) Query(r Rect) []Point {
	var results = []Point{}

	if q.boundary.IsDisjoint(r) {
		return results
	}

	for _, p := range q.points {
		if r.ContainsPoint(p.Point) {
			results = append(results, p.Point)
		}
	}

	if q.ne == nil {
		return results
	}

	results = append(results, q.ne.Query(r)...)
	results = append(results, q.se.Query(r)...)
	results = append(results, q.sw.Query(r)...)
	results = append(results, q.nw.Query(r)...)

	return results
}

func (q *QuadTree) QueryExcludeIndex(r Rect, index int) []Point {
	var results = []Point{}

	if q.boundary.IsDisjoint(r) {
		return results
	}

	for _, p := range q.points {
		if r.ContainsPoint(p.Point) && p.Index != index {
			results = append(results, p.Point)
		}
	}

	if q.ne == nil {
		return results
	}

	results = append(results, q.ne.QueryExcludeIndex(r, index)...)
	results = append(results, q.se.QueryExcludeIndex(r, index)...)
	results = append(results, q.sw.QueryExcludeIndex(r, index)...)
	results = append(results, q.nw.QueryExcludeIndex(r, index)...)

	return results
}

func (q *QuadTree) QueryCircle(center Point, radius float64) []Point {
	rect := Rect{
		X: center.X - radius,
		Y: center.Y - radius,
		W: radius,
		H: radius,
	}
	rectQuery := q.Query(rect)
	var results []Point
	R2 := radius * radius
	for _, p := range rectQuery {
		if SquaredDistance(center, p) < R2 {
			results = append(results, p)
		}
	}
	return results
}

func (q *QuadTree) QueryCircleExcludeIndex(center Point, radius float64, index int) []Point {
	rect := Rect{
		X: center.X - radius,
		Y: center.Y - radius,
		W: radius,
		H: radius,
	}
	rectQuery := q.QueryExcludeIndex(rect, index)
	var results []Point
	R2 := radius * radius
	for _, p := range rectQuery {
		if SquaredDistance(center, p) < R2 {
			results = append(results, p)
		}
	}
	return results
}

func (q *QuadTree) NearestNeighbors(point IndexPoint, k int) []IndexPoint {
	var result []IndexPoint
	if k <= 0 {
		return result
	}
	ph := NewMaxPointHeap()
	q.pushOnHeap(point, ph, k)
	points := ph.ReportReversed()
	for _, p := range points {
		result = append(result, p.ToIndexPoint())
	}
	return result
}

func (q *QuadTree) Clear() {
	q.points = []IndexPoint{}
	q.ne = nil
	q.se = nil
	q.sw = nil
	q.nw = nil
}

func (q *QuadTree) Size() int {
	var count int
	count += len(q.points)
	if q.ne == nil {
		return count
	}
	count += q.ne.Size()
	count += q.se.Size()
	count += q.sw.Size()
	count += q.nw.Size()
	return count
}

func (q *QuadTree) Draw(ctx *canvas.Context) {
	q.boundary.Draw(ctx)

	if q.ne == nil {
		return
	}

	q.ne.Draw(ctx)
	q.se.Draw(ctx)
	q.sw.Draw(ctx)
	q.nw.Draw(ctx)
}

func (q *QuadTree) DrawWithPoints(s float64, ctx *canvas.Context) {
	for _, p := range q.points {
		p.Draw(s, ctx)
	}
	q.boundary.Draw(ctx)
	if q.ne == nil {
		return
	}

	q.ne.DrawWithPoints(s, ctx)
	q.se.DrawWithPoints(s, ctx)
	q.sw.DrawWithPoints(s, ctx)
	q.nw.DrawWithPoints(s, ctx)
}

func (q *QuadTree) subdivide() {
	x := q.boundary.X
	y := q.boundary.Y
	w := q.boundary.W / 2
	h := q.boundary.H / 2
	q.ne = NewQuadTree(Rect{X: x + w, Y: y, W: w, H: h})
	q.se = NewQuadTree(Rect{X: x + w, Y: y + h, W: w, H: h})
	q.sw = NewQuadTree(Rect{X: x, Y: y + h, W: w, H: h})
	q.nw = NewQuadTree(Rect{X: x, Y: y, W: w, H: h})
}

func (q *QuadTree) pushOnHeap(target IndexPoint, h *PointHeap, k int) {
	for _, p := range q.points {
		if p.Index == target.Index {
			continue
		}
		metric := SquaredDistance(target.Point, p.Point)
		mp := MetricPoint{
			Metric: metric,
			Index:  p.Index,
			Point:  p.Point,
		}
		if h.Len() < k {
			h.Push(mp)
		} else {
			if metric < h.Peek().Metric {
				_ = h.Pop()
				h.Push(mp)
			}
		}
	}
	if q.ne == nil {
		return
	}
	q.ne.pushOnHeap(target, h, k)
	q.se.pushOnHeap(target, h, k)
	q.sw.pushOnHeap(target, h, k)
	q.nw.pushOnHeap(target, h, k)
}
