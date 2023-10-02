package ui

import (
	"math"
	"time"
)

type Tween struct {
	Ended bool

	value       *float64
	startValue  float64
	targetValue float64
	startTime   int64
	animLength  time.Duration
}

func NewTween(value *float64, target float64, length time.Duration) *Tween {
	return &Tween{
		value:       value,
		startValue:  *value,
		targetValue: target,
		startTime:   time.Now().UnixMilli(),
		animLength:  length,
	}
}

func (t *Tween) Update() {
	if t.Ended {
		return
	}

	p := float64(time.Now().UnixMilli()-t.startTime) / float64(t.animLength.Milliseconds())
	if p > 1.0 {
		*t.value = t.targetValue
		t.Ended = true
	} else {
		*t.value = t.startValue + t.easeInOutQuad(p)*(t.targetValue-t.startValue)
	}
}

func (t *Tween) easeInOutQuad(x float64) float64 {
	if x < 0.5 {
		return 2 * x * x
	}

	return 1 - math.Pow(-2*x+2, 2)/2
}

type Tweens []*Tween

func (t *Tweens) Add(newTween *Tween) {
	*t = append(*t, newTween)
}

func (t *Tweens) Update() {
	var notEnded []*Tween
	for _, tween := range *t {
		tween.Update()
		if !tween.Ended {
			notEnded = append(notEnded, tween)
		}
	}
	*t = notEnded
}

func (t Tweens) CancelAll() {
	for _, tween := range t {
		tween.Ended = true
	}
}
