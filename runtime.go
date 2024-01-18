package routine

import (
	"reflect"
	"unsafe"

	_ "github.com/xg0n/routine/g"
)

// getgp returns the pointer to the current runtime.g.
//
//go:linkname getgp github.com/xg0n/routine/g.getgp
func getgp() unsafe.Pointer

// getg0 returns the value of runtime.g0.
//
//go:linkname getg0 github.com/xg0n/routine/g.getg0
func getg0() interface{}

// getgt returns the type of runtime.g.
//
//go:linkname getgt github.com/xg0n/routine/g.getgt
func getgt() reflect.Type
