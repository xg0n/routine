package routine

import (
	"fmt"
	"sync/atomic"
)

// ThreadLocal provides goroutine-local variables.
type ThreadLocal interface {
	// Id returns the global id of instance
	Id() int

	// Get returns the value in the current goroutine's local threadLocalImpl, if it was set before.
	Get() interface{}

	// Set copy the value into the current goroutine's local threadLocalImpl, and return the old value.
	Set(value interface{})

	// Remove delete the value from the current goroutine's local threadLocalImpl, and return it.
	Remove()
}

// Go starts a new goroutine, and copy all local table from current goroutine.
func Go(f func()) {
	// backup
	copied := createInheritedMap()
	go func() {
		// catch
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		// restore
		t := currentThread(copied != nil)
		if t == nil {
			f()
		} else {
			backup := t.inheritableThreadLocals
			t.inheritableThreadLocals = copied
			f()
			t.inheritableThreadLocals = backup
		}
	}()
}

var threadLocalIndex int32 = -1

// NewThreadLocal create and return a new ThreadLocal instance.
func NewThreadLocal() ThreadLocal {
	return &threadLocalImpl{id: int(atomic.AddInt32(&threadLocalIndex, 1))}
}

// NewThreadLocalWithInitial create and return a new ThreadLocal instance. The initial value is determined by invoking the supplier method.
func NewThreadLocalWithInitial(supplier func() interface{}) ThreadLocal {
	return &threadLocalImpl{id: int(atomic.AddInt32(&threadLocalIndex, 1)), supplier: supplier}
}

var inheritableThreadLocalIndex int32 = -1

// NewInheritableThreadLocal create and return a new ThreadLocal instance.
func NewInheritableThreadLocal() ThreadLocal {
	return &inheritableThreadLocalImpl{id: int(atomic.AddInt32(&inheritableThreadLocalIndex, 1))}
}

// NewInheritableThreadLocalWithInitial create and return a new ThreadLocal instance. The initial value is determined by invoking the supplier method.
func NewInheritableThreadLocalWithInitial(supplier func() interface{}) ThreadLocal {
	return &inheritableThreadLocalImpl{id: int(atomic.AddInt32(&inheritableThreadLocalIndex, 1)), supplier: supplier}
}

// Goid return the current goroutine's unique id.
// It will try get gid by native cgo/asm for better performance,
// and could parse gid from stack for failover supporting.
func Goid() int64 {
	if goid, success := getGoidByNative(); success {
		return goid
	}
	return getGoidByStack()
}

// AllGoids return all goroutine's goid in the current golang process.
// It will try load all goid from runtime natively for better performance,
// and fallover to runtime.Stack, which is realy inefficient.
func AllGoids() []int64 {
	if goids, err := getAllGoidByNative(); err == nil {
		return goids
	}
	fmt.Println("[WARNING] cannot get all goids from runtime natively, now fall over to stack info, this will be very inefficient!!!")
	return getAllGoidByStack()
}
