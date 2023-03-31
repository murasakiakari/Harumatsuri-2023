package utility

import "sync"

type Pool[T any] struct {
	New func() any

	internal sync.Pool
	once     sync.Once
}

func (pool *Pool[T]) Get() *T {
	pool.once.Do(func() {
		pool.internal = sync.Pool{New: pool.New}
	})
	return pool.internal.Get().(*T)
}

func (pool *Pool[T]) Put(t *T) {
	pool.internal.Put(t)
}
