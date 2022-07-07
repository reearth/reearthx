package util

import "net/url"

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func IsZero[T comparable](v T) bool {
	var z T
	return v == z
}

func IsNotZero[T comparable](v T) bool {
	return !IsZero(v)
}

func Deref[T any](r *T) T {
	if r == nil {
		var z T
		return z
	}
	return *r
}

func DerefOr[T any](ref *T, def T) T {
	if ref == nil {
		return def
	}
	return *ref
}

func Unwrap[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func CloneRef[T any](r *T) *T {
	if r == nil {
		return nil
	}
	r2 := *r
	return &r2
}

func CopyURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	v := CloneRef(u)
	v.User = CloneRef(u.User)
	return v
}

// DR discard right
func DR[A, B any](a A, _ B) A {
	return a
}
