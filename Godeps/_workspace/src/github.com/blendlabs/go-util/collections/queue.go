package collections

type queueNode struct {
	Next     *queueNode
	Previous *queueNode
	Value    interface{}
}

type Queue struct {
	head   *queueNode
	tail   *queueNode
	length int
}

func (q *Queue) Length() int {
	return q.length
}

func (q *Queue) Push(value interface{}) {
	node := queueNode{Value: value}

	//the queue is empty, that is to say head is nil
	if q.head == nil {
		q.head = &node
		q.tail = &node
	} else {
		//the queue is not empty, we have a (valid) tail pointer
		//make
		//	- tail.prev point to new node
		//	- new node.next point to tail
		//	- tail to new node

		q.tail.Previous = &node
		node.Next = q.tail
		q.tail = &node
	}

	q.length = q.length + 1
}

func (q *Queue) Dequeue() interface{} {
	if q.head == nil {
		return nil
	}

	headValue := q.head.Value

	if q.length == 1 && q.head == q.tail {
		q.head = nil
		q.tail = nil
	} else if q.length == 1 {
		panic("Inconsistent queue state; length is 1 but head != tail")
	} else {
		q.head = q.head.Previous
		if q.head != nil {
			q.head.Next = nil
		}
	}

	q.length = q.length - 1
	return headValue
}

func (q *Queue) Peek() interface{} {
	if q.head == nil {
		return nil
	}
	return q.head.Value
}

func (q *Queue) PeekBack() interface{} {
	if q.tail == nil {
		return nil
	}
	return q.tail.Value
}

func (q *Queue) ToArray() []interface{} {
	if q.head == nil {
		return []interface{}{}
	}

	values := []interface{}{}
	nodePtr := q.head
	for nodePtr != nil {
		values = append(values, nodePtr.Value)
		nodePtr = nodePtr.Previous
	}
	return values
}
