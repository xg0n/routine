package routine

import (
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunnable(t *testing.T) {
	count := 0
	var runnable Runnable = func() {
		count++
	}
	runnable()
	assert.Equal(t, 1, count)
	//
	var fun func() = runnable
	fun()
	assert.Equal(t, 2, count)
}

func TestCallable(t *testing.T) {
	var callable Callable = func() interface{} {
		return "Hello"
	}
	assert.Equal(t, "Hello", callable())
	//
	var fun func() any = callable
	assert.Equal(t, "Hello", fun())
}

func TestCancelRunnable(t *testing.T) {
	count := 0
	var cancelRunnable CancelRunnable = func(token CancelToken) {
		count++
	}
	cancelRunnable(nil)
	assert.Equal(t, 1, count)
	//
	var fun func(CancelToken) = cancelRunnable
	fun(nil)
	assert.Equal(t, 2, count)
}

func TestCancelCallable(t *testing.T) {
	var cancelCallable CancelCallable = func(token CancelToken) interface{} {
		return "Hello"
	}
	assert.Equal(t, "Hello", cancelCallable(nil))
	//
	var fun func(CancelToken) any = cancelCallable
	assert.Equal(t, "Hello", fun(nil))
}

func TestWrapTask_NoContext(t *testing.T) {
	run := false
	wrappedRun := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tls := NewThreadLocal()
	tlsInherit := NewInheritableThreadLocal()
	tls.Set("hello")
	tlsInherit.Set("inherit hello")
	assert.Equal(t, "hello", tls.Get())
	assert.Equal(t, "inherit hello", tlsInherit.Get())
	task := WrapTask(func() {
		assert.Nil(t, tls.Get())
		assert.Equal(t, "inherit hello", tlsInherit.Get())
		tls.Set("世界")
		tlsInherit.Set("inherit 世界")
		assert.Equal(t, "世界", tls.Get())
		assert.Equal(t, "inherit 世界", tlsInherit.Get())
		wrappedRun = true
	})
	tls.Set("world")
	tlsInherit.Set("inherit world")
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	go func() {
		task.Run()
		assert.Nil(t, tls.Get())
		assert.Nil(t, tlsInherit.Get())
		run = true
		wg.Done()
	}()
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	wg.Wait()
	assert.True(t, wrappedRun)
	assert.True(t, run)
}

func TestWrapTask_HasContext(t *testing.T) {
	run := false
	wrappedRun := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tls := NewThreadLocal()
	tlsInherit := NewInheritableThreadLocal()
	tls.Set("hello")
	tlsInherit.Set("inherit hello")
	assert.Equal(t, "hello", tls.Get())
	assert.Equal(t, "inherit hello", tlsInherit.Get())
	task := WrapTask(func() {
		assert.Nil(t, tls.Get())
		assert.Equal(t, "inherit hello", tlsInherit.Get())
		tls.Set("世界")
		tlsInherit.Set("inherit 世界")
		assert.Equal(t, "世界", tls.Get())
		assert.Equal(t, "inherit 世界", tlsInherit.Get())
		wrappedRun = true
	})
	tls.Set("world")
	tlsInherit.Set("inherit world")
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	go func() {
		tls.Set("你好")
		tlsInherit.Set("inherit 你好")
		task.Run()
		assert.Equal(t, "你好", tls.Get())
		assert.Equal(t, "inherit 你好", tlsInherit.Get())
		run = true
		wg.Done()
	}()
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	wg.Wait()
	assert.True(t, wrappedRun)
	assert.True(t, run)
}

func TestWrapTask_Complete_ThenFail(t *testing.T) {
	newStdout, oldStdout := captureStdout()
	defer restoreStdout(newStdout, oldStdout)
	//
	run := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg2 := &sync.WaitGroup{}
	wg2.Add(1)
	wg3 := &sync.WaitGroup{}
	wg3.Add(1)
	task := WrapTask(func() {
		wg.Done()  //1
		wg2.Wait() //4
		run = true
		wg3.Done() //5
		panic(1)
	})
	go task.Run()
	wg.Wait() //2
	task.Complete(nil)
	assert.Nil(t, task.Get())
	wg2.Done() //3
	wg3.Wait() //6
	assert.True(t, task.IsDone())
	assert.False(t, task.IsFailed())
	assert.False(t, task.IsCanceled())
	assert.True(t, run)
	//
	time.Sleep(10 * time.Millisecond)
	output := readAll(newStdout)
	assert.Equal(t, "", output)
}

func TestWrapWaitTask_NoContext(t *testing.T) {
	run := false
	wrappedRun := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tls := NewThreadLocal()
	tlsInherit := NewInheritableThreadLocal()
	tls.Set("hello")
	tlsInherit.Set("inherit hello")
	assert.Equal(t, "hello", tls.Get())
	assert.Equal(t, "inherit hello", tlsInherit.Get())
	task := WrapWaitTask(func(token CancelToken) {
		assert.Nil(t, tls.Get())
		assert.Equal(t, "inherit hello", tlsInherit.Get())
		tls.Set("世界")
		tlsInherit.Set("inherit 世界")
		assert.Equal(t, "世界", tls.Get())
		assert.Equal(t, "inherit 世界", tlsInherit.Get())
		wrappedRun = true
	})
	tls.Set("world")
	tlsInherit.Set("inherit world")
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	go func() {
		task.Run()
		assert.Nil(t, tls.Get())
		assert.Nil(t, tlsInherit.Get())
		run = true
		wg.Done()
	}()
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	assert.Nil(t, task.Get())
	assert.True(t, wrappedRun)
	wg.Wait()
	assert.True(t, wrappedRun)
	assert.True(t, run)
}

func TestWrapWaitTask_NoContext_Cancel(t *testing.T) {
	run := false
	wrappedRun := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tls := NewThreadLocal()
	tlsInherit := NewInheritableThreadLocal()
	tls.Set("hello")
	tlsInherit.Set("inherit hello")
	assert.Equal(t, "hello", tls.Get())
	assert.Equal(t, "inherit hello", tlsInherit.Get())
	task := WrapWaitTask(func(token CancelToken) {
		for i := 0; i < 1000; i++ {
			if token.IsCanceled() {
				return
			}
			time.Sleep(1 * time.Millisecond)
		}
		wrappedRun = true
	})
	tls.Set("world")
	tlsInherit.Set("inherit world")
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	go func() {
		task.Run()
		assert.Nil(t, tls.Get())
		assert.Nil(t, tlsInherit.Get())
		run = true
		wg.Done()
	}()
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	task.Cancel()
	assert.True(t, task.IsCanceled())
	assert.Panics(t, func() {
		task.Get()
	})
	assert.False(t, wrappedRun)
	wg.Wait()
	assert.False(t, wrappedRun)
	assert.True(t, run)
}

func TestWrapWaitTask_HasContext(t *testing.T) {
	run := false
	wrappedRun := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tls := NewThreadLocal()
	tlsInherit := NewInheritableThreadLocal()
	tls.Set("hello")
	tlsInherit.Set("inherit hello")
	assert.Equal(t, "hello", tls.Get())
	assert.Equal(t, "inherit hello", tlsInherit.Get())
	task := WrapWaitTask(func(token CancelToken) {
		assert.Nil(t, tls.Get())
		assert.Equal(t, "inherit hello", tlsInherit.Get())
		tls.Set("世界")
		tlsInherit.Set("inherit 世界")
		assert.Equal(t, "世界", tls.Get())
		assert.Equal(t, "inherit 世界", tlsInherit.Get())
		wrappedRun = true
	})
	tls.Set("world")
	tlsInherit.Set("inherit world")
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	go func() {
		tls.Set("你好")
		tlsInherit.Set("inherit 你好")
		task.Run()
		assert.Equal(t, "你好", tls.Get())
		assert.Equal(t, "inherit 你好", tlsInherit.Get())
		run = true
		wg.Done()
	}()
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	assert.Nil(t, task.Get())
	assert.True(t, wrappedRun)
	wg.Wait()
	assert.True(t, wrappedRun)
	assert.True(t, run)
}

func TestWrapWaitTask_HasContext_Cancel(t *testing.T) {
	run := false
	wrappedRun := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tls := NewThreadLocal()
	tlsInherit := NewInheritableThreadLocal()
	tls.Set("hello")
	tlsInherit.Set("inherit hello")
	assert.Equal(t, "hello", tls.Get())
	assert.Equal(t, "inherit hello", tlsInherit.Get())
	task := WrapWaitTask(func(token CancelToken) {
		for i := 0; i < 1000; i++ {
			if token.IsCanceled() {
				return
			}
			time.Sleep(1 * time.Millisecond)
		}
		wrappedRun = true
	})
	tls.Set("world")
	tlsInherit.Set("inherit world")
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	go func() {
		tls.Set("你好")
		tlsInherit.Set("inherit 你好")
		task.Run()
		assert.Equal(t, "你好", tls.Get())
		assert.Equal(t, "inherit 你好", tlsInherit.Get())
		run = true
		wg.Done()
	}()
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	task.Cancel()
	assert.True(t, task.IsCanceled())
	assert.Panics(t, func() {
		task.Get()
	})
	assert.False(t, wrappedRun)
	wg.Wait()
	assert.False(t, wrappedRun)
	assert.True(t, run)
}

func TestWrapWaitTask_Complete_ThenFail(t *testing.T) {
	newStdout, oldStdout := captureStdout()
	defer restoreStdout(newStdout, oldStdout)
	//
	run := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg2 := &sync.WaitGroup{}
	wg2.Add(1)
	wg3 := &sync.WaitGroup{}
	wg3.Add(1)
	task := WrapWaitTask(func(token CancelToken) {
		wg.Done()  //1
		wg2.Wait() //4
		run = true
		wg3.Done() //5
		panic(1)
	})
	go task.Run()
	wg.Wait() //2
	task.Complete(nil)
	assert.Nil(t, task.Get())
	wg2.Done() //3
	wg3.Wait() //6
	assert.True(t, task.IsDone())
	assert.False(t, task.IsFailed())
	assert.False(t, task.IsCanceled())
	assert.True(t, run)
	//
	time.Sleep(10 * time.Millisecond)
	output := readAll(newStdout)
	assert.Equal(t, "", output)
}

func TestWrapWaitResultTask_NoContext(t *testing.T) {
	run := false
	wrappedRun := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tls := NewThreadLocal()
	tlsInherit := NewInheritableThreadLocal()
	tls.Set("hello")
	tlsInherit.Set("inherit hello")
	assert.Equal(t, "hello", tls.Get())
	assert.Equal(t, "inherit hello", tlsInherit.Get())
	task := WrapWaitResultTask(func(token CancelToken) any {
		assert.Nil(t, tls.Get())
		assert.Equal(t, "inherit hello", tlsInherit.Get())
		tls.Set("世界")
		tlsInherit.Set("inherit 世界")
		assert.Equal(t, "世界", tls.Get())
		assert.Equal(t, "inherit 世界", tlsInherit.Get())
		wrappedRun = true
		return 1
	})
	tls.Set("world")
	tlsInherit.Set("inherit world")
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	go func() {
		task.Run()
		assert.Nil(t, tls.Get())
		assert.Nil(t, tlsInherit.Get())
		run = true
		wg.Done()
	}()
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	assert.Equal(t, 1, task.Get())
	assert.True(t, wrappedRun)
	wg.Wait()
	assert.True(t, wrappedRun)
	assert.True(t, run)
}

func TestWrapWaitResultTask_NoContext_Cancel(t *testing.T) {
	run := false
	wrappedRun := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tls := NewThreadLocal()
	tlsInherit := NewInheritableThreadLocal()
	tls.Set("hello")
	tlsInherit.Set("inherit hello")
	assert.Equal(t, "hello", tls.Get())
	assert.Equal(t, "inherit hello", tlsInherit.Get())
	task := WrapWaitResultTask(func(token CancelToken) any {
		for i := 0; i < 1000; i++ {
			if token.IsCanceled() {
				return 0
			}
			time.Sleep(1 * time.Millisecond)
		}
		wrappedRun = true
		return 1
	})
	tls.Set("world")
	tlsInherit.Set("inherit world")
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	go func() {
		task.Run()
		assert.Nil(t, tls.Get())
		assert.Nil(t, tlsInherit.Get())
		run = true
		wg.Done()
	}()
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	task.Cancel()
	assert.True(t, task.IsCanceled())
	assert.Panics(t, func() {
		task.Get()
	})
	assert.False(t, wrappedRun)
	wg.Wait()
	assert.False(t, wrappedRun)
	assert.True(t, run)
}

func TestWrapWaitResultTask_HasContext(t *testing.T) {
	run := false
	wrappedRun := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tls := NewThreadLocal()
	tlsInherit := NewInheritableThreadLocal()
	tls.Set("hello")
	tlsInherit.Set("inherit hello")
	assert.Equal(t, "hello", tls.Get())
	assert.Equal(t, "inherit hello", tlsInherit.Get())
	task := WrapWaitResultTask(func(token CancelToken) any {
		assert.Nil(t, tls.Get())
		assert.Equal(t, "inherit hello", tlsInherit.Get())
		tls.Set("世界")
		tlsInherit.Set("inherit 世界")
		assert.Equal(t, "世界", tls.Get())
		assert.Equal(t, "inherit 世界", tlsInherit.Get())
		wrappedRun = true
		return 1
	})
	tls.Set("world")
	tlsInherit.Set("inherit world")
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	go func() {
		tls.Set("你好")
		tlsInherit.Set("inherit 你好")
		task.Run()
		assert.Equal(t, "你好", tls.Get())
		assert.Equal(t, "inherit 你好", tlsInherit.Get())
		run = true
		wg.Done()
	}()
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	assert.Equal(t, 1, task.Get())
	assert.True(t, wrappedRun)
	wg.Wait()
	assert.True(t, wrappedRun)
	assert.True(t, run)
}

func TestWrapWaitResultTask_HasContext_Cancel(t *testing.T) {
	run := false
	wrappedRun := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	tls := NewThreadLocal()
	tlsInherit := NewInheritableThreadLocal()
	tls.Set("hello")
	tlsInherit.Set("inherit hello")
	assert.Equal(t, "hello", tls.Get())
	assert.Equal(t, "inherit hello", tlsInherit.Get())
	task := WrapWaitResultTask(func(token CancelToken) any {
		for i := 0; i < 1000; i++ {
			if token.IsCanceled() {
				return 0
			}
			time.Sleep(1 * time.Millisecond)
		}
		wrappedRun = true
		return 1
	})
	tls.Set("world")
	tlsInherit.Set("inherit world")
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	go func() {
		tls.Set("你好")
		tlsInherit.Set("inherit 你好")
		task.Run()
		assert.Equal(t, "你好", tls.Get())
		assert.Equal(t, "inherit 你好", tlsInherit.Get())
		run = true
		wg.Done()
	}()
	assert.Equal(t, "world", tls.Get())
	assert.Equal(t, "inherit world", tlsInherit.Get())
	task.Cancel()
	assert.True(t, task.IsCanceled())
	assert.Panics(t, func() {
		task.Get()
	})
	assert.False(t, wrappedRun)
	wg.Wait()
	assert.False(t, wrappedRun)
	assert.True(t, run)
}

func TestWrapWaitResultTask_Complete_ThenFail(t *testing.T) {
	newStdout, oldStdout := captureStdout()
	defer restoreStdout(newStdout, oldStdout)
	//
	run := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg2 := &sync.WaitGroup{}
	wg2.Add(1)
	wg3 := &sync.WaitGroup{}
	wg3.Add(1)
	task := WrapWaitResultTask(func(token CancelToken) any {
		wg.Done()  //1
		wg2.Wait() //4
		run = true
		wg3.Done() //5
		panic(1)
	})
	go task.Run()
	wg.Wait() //2
	task.Complete(nil)
	assert.Nil(t, task.Get())
	wg2.Done() //3
	wg3.Wait() //6
	assert.True(t, task.IsDone())
	assert.False(t, task.IsFailed())
	assert.False(t, task.IsCanceled())
	assert.True(t, run)
	//
	time.Sleep(10 * time.Millisecond)
	output := readAll(newStdout)
	assert.Equal(t, "", output)
}

func TestGo_Error(t *testing.T) {
	newStdout, oldStdout := captureStdout()
	defer restoreStdout(newStdout, oldStdout)
	//
	run := false
	assert.NotPanics(t, func() {
		wg := &sync.WaitGroup{}
		wg.Add(1)
		Go(func() {
			run = true
			wg.Done()
			panic("error")
		})
		wg.Wait()
	})
	assert.True(t, run)
	//
	time.Sleep(10 * time.Millisecond)
	output := readAll(newStdout)
	lines := strings.Split(output, newLine)
	assert.Equal(t, 7, len(lines))
	//
	line := lines[0]
	assert.Equal(t, "RuntimeError: error", line)
	//
	line = lines[1]
	assert.True(t, strings.HasPrefix(line, "   at github.com/timandy/routine.TestGo_Error."))
	assert.True(t, strings.HasSuffix(line, "api_routine_test.go:601"))
	//
	line = lines[2]
	assert.True(t, strings.HasPrefix(line, "   at github.com/timandy/routine.WrapTask."))
	assert.True(t, strings.HasSuffix(line, "api_routine.go:45"))
	//
	line = lines[3]
	assert.True(t, strings.HasPrefix(line, "   at github.com/timandy/routine.(*futureTask).Run()"))
	assert.True(t, strings.HasSuffix(line, "future_task.go:108"))
	//
	line = lines[4]
	assert.Equal(t, "   --- End of error stack trace ---", line)
	//
	line = lines[5]
	assert.True(t, strings.HasPrefix(line, "   created by github.com/timandy/routine.Go()"))
	assert.True(t, strings.HasSuffix(line, "api_routine.go:151"))
	//
	line = lines[6]
	assert.Equal(t, "", line)
}

func TestGo_Nil(t *testing.T) {
	assert.Nil(t, createInheritedMap())
	//
	run := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	Go(func() {
		assert.Nil(t, createInheritedMap())
		run = true
		wg.Done()
	})
	wg.Wait()
	assert.True(t, run)
}

func TestGo_Value(t *testing.T) {
	tls := NewThreadLocal()
	tls.Set("Hello")
	assert.Equal(t, "Hello", tls.Get())
	//
	inheritableTls := NewInheritableThreadLocal()
	inheritableTls.Set("World")
	assert.Equal(t, "World", inheritableTls.Get())
	//
	assert.NotNil(t, createInheritedMap())
	//
	run := false
	wg := &sync.WaitGroup{}
	wg.Add(1)
	Go(func() {
		assert.NotNil(t, createInheritedMap())
		//
		assert.Nil(t, tls.Get())
		assert.Equal(t, "World", inheritableTls.Get())
		//
		tls.Set("Hello2")
		assert.Equal(t, "Hello2", tls.Get())
		//
		inheritableTls.Remove()
		assert.Nil(t, inheritableTls.Get())
		//
		run = true
		wg.Done()
	})
	wg.Wait()
	assert.True(t, run)
	//
	assert.Equal(t, "Hello", tls.Get())
	assert.Equal(t, "World", inheritableTls.Get())
}

func TestGo_Cross(t *testing.T) {
	tls := NewThreadLocal()
	tls.Set("Hello")
	assert.Equal(t, "Hello", tls.Get())
	//
	wg := &sync.WaitGroup{}
	wg.Add(1)
	Go(func() {
		assert.Nil(t, tls.Get())
		wg.Done()
	})
	wg.Wait()
}

func TestGoWait_Error(t *testing.T) {
	run := false
	assert.Panics(t, func() {
		task := GoWait(func(token CancelToken) {
			run = true
			panic("error")
		})
		task.Get()
	})
	assert.True(t, run)
}

func TestGoWait_Nil(t *testing.T) {
	assert.Nil(t, createInheritedMap())
	//
	run := false
	task := GoWait(func(token CancelToken) {
		assert.Nil(t, createInheritedMap())
		run = true
	})
	assert.Nil(t, task.Get())
	assert.True(t, run)
}

func TestGoWait_Value(t *testing.T) {
	tls := NewThreadLocal()
	tls.Set("Hello")
	assert.Equal(t, "Hello", tls.Get())
	//
	inheritableTls := NewInheritableThreadLocal()
	inheritableTls.Set("World")
	assert.Equal(t, "World", inheritableTls.Get())
	//
	assert.NotNil(t, createInheritedMap())
	//
	run := false
	task := GoWait(func(token CancelToken) {
		assert.NotNil(t, createInheritedMap())
		//
		assert.Nil(t, tls.Get())
		assert.Equal(t, "World", inheritableTls.Get())
		//
		tls.Set("Hello2")
		assert.Equal(t, "Hello2", tls.Get())
		//
		inheritableTls.Remove()
		assert.Nil(t, inheritableTls.Get())
		//
		run = true
	})
	assert.Nil(t, task.Get())
	assert.True(t, run)
	//
	assert.Equal(t, "Hello", tls.Get())
	assert.Equal(t, "World", inheritableTls.Get())
}

func TestGoWait_Cross(t *testing.T) {
	tls := NewThreadLocal()
	tls.Set("Hello")
	assert.Equal(t, "Hello", tls.Get())
	//
	GoWait(func(token CancelToken) {
		assert.Nil(t, tls.Get())
	}).Get()
}

func TestGoWaitResult_Error(t *testing.T) {
	run := false
	assert.Panics(t, func() {
		task := GoWaitResult(func(token CancelToken) any {
			run = true
			if run {
				panic("error")
			}
			return 1
		})
		task.Get()
	})
	assert.True(t, run)
}

func TestGoWaitResult_Nil(t *testing.T) {
	assert.Nil(t, createInheritedMap())
	//
	run := false
	task := GoWaitResult(func(token CancelToken) any {
		assert.Nil(t, createInheritedMap())
		run = true
		return true
	})
	assert.True(t, task.Get().(bool))
	assert.True(t, run)
}

func TestGoWaitResult_Value(t *testing.T) {
	tls := NewThreadLocal()
	tls.Set("Hello")
	assert.Equal(t, "Hello", tls.Get())
	//
	inheritableTls := NewInheritableThreadLocal()
	inheritableTls.Set("World")
	assert.Equal(t, "World", inheritableTls.Get())
	//
	assert.NotNil(t, createInheritedMap())
	//
	run := false
	task := GoWaitResult(func(token CancelToken) any {
		assert.NotNil(t, createInheritedMap())
		//
		assert.Nil(t, tls.Get())
		assert.Equal(t, "World", inheritableTls.Get())
		//
		tls.Set("Hello2")
		assert.Equal(t, "Hello2", tls.Get())
		//
		inheritableTls.Remove()
		assert.Nil(t, inheritableTls.Get())
		//
		run = true
		return true
	})
	assert.True(t, task.Get().(bool))
	assert.True(t, run)
	//
	assert.Equal(t, "Hello", tls.Get())
	assert.Equal(t, "World", inheritableTls.Get())
}

func TestGoWaitResult_Cross(t *testing.T) {
	tls := NewThreadLocal()
	tls.Set("Hello")
	assert.Equal(t, "Hello", tls.Get())
	//
	result := GoWaitResult(func(token CancelToken) any {
		assert.Nil(t, tls.Get())
		return tls.Get()
	}).Get()
	assert.Nil(t, result)
}

func captureStdout() (newStdout, oldStdout *os.File) {
	oldStdout = os.Stdout
	fileName := path.Join(os.TempDir(), "go_test_"+strconv.FormatInt(time.Now().UnixNano(), 10)+".txt")
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	os.Stdout = file
	newStdout = file
	return
}

func restoreStdout(newStdout, oldStdout *os.File) {
	os.Stdout = oldStdout
	if err := newStdout.Close(); err != nil {
		panic(err)
	}
	if err := os.Remove(newStdout.Name()); err != nil {
		panic(err)
	}
}

func readAll(rs io.ReadSeeker) string {
	if _, err := rs.Seek(0, io.SeekStart); err != nil {
		panic(err)
	}
	b := make([]byte, 0, 512)
	for {
		if len(b) == cap(b) {
			b = append(b, 0)[:len(b)]
		}
		n, err := rs.Read(b[len(b):cap(b)])
		b = b[:len(b)+n]
		if err != nil {
			if err == io.EOF {
				return string(b)
			}
			panic(err)
		}
	}
}
