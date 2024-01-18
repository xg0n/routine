package routine

// GetRoutineIds return current and parent go routine ids in one call.
func GetRoutineIds() (int64, uint64) {
	runtime_g := getg()
	return runtime_g.goid, runtime_g.parentGoid
}
