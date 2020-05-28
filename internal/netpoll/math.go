package netpoll

const (
	bitsize = 32 << (^uint(0) >> 63)
	//maxint        = int(1<<(bitsize-1) - 1)
	maxintHeadBit = 1 << (bitsize - 2)

	maximumCapacity = 1 << 30
)

// TableSizeFor .
func TableSizeFor(cap int) int {
	n := cap - 1
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	if n < 0 {
		return 1
	}
	if n >= maximumCapacity {
		return maximumCapacity
	}
	return n + 1
}

// CeilToPowerOfTwo returns the least power of two integer value greater than
// or equal to n.
func CeilToPowerOfTwo(n int) int {
	if n&maxintHeadBit != 0 && n > maxintHeadBit {
		panic("argument is too large")
	}
	if n <= 2 {
		return n
	}
	//n--
	//n = fillBits(n)
	//n++
	n = TableSizeFor(n)
	return n
}

func fillBits(n int) int {
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n |= n >> 32
	return n
}
