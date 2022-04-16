package sketchy

import (
	"github.com/tdewolff/canvas"
)

type KDTree struct {
	point  *IndexPoint
	region Rect
	left   *KDTree
	right  *KDTree
}

func NewKDTree(r Rect) *KDTree {
	return &KDTree{
		point:  nil,
		region: r,
		left:   nil,
		right:  nil,
	}
}

func NewKDTreeWithPoint(p IndexPoint, r Rect) *KDTree {
	return &KDTree{
		point:  &p,
		region: r,
		left:   nil,
		right:  nil,
	}
}

func (k *KDTree) IsLeaf() bool {
	return k.left == nil && k.right == nil
}

func (k *KDTree) Insert(p IndexPoint) {
	k.insert(p, 0)
}

func (k *KDTree) Query(r Rect) []Point {
	var results []Point
	if k.point != nil && r.ContainsPoint(k.point.Point) {
		results = append(results, k.point.Point)
	}
	if k.left != nil {
		if r.Contains(k.left.region) {
			results = append(results, k.left.reportSubtree()...)
		} else {
			if r.Overlaps(k.left.region) {
				query := k.left.Query(r)
				if len(query) > 0 {
					results = append(results, query...)
				}
			}
		}
	}
	if k.right != nil {
		if r.Contains(k.right.region) {
			results = append(results, k.right.reportSubtree()...)
		} else {
			if r.Overlaps(k.right.region) {
				query := k.right.Query(r)
				if len(query) > 0 {
					results = append(results, query...)
				}
			}
		}
	}
	return results
}

func (k *KDTree) NearestNeighbors(point IndexPoint, s int) []IndexPoint {
	var result []IndexPoint
	if s <= 0 {
		return result
	}
	ph := NewMaxPointHeap()
	k.pushOnHeap(point, ph, s)
	points := ph.ReportReversed()
	for _, p := range points {
		result = append(result, p.ToIndexPoint())
	}
	return result
}

func (k *KDTree) Clear() {
	k.point = nil
	k.left = nil
	k.right = nil
}

func (k *KDTree) Size() int {
	count := 1
	if k.left != nil {
		count += k.left.Size()
	}
	if k.right != nil {
		count += k.right.Size()
	}
	return count
}

func (k *KDTree) Draw(ctx *canvas.Context) {
	k.draw(ctx, 0, 0)
}

func (k *KDTree) DrawWithPoints(s float64, ctx *canvas.Context) {
	k.draw(ctx, 0, s)
}

func (k *KDTree) insert(p IndexPoint, d int) {
	if k.point == nil {
		k.point = &p
		return
	}
	if d%2 == 0 { // compare x value
		if p.X < k.point.X {
			if k.left == nil {
				rect := Rect{
					X: k.region.X,
					Y: k.region.Y,
					W: k.point.X - k.region.X,
					H: k.region.H,
				}
				k.left = NewKDTreeWithPoint(p, rect)
			} else {
				k.left.insert(p, d+1)
			}
		} else {
			if k.right == nil {
				rect := Rect{
					X: k.point.X,
					Y: k.region.Y,
					W: k.region.W - (k.point.X - k.region.X),
					H: k.region.H,
				}
				k.right = NewKDTreeWithPoint(p, rect)
			} else {
				k.right.insert(p, d+1)
			}
		}
	} else { // compare y value
		if p.Y < k.point.Y {
			if k.left == nil {
				rect := Rect{
					X: k.region.X,
					Y: k.region.Y,
					W: k.region.W,
					H: k.point.Y - k.region.Y,
				}
				k.left = NewKDTreeWithPoint(p, rect)
			} else {
				k.left.insert(p, d+1)
			}
		} else {
			if k.right == nil {
				rect := Rect{
					X: k.region.X,
					Y: k.point.Y,
					W: k.region.W,
					H: k.region.H - (k.point.Y - k.region.Y),
				}
				k.right = NewKDTreeWithPoint(p, rect)
			} else {
				k.right.insert(p, d+1)
			}
		}
	}
}

func (k *KDTree) reportSubtree() []Point {
	var results []Point
	results = append(results, k.point.Point)
	if k.left != nil {
		results = append(results, k.left.reportSubtree()...)
	}
	if k.right != nil {
		results = append(results, k.right.reportSubtree()...)
	}
	return results
}

func (k *KDTree) draw(ctx *canvas.Context, depth int, pointSize float64) {
	if k.point == nil {
		return
	}
	if depth%2 == 0 {
		ctx.MoveTo(k.point.X, k.region.Y)
		ctx.LineTo(k.point.X, k.region.Y+k.region.H)
		ctx.Stroke()
	} else {
		ctx.MoveTo(k.region.X, k.point.Y)
		ctx.LineTo(k.region.X+k.region.W, k.point.Y)
		ctx.Stroke()
	}
	if pointSize > 0 {
		ctx.DrawPath(k.point.X, k.point.Y, canvas.Circle(pointSize))
	}

	if k.left != nil {
		k.left.draw(ctx, depth+1, pointSize)
	}
	if k.right != nil {
		k.right.draw(ctx, depth+1, pointSize)
	}
}

func (k *KDTree) pushOnHeap(target IndexPoint, h *PointHeap, s int) {
	if k.point != nil {
		metric := SquaredDistance(target.Point, k.point.Point)
		mp := MetricPoint{
			Metric: metric,
			Index:  k.point.Index,
			Point:  k.point.Point,
		}
		if h.Len() < s {
			h.Push(mp)
		} else {
			if metric < h.Peek().Metric {
				_ = h.Pop()
				h.Push(mp)
			}
		}
	}
	if k.left != nil {
		k.left.pushOnHeap(target, h, s)
	}
	if k.right != nil {
		k.right.pushOnHeap(target, h, s)
	}
}
