package downsampling

import (
	"errors"
	"math"
)

type LTTB struct {
	source ReducedDataPoints
}

func (l *LTTB) UseSource(src ReducedDataPoints) {
	l.source = src
}

func (l *LTTB) Make(reduction int64, reductionB []*DataPoint) (*DataPoint, error) {
	if len(reductionB) == 0 {
		return nil, errors.New("reduction is empty")
	}

	prevReduction, errPrev := l.source.PreviousReduction(reduction)
	nextReduction, errNext := l.source.NextReduction(reduction)

	if errPrev != nil || errNext != nil {
		// Fallback to simple average if we're on the edge
		return (&MovingAverage{source: l.source}).Make(reduction, reductionB)
	}

	// A = previous point (from previous reduction)
	aPoints := l.source.GetReducedDataPoints(prevReduction)
	if len(aPoints) == 0 {
		return (&MovingAverage{source: l.source}).Make(reduction, reductionB)
	}
	a := aPoints[len(aPoints)-1] // choose the latest point in previous reduction
	aX := float64(a.GetTimestamp().UnixMilli())
	aY := a.GetValue()

	// C = average of next reduction
	cPoints := l.source.GetReducedDataPoints(nextReduction)
	if len(cPoints) == 0 {
		return (&MovingAverage{source: l.source}).Make(reduction, reductionB)
	}
	var sumX, sumY float64
	for _, p := range cPoints {
		sumX += float64(p.GetTimestamp().UnixMilli())
		sumY += p.GetValue()
	}
	avgX := sumX / float64(len(cPoints))
	avgY := sumY / float64(len(cPoints))

	// B = find max triangle area point in this reduction
	var maxArea float64 = -1
	var maxPoint *DataPoint

	for _, p := range reductionB {
		pX := float64(p.GetTimestamp().UnixMilli())
		pY := p.GetValue()

		area := math.Abs((aX-avgX)*(pY-aY) - (aX-pX)*(avgY-aY))

		if area > maxArea {
			maxArea = area
			maxPoint = p
		}
	}

	if maxPoint == nil {
		return (&MovingAverage{source: l.source}).Make(reduction, reductionB)
	}

	return maxPoint, nil
}