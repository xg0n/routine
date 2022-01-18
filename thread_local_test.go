package routine

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestThreadLocal_Id(t *testing.T) {
	threadLocal := NewThreadLocal()
	assert.GreaterOrEqual(t, threadLocal.Id(), 0)
	threadLocal2 := NewThreadLocalWithInitial(func() Any {
		return "Hello"
	})
	assert.GreaterOrEqual(t, threadLocal2.Id(), 0)
	assert.NotEqual(t, threadLocal.Id(), threadLocal2)
}

func TestThreadLocal(t *testing.T) {
	threadLocal := NewThreadLocal()
	threadLocal2 := NewThreadLocal()
	assert.Nil(t, threadLocal.Get())
	assert.Nil(t, threadLocal2.Get())
	//
	threadLocal.Set(1)
	threadLocal2.Set("World")
	assert.Equal(t, 1, threadLocal.Get())
	assert.Equal(t, "World", threadLocal2.Get())
	//
	threadLocal.Set(nil)
	threadLocal2.Set(nil)
	assert.Nil(t, threadLocal.Get())
	assert.Nil(t, threadLocal2.Get())
	//
	threadLocal.Set(2)
	threadLocal2.Set("!")
	assert.Equal(t, 2, threadLocal.Get())
	assert.Equal(t, "!", threadLocal2.Get())
	//
	threadLocal.Remove()
	threadLocal2.Remove()
	assert.Nil(t, threadLocal.Get())
	assert.Nil(t, threadLocal2.Get())
	//
	threadLocal.Set(2)
	threadLocal2.Set("!")
	assert.Equal(t, 2, threadLocal.Get())
	assert.Equal(t, "!", threadLocal2.Get())
	waiter := &sync.WaitGroup{}
	waiter.Add(100)
	for i := 0; i < 100; i++ {
		Go(func() {
			assert.Nil(t, threadLocal.Get())
			assert.Nil(t, threadLocal2.Get())
			waiter.Done()
		})
	}
	waiter.Wait()
	assert.Equal(t, 2, threadLocal.Get())
	assert.Equal(t, "!", threadLocal2.Get())
}

func TestThreadLocalMixed(t *testing.T) {
	threadLocal := NewThreadLocal()
	threadLocal2 := NewThreadLocalWithInitial(func() Any {
		return "Hello"
	})
	assert.Nil(t, threadLocal.Get())
	assert.Equal(t, "Hello", threadLocal2.Get())
	//
	threadLocal.Set(1)
	threadLocal2.Set("World")
	assert.Equal(t, 1, threadLocal.Get())
	assert.Equal(t, "World", threadLocal2.Get())
	//
	threadLocal.Set(nil)
	threadLocal2.Set(nil)
	assert.Nil(t, threadLocal.Get())
	assert.Nil(t, threadLocal2.Get())
	//
	threadLocal.Set(2)
	threadLocal2.Set("!")
	assert.Equal(t, 2, threadLocal.Get())
	assert.Equal(t, "!", threadLocal2.Get())
	//
	threadLocal.Remove()
	threadLocal2.Remove()
	assert.Nil(t, threadLocal.Get())
	assert.Equal(t, "Hello", threadLocal2.Get())
	//
	threadLocal.Set(2)
	threadLocal2.Set("!")
	assert.Equal(t, 2, threadLocal.Get())
	assert.Equal(t, "!", threadLocal2.Get())
	waiter := &sync.WaitGroup{}
	waiter.Add(100)
	for i := 0; i < 100; i++ {
		Go(func() {
			assert.Nil(t, threadLocal.Get())
			assert.Equal(t, "Hello", threadLocal2.Get())
			waiter.Done()
		})
	}
	waiter.Wait()
	assert.Equal(t, 2, threadLocal.Get())
	assert.Equal(t, "!", threadLocal2.Get())
}

func TestThreadLocalWithInitial(t *testing.T) {
	src := &person{Id: 1, Name: "Tim"}
	threadLocal := NewThreadLocalWithInitial(nil)
	threadLocal2 := NewThreadLocalWithInitial(func() Any {
		return nil
	})
	threadLocal3 := NewThreadLocalWithInitial(func() Any {
		return src
	})
	threadLocal4 := NewThreadLocalWithInitial(func() Any {
		return *src
	})

	for i := 0; i < 100; i++ {
		p := threadLocal.Get()
		assert.Nil(t, p)
		//
		p2 := threadLocal2.Get()
		assert.Nil(t, p2)
		//
		p3 := threadLocal3.Get().(*person)
		assert.Same(t, src, p3)

		p4 := threadLocal4.Get().(person)
		assert.NotSame(t, src, &p4)
		assert.Equal(t, *src, p4)

		waiter := &sync.WaitGroup{}
		waiter.Add(1)
		Go(func() {
			assert.Same(t, src, threadLocal3.Get().(*person))
			p5 := threadLocal4.Get().(person)
			assert.NotSame(t, src, &p5)
			assert.Equal(t, *src, p5)
			//
			waiter.Done()
		})
		waiter.Wait()
	}

	threadLocal3.Set(nil)
	threadLocal4.Set(nil)
	assert.Nil(t, threadLocal3.Get())
	assert.Nil(t, threadLocal4.Get())

	threadLocal3.Remove()
	threadLocal4.Remove()
	assert.Same(t, src, threadLocal3.Get().(*person))
	p6 := threadLocal4.Get().(person)
	assert.NotSame(t, src, &p6)
	assert.Equal(t, *src, p6)
}

type person struct {
	Id   int
	Name string
}
