package models

type Interface interface {
	Add(items ...interface{})
	Remove(items ...interface{})
	Size() int
	Clear()
	IsEmpty() bool
	List() []interface{}
}

var keyExists = struct{}{}

type Set struct {
	m map[interface{}]struct{} // struct{} doesn't take up space
}

func (s Set) Add(items ...interface{}) {
	if len(items) == 0 {
		return
	}

	for _, item := range items {
		s.m[item] = keyExists
	}
}

func (s Set) Remove(items ...interface{}) {
	if len(items) == 0 {
		return
	}

	for _, item := range items {
		delete(s.m, item)
	}
}

func (s Set) Size() int {
	return len(s.m)
}

func (s Set) Clear() {
	s.m = make(map[interface{}]struct{})
}

func (s Set) IsEmpty() bool {
	return s.Size() == 0
}

func (s Set) List() []interface{} {
	list := make([]interface{}, 0, len(s.m))

	for item := range s.m {
		list = append(list, item)
	}
	return list
}

func New() Interface {
	s := Set{}
	s.m = make(map[interface{}]struct{})
	// Ensure interface compliance
	var _ Interface = s

	return s
}
