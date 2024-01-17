package routine

// Goid return the current goroutine's unique id.
func ParentGoid() uint64 {
	return getg().parentGoid
}
