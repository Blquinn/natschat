package utils

import "sync"

type LazyString func() string

func MakeLazyString(f func() string) LazyString {
	var v string
	var once sync.Once
	return func() string {
		once.Do(func() {
			v = f()
			f = nil // so that f can now be GC'ed
		})
		return v
	}
}

type LazyByteArray func() []byte

func MakeLazyByteArray(f func() []byte) LazyByteArray {
	var v []byte
	var once sync.Once
	return func() []byte {
		once.Do(func() {
			v = f()
			f = nil // so that f can now be GC'ed
		})
		return v
	}
}
