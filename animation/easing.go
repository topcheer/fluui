package animation

import "math"

// Easing is a function that maps a raw linear progress value in [0, 1] to
// an eased progress value, also typically in [0, 1]. Easing functions are
// used to make transitions feel natural by accelerating/decelerating over time.
type Easing func(t float64) float64

// --- Identity / linear ---

// Linear is the identity easing: returns t unchanged.
func Linear(t float64) float64 {
	return t
}

// --- Power-based easings ---

// EaseIn starts slow and accelerates (quadratic by default).
func EaseIn(t float64) float64 {
	return t * t
}

// EaseOut starts fast and decelerates (quadratic by default).
func EaseOut(t float64) float64 {
	return 1 - (1-t)*(1-t)
}

// EaseInOut combines ease-in and ease-out: slow at both ends, fast in middle.
func EaseInOut(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return 1 - math.Pow(-2*t+2, 2)/2
}

// --- Cubic variants ---

// EaseInCubic starts very slow (cubic acceleration).
func EaseInCubic(t float64) float64 {
	return t * t * t
}

// EaseOutCubic ends very slow (cubic deceleration).
func EaseOutCubic(t float64) float64 {
	return 1 - math.Pow(1-t, 3)
}

// EaseInOutCubic combines cubic ease-in and ease-out.
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// --- Overshoot / elastic ---

// EaseOutBack overshoots slightly past 1, then settles back. Good for
// "pop-in" effects. back controls the overshoot amount (default 1.70158).
func EaseOutBack(t float64, back ...float64) float64 {
	c1 := 1.70158
	if len(back) > 0 {
		c1 = back[0]
	}
	c3 := c1 + 1
	return 1 + c3*math.Pow(t-1, 3) + c1*math.Pow(t-1, 2)
}

// EaseInBack undershoots slightly before settling to 0. Good for
// "suck-in" exit effects.
func EaseInBack(t float64, back ...float64) float64 {
	c1 := 1.70158
	if len(back) > 0 {
		c1 = back[0]
	}
	c3 := c1 + 1
	return c3*t*t*t - c1*t*t
}

// EaseOutElastic oscillates a few times while settling, like a spring.
func EaseOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	c4 := (2 * math.Pi) / 3
	return math.Pow(2, -10*t) * math.Sin((t*10-0.75)*c4) + 1
}

// --- Bounce ---

// EaseOutBounce simulates a bouncing deceleration.
func EaseOutBounce(t float64) float64 {
	n1, d1 := 7.5625, 2.75
	if t < 1/d1 {
		return n1 * t * t
	}
	if t < 2/d1 {
		t -= 1.5 / d1
		return n1*t*t + 0.75
	}
	if t < 2.5/d1 {
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	}
	t -= 2.625 / d1
	return n1*t*t + 0.984375
}

// EaseInBounce is the reverse of EaseOutBounce.
func EaseInBounce(t float64) float64 {
	return 1 - EaseOutBounce(1-t)
}

// --- Helpers ---

// Clamp clamps a value to the [0, 1] range.
func Clamp(t float64) float64 {
	if t < 0 {
		return 0
	}
	if t > 1 {
		return 1
	}
	return t
}

// Lerp linearly interpolates between from and to by factor t (typically [0,1]).
func Lerp(from, to, t float64) float64 {
	return from + (to-from)*t
}

// EasedProgress takes a raw linear progress [0,1] and an easing function,
// returning the eased progress. If easing is nil, Linear is used.
func EasedProgress(t float64, easing Easing) float64 {
	if easing == nil {
		return Linear(t)
	}
	return easing(Clamp(t))
}
