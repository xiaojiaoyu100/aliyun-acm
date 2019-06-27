package aliacm

// Setter configures the diamond.
type Setter func(d *Diamond) error

// WithLongPullRate sets long pull rate.
func WithLongPullRate(rate int) Setter {
	return func(d *Diamond) error {
		d.longPullRate = rate
		return nil
	}
}
