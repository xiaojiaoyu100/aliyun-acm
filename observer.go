package aliacm

// Observer observes the config change.
type Observer interface {
	Modify(unit Unit, config Config)
}
