package idx

type Set[T Type] struct {
	l List[T]
	m map[string]ID[T]
}

func NewSet[T Type](id ...ID[T]) *Set[T] {
	s := &Set[T]{
		m: make(map[string]ID[T]),
	}
	s.Add(id...)
	return s
}

func (s *Set[T]) Has(id ...ID[T]) bool {
	if s == nil || s.m == nil {
		return false
	}
	for _, i := range id {
		if i.nid == nil {
			continue
		}
		if _, ok := s.m[i.String()]; ok {
			return true
		}
	}
	return false
}

func (s *Set[T]) List() List[T] {
	if s == nil {
		return nil
	}
	return s.l.Clone()
}

func (s *Set[T]) Clone() *Set[T] {
	if s == nil {
		return nil
	}

	ns := &Set[T]{
		m: make(map[string]ID[T], len(s.m)),
		l: s.l.Clone(),
	}
	for k, v := range s.m {
		ns.m[k] = v
	}

	return ns
}

func (s *Set[T]) Add(id ...ID[T]) {
	if s == nil {
		return
	}
	for _, i := range id {
		if i.nid == nil {
			continue
		}
		str := i.String()
		if _, ok := s.m[str]; !ok {
			if s.m == nil {
				s.m = map[string]ID[T]{}
			}
			s.m[str] = i
			s.l = append(s.l, i)
		}
	}
}

func (s *Set[T]) Merge(sets ...*Set[T]) {
	if s == nil {
		return
	}
	for _, t := range sets {
		if t != nil {
			s.Add(t.l...)
		}
	}
}

func (s *Set[T]) Concat(sets ...*Set[T]) *Set[T] {
	if s == nil {
		return nil
	}
	ns := s.Clone()
	ns.Merge(sets...)
	return ns
}

func (s *Set[T]) Delete(id ...ID[T]) {
	if s == nil {
		return
	}
	for _, i := range id {
		if i.nid == nil {
			continue
		}
		s.l = s.l.Delete(i)
		if s.m != nil {
			delete(s.m, i.String())
		}
	}
}
