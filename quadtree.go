package sketchy

import "github.com/tdewolff/canvas"

const defaultCapacity = 8

type QuadTree struct {
	capacity int
	points   []Point
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

func (q *QuadTree) Insert(p Point) bool {
	if !q.boundary.ContainsPoint(p) {
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

func (q *QuadTree) Query(r Rect) []Point {
	var results = []Point{}

	if !q.boundary.Intersects(r) {
		return results
	}

	for _, p := range q.points {
		if r.ContainsPoint(p) {
			results = append(results, p)
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

func (q *QuadTree) DrawWithPoints(size float64, ctx *canvas.Context) {
	q.boundary.Draw(ctx)
	for _, p := range q.points {
		p.Draw(size, ctx)
	}
	if q.ne == nil {
		return
	}

	q.ne.Draw(ctx)
	q.se.Draw(ctx)
	q.sw.Draw(ctx)
	q.nw.Draw(ctx)
}
