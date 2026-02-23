package testutil

import "time"

// DefaultTime is a fixed reference time used across all factories.
var DefaultTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

// StringPtr returns a pointer to s.
func StringPtr(s string) *string { return &s }

// TimePtr returns a pointer to t.
func TimePtr(t time.Time) *time.Time { return &t }

// Float64Ptr returns a pointer to f.
func Float64Ptr(f float64) *float64 { return &f }

// IntPtr returns a pointer to i.
func IntPtr(i int) *int { return &i }
