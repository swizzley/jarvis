package collections

func NewIntSet(values []int) IntSet {
	is := IntSet{}
	for _, v := range values {
		is[v] = len(is)
	}
	return is
}

type IntSet map[int]int

func (is IntSet) Add(i int) {
	is[i] = len(is)
}

func (is IntSet) Remove(i int) {
	delete(is, i)
}

func (is IntSet) Contains(i int) bool {
	_, ok := is[i]
	return ok
}
