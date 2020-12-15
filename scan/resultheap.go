package scan

type resultHeap []Result

func (r resultHeap) Len() int {
	return len(r)
}

func (r resultHeap) Less(i, j int) bool {
	// we want the highest score so we use greater-than
	return r[i].Score > r[j].Score
}

func (r resultHeap) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r resultHeap) Push(x interface{}) {
	item := x.(Result)
	r = append(r, item)
}

func (r resultHeap) Pop() interface{} {
	old := r
	n := len(old)
	item := old[n-1]
	r = old[:n-1]
	return item
}
