package worker

import (
	"math"
	"math/rand"
	"time"
)

//nolint:gosec
var randFloat = rand.New(rand.NewSource(time.Now().UnixNano())).Float64

// BackoffStrategy defines the different kind of Backoff strategy for retry.
// eg. Linear, Const, Exponential ..etc
type BackoffStrategy interface {
	// Backoff returns how much duration to wait for a given attempt
	Backoff(attempt int) time.Duration
}

// BackoffFunc is a adapter to use ordinary function as retry BackoffStrategy
type BackoffFunc func(attempt int) time.Duration

func (s BackoffFunc) Backoff(attempt int) time.Duration { return s(attempt) }

type ConstBackoff struct {
	// Delay is the time duration to wait before each retry attempt
	Delay time.Duration
}

func (c ConstBackoff) Backoff(int) time.Duration { return c.Delay }

// LinearModBackoff will backoff linearly. It will start at InitialDelay.
// Backoff duration will oscillate linearly between [0,MaxDelay] forming a sawtooth pattern.
// refer https://www.researchgate.net/figure/Types-of-back-off-algorithms-constant-linear-linear-modulus-exponential-and_fig1_224440820
type LinearModBackoff struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
}

func (l LinearModBackoff) Backoff(attempt int) time.Duration {
	return (l.InitialDelay * time.Duration(attempt)) % l.MaxDelay
}

// LinearBackoff will backoff linearly. It will start at InitialDelay, capping at MaxDelay.
type LinearBackoff struct {
	// backoff will start at InitialDelay, i.e. backoff duration for 1st retry attempt
	InitialDelay time.Duration
	// Backoff duration will be capped at MaxDelay duration.
	MaxDelay time.Duration
}

func (l LinearBackoff) Backoff(attempt int) time.Duration {
	if d := l.InitialDelay * time.Duration(attempt); d < l.MaxDelay {
		return d
	}
	return l.MaxDelay
}

// ExponentialBackoff implements exponential backoff. It is capped at MaxDelay.
type ExponentialBackoff struct {
	// InitialDelay is multiplied by Multiplier after each attempt
	Multiplier float64
	// InitialDelay is the initial duration for retrial, i.e. backoff duration for 1st retry attempt
	InitialDelay time.Duration
	// Backoff duration will be capped at MaxDelay.
	MaxDelay time.Duration
	// Amount of jitter applied after each iteration.
	// This can be used to randomize duration after each attempt
	Jitter float64
}

func (b *ExponentialBackoff) Backoff(attempt int) time.Duration {
	duration := b.InitialDelay * time.Duration(math.Pow(b.Multiplier, float64(attempt-1)))

	if b.MaxDelay > 0 && duration > b.MaxDelay {
		duration = b.MaxDelay
	}

	if b.Jitter > 0 {
		duration += time.Duration(randFloat() * b.Jitter * float64(duration))
	}

	return duration
}

var DefaultExponentialBackoff = &ExponentialBackoff{
	Multiplier:   1.6,
	InitialDelay: 1 * time.Second,
	MaxDelay:     900 * time.Second,
	Jitter:       0.2,
}
