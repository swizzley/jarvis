package collections

import (
	"testing"
	"time"

	"github.com/blendlabs/go-assert"
)

func TestQueue(t *testing.T) {
	a := assert.New(t)

	q := &Queue{}
	a.Nil(q.head)
	a.Nil(q.tail)
	a.Empty(q.ToArray())
	a.Nil(q.Dequeue())
	a.Equal(q.head, q.tail)
	a.Nil(q.Peek())
	a.Nil(q.PeekBack())
	a.Equal(0, q.Length())

	q.Push("foo")
	a.NotNil(q.head)
	a.Nil(q.head.Previous)
	a.NotNil(q.tail)
	a.Equal(q.head, q.tail)
	a.Equal(1, q.Length())
	a.Equal("foo", q.Peek())
	a.Equal("foo", q.PeekBack())

	q.Push("bar")
	a.NotNil(q.head)
	a.NotNil(q.head.Previous)
	a.Nil(q.head.Previous.Previous)
	a.Equal(q.head.Previous, q.tail)
	a.NotNil(q.tail)
	a.NotEqual(q.head, q.tail)
	a.Equal(2, q.Length())
	a.Equal("foo", q.Peek())
	a.Equal("bar", q.PeekBack())

	q.Push("baz")
	a.NotNil(q.head)
	a.NotNil(q.head.Previous)
	a.NotNil(q.head.Previous.Previous)
	a.Nil(q.head.Previous.Previous.Previous)
	a.Equal(q.head.Previous.Previous, q.tail)
	a.NotNil(q.tail)
	a.NotEqual(q.head, q.tail)
	a.Equal(3, q.Length())
	a.Equal("foo", q.Peek())
	a.Equal("baz", q.PeekBack())

	q.Push("fizz")
	a.NotNil(q.head)
	a.NotNil(q.head.Previous)
	a.NotNil(q.head.Previous.Previous)
	a.NotNil(q.head.Previous.Previous.Previous)
	a.Nil(q.head.Previous.Previous.Previous.Previous)
	a.Equal(q.head.Previous.Previous.Previous, q.tail)
	a.NotNil(q.tail)
	a.NotEqual(q.head, q.tail)
	a.Equal(4, q.Length())
	a.Equal("foo", q.Peek())
	a.Equal("fizz", q.PeekBack())

	values := q.ToArray()
	a.Len(values, 4)
	a.Equal("foo", values[0])
	a.Equal("bar", values[1])
	a.Equal("baz", values[2])
	a.Equal("fizz", values[3])

	shouldBeFoo := q.Dequeue()
	a.Equal("foo", shouldBeFoo)
	a.NotNil(q.head)
	a.NotNil(q.head.Previous)
	a.NotNil(q.head.Previous.Previous)
	a.Nil(q.head.Previous.Previous.Previous)
	a.Equal(q.head.Previous.Previous, q.tail)
	a.NotNil(q.tail)
	a.NotEqual(q.head, q.tail)
	a.Equal(3, q.Length())
	a.Equal("bar", q.Peek())
	a.Equal("fizz", q.PeekBack())

	shouldBeBar := q.Dequeue()
	a.Equal("bar", shouldBeBar)
	a.NotNil(q.head)
	a.NotNil(q.head.Previous)
	a.Nil(q.head.Previous.Previous)
	a.Equal(q.head.Previous, q.tail)
	a.NotNil(q.tail)
	a.NotEqual(q.head, q.tail)
	a.Equal(2, q.Length())
	a.Equal("baz", q.Peek())
	a.Equal("fizz", q.PeekBack())

	shouldBeBaz := q.Dequeue()
	a.Equal("baz", shouldBeBaz)
	a.NotNil(q.head)
	a.Nil(q.head.Previous)
	a.NotNil(q.tail)
	a.Equal(q.head, q.tail)
	a.Equal(1, q.Length())
	a.Equal("fizz", q.Peek())
	a.Equal("fizz", q.PeekBack())

	shouldBeFizz := q.Dequeue()
	a.Equal("fizz", shouldBeFizz)
	a.Nil(q.head)
	a.Nil(q.tail)
	a.Nil(q.Peek())
	a.Nil(q.PeekBack())
	a.Equal(0, q.Length())
}

func doQueueBenchmark(iterations int, b *testing.B) {
	for iteration := 0; iteration < iterations; iteration++ {
		q := &Queue{}
		for x := 0; x < 5000; x++ {
			q.Push(time.Now().UTC())
		}
		for x := 0; x < 5000; x++ {
			q.Dequeue()
		}
	}
}

func BenchmarkQueue(b *testing.B) {
	doQueueBenchmark(b.N, b)
}
