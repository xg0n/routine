package routine

type threadLocalImpl struct {
	id       int
	supplier func() interface{}
}

func (tls *threadLocalImpl) Id() int {
	return tls.id
}

func (tls *threadLocalImpl) Get() interface{} {
	t := currentThread(true)
	mp := tls.getMap(t)
	if mp != nil {
		e := mp.getEntry(tls)
		if e != nil {
			return e.value
		}
	}
	return tls.setInitialValue(t)
}

func (tls *threadLocalImpl) Set(value interface{}) {
	t := currentThread(true)
	mp := tls.getMap(t)
	if mp != nil {
		mp.set(tls, value)
	} else {
		tls.createMap(t, value)
	}
}

func (tls *threadLocalImpl) Remove() {
	t := currentThread(true)
	mp := tls.getMap(t)
	if mp != nil {
		mp.remove(tls)
	}
}

func (tls *threadLocalImpl) getMap(t *thread) *threadLocalMap {
	return t.threadLocals
}

func (tls *threadLocalImpl) createMap(t *thread, firstValue interface{}) {
	mp := &threadLocalMap{}
	mp.set(tls, firstValue)
	t.threadLocals = mp
}

func (tls *threadLocalImpl) setInitialValue(t *thread) interface{} {
	value := tls.initialValue()
	mp := tls.getMap(t)
	if mp != nil {
		mp.set(tls, value)
	} else {
		tls.createMap(t, value)
	}
	return value
}

func (tls *threadLocalImpl) initialValue() interface{} {
	if tls.supplier == nil {
		return nil
	}
	return tls.supplier()
}
