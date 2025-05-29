package reduction

import (
	"duolingo/lib/metric"
	"errors"
	"math"
)

type LTTB struct {
	source SnapshotReduction
}

func (l *LTTB) UseSource(src SnapshotReduction) {
	l.source = src
}

func (l *LTTB) Make(reduction int64, reductionB []*metric.Snapshot) (*metric.Snapshot, error) {
	if len(reductionB) == 0 {
		return nil, errors.New("reduction is empty")
	}

	prevReduction, errPrev := l.source.PreviousReduction(reduction)
	nextReduction, errNext := l.source.NextReduction(reduction)

	if errPrev != nil || errNext != nil {
		// Fallback to simple average if we're on the edge
		return (&Median{source: l.source}).Make(reduction, reductionB)
	}

	// A = previous point (from previous reduction)
	aPoints := l.source.GetSnapshots(prevReduction)
	if len(aPoints) == 0 {
		return (&Median{source: l.source}).Make(reduction, reductionB)
	}
	a := aPoints[len(aPoints)-1] // choose the latest point in previous reduction
	aX := float64(a.Timestamp.UnixMilli())
	aY := a.Value

	// C = average of next reduction
	cPoints := l.source.GetSnapshots(nextReduction)
	if len(cPoints) == 0 {
		return (&Median{source: l.source}).Make(reduction, reductionB)
	}
	var sumX, sumY float64
	for _, p := range cPoints {
		sumX += float64(p.Timestamp.UnixMilli())
		sumY += p.Value
	}
	avgX := sumX / float64(len(cPoints))
	avgY := sumY / float64(len(cPoints))

	// B = find max triangle area point in this reduction
	var maxArea float64 = -1
	var maxPoint *metric.Snapshot

	for _, p := range reductionB {
		pX := float64(p.Timestamp.UnixMilli())
		pY := p.Value

		area := math.Abs((aX-avgX)*(pY-aY) - (aX-pX)*(avgY-aY))

		if area > maxArea {
			maxArea = area
			maxPoint = p
		}
	}

	if maxPoint == nil {
		return (&Median{source: l.source}).Make(reduction, reductionB)
	}

	return maxPoint, nil
}
