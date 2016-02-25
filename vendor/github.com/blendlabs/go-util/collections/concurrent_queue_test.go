package collections

import (
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
)

func TestConcurrentQueue(t *testing.T) {
	a := assert.New(t)

	q := NewConcurrentQueue(4)
	a.Empty(q.ToArray())
	a.Nil(q.Dequeue())
	a.Equal(0, q.Length())

	q.Push("foo")
	a.Equal(1, q.Length())

	q.Push("bar")
	a.Equal(2, q.Length())

	q.Push("baz")
	a.Equal(3, q.Length())

	q.Push("fizz")
	a.Equal(4, q.Length())

	values := q.ToArray()
	a.Len(values, 4)
	a.Equal("foo", values[0])
	a.Equal("bar", values[1])
	a.Equal("baz", values[2])
	a.Equal("fizz", values[3])

	shouldBeFoo := q.Dequeue()
	a.Equal("foo", shouldBeFoo)
	a.Equal(3, q.Length())

	shouldBeBar := q.Dequeue()
	a.Equal("bar", shouldBeBar)
	a.Equal(2, q.Length())

	shouldBeBaz := q.Dequeue()
	a.Equal("baz", shouldBeBaz)
	a.Equal(1, q.Length())

	shouldBeFizz := q.Dequeue()
	a.Equal("fizz", shouldBeFizz)
	a.Equal(0, q.Length())
}

func doConcurrentQueueBenchmark(iterations int, b *testing.B) {
	for iteration := 0; iteration < iterations; iteration++ {
		q := NewConcurrentQueue(5000)
		for x := 0; x < 5000; x++ {
			q.Push(time.Now().UTC())
		}
		for x := 0; x < 5000; x++ {
			q.Dequeue()
		}
	}
}

func BenchmarkConcurrentQueue(b *testing.B) {
	doConcurrentQueueBenchmark(b.N, b)
}
