package worker //nolint:testpackage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConstBackoff(t *testing.T) {
	cases := []struct {
		b        ConstBackoff
		attempt  int
		expected time.Duration
	}{
		0: {
			b:        ConstBackoff{Delay: 1 * time.Second},
			attempt:  1,
			expected: time.Second,
		},
		1: {
			b:        ConstBackoff{Delay: 133 * time.Millisecond},
			attempt:  2,
			expected: 133 * time.Millisecond,
		},
		2: {
			b:        ConstBackoff{Delay: 350 * time.Millisecond},
			attempt:  10,
			expected: 350 * time.Millisecond,
		},
	}
	for i, tc := range cases {
		assert.Equal(t, tc.expected, tc.b.Backoff(tc.attempt), "test[%d]", i)
	}
}

func TestLinearModBackoff(t *testing.T) {
	cases := []struct {
		b        LinearModBackoff
		attempt  int
		expected time.Duration
	}{
		0: {
			b: LinearModBackoff{
				InitialDelay: 900 * time.Millisecond,
				MaxDelay:     3 * time.Second,
			},
			attempt:  1,
			expected: 900 * time.Millisecond,
		},
		1: {
			b: LinearModBackoff{
				InitialDelay: 900 * time.Millisecond,
				MaxDelay:     3 * time.Second,
			},
			attempt:  2,
			expected: 1800 * time.Millisecond,
		},
		2: {
			b: LinearModBackoff{
				InitialDelay: 900 * time.Millisecond,
				MaxDelay:     3 * time.Second,
			},
			attempt:  3,
			expected: 2700 * time.Millisecond,
		},
		3: {
			b: LinearModBackoff{
				InitialDelay: 900 * time.Millisecond,
				MaxDelay:     3 * time.Second,
			},
			attempt:  4,
			expected: 600 * time.Millisecond,
		},
		4: {
			b: LinearModBackoff{
				InitialDelay: 900 * time.Millisecond,
				MaxDelay:     3 * time.Second,
			},
			attempt:  5,
			expected: 1500 * time.Millisecond,
		},
	}
	for i, tc := range cases {
		assert.Equal(t, tc.expected, tc.b.Backoff(tc.attempt), "test[%d]", i)
	}
}

func TestLinearBackoff(t *testing.T) {
	cases := []struct {
		b        LinearBackoff
		attempt  int
		expected time.Duration
	}{
		0: {
			b: LinearBackoff{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     1 * time.Second,
			},
			attempt:  1,
			expected: 100 * time.Millisecond,
		},
		1: {
			b: LinearBackoff{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     1 * time.Second,
			},
			attempt:  5,
			expected: 500 * time.Millisecond,
		},
		2: {
			b: LinearBackoff{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     1 * time.Second,
			},
			attempt:  10,
			expected: 1 * time.Second,
		},
		3: {
			b: LinearBackoff{
				InitialDelay: 100 * time.Millisecond,
				MaxDelay:     1 * time.Second,
			},
			attempt:  11,
			expected: 1 * time.Second,
		},
	}
	for i, tc := range cases {
		assert.Equal(t, tc.expected, tc.b.Backoff(tc.attempt), "test[%d]", i)
	}
}

func TestExponentialBackoff(t *testing.T) {
	randFloat = func() float64 { return 0.5 }

	cases := []struct {
		b        *ExponentialBackoff
		attempt  int
		expected time.Duration
	}{
		0: {
			b: &ExponentialBackoff{
				Multiplier:   2,
				InitialDelay: time.Second * 4,
				MaxDelay:     time.Second * 5,
			},
			attempt:  1,
			expected: time.Second * 4,
		},
		1: {
			b: &ExponentialBackoff{
				Multiplier:   2,
				InitialDelay: time.Second * 4,
			},
			attempt:  3,
			expected: time.Second * 16,
		},
		2: {
			b: &ExponentialBackoff{
				Multiplier:   2,
				InitialDelay: time.Second * 4,
				MaxDelay:     time.Second * 10,
			},
			attempt:  3,
			expected: time.Second * 10,
		},
		3: {
			b: &ExponentialBackoff{
				Multiplier:   1,
				InitialDelay: time.Second * 4,
			},
			attempt:  10,
			expected: time.Second * 4,
		},
		4: {
			b: &ExponentialBackoff{
				Multiplier:   2,
				InitialDelay: time.Second * 4,
			},
			attempt:  3,
			expected: time.Second * 16,
		},
		5: {
			b: &ExponentialBackoff{
				Multiplier:   2,
				InitialDelay: time.Second * 4,
				Jitter:       0.4,
			},
			attempt:  3,
			expected: time.Second*19 + time.Millisecond*200,
		},
		6: {
			b: &ExponentialBackoff{
				Multiplier:   4,
				InitialDelay: time.Second * 1,
				MaxDelay:     time.Second * 10,
				Jitter:       1,
			},
			attempt:  11,
			expected: time.Second * 15,
		},
		7: {
			b: &ExponentialBackoff{
				Multiplier:   4,
				InitialDelay: time.Second * 1,
				MaxDelay:     time.Second * 10,
				Jitter:       1,
			},
			attempt:  111,
			expected: time.Second * 15,
		},
	}
	for i, tc := range cases {
		assert.Equal(t, tc.expected, tc.b.Backoff(tc.attempt), "test[%d]", i)
	}
}
