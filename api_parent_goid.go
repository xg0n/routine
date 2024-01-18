package routine

// ParentGoid return the parent goroutine's unique id.
func ParentGoid() uint64 {
	return getg().parentGoid
}
