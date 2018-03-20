package neat

import "testing"

func TestSQueue(t *testing.T) {
	q := newsqueue()

	limit := 10
	for i := 0; i < limit; i++ {
		q.Push(i)
	}

	if q.Size() != limit {
		t.Error("Wrong size")
	}

	for i := 0; i < limit; i++ {
		x := q.Pop().(int)

		if x != i {
			t.Error("Expected ", i, " not ", x)
		}
	}

	if q.Size() != 0 {
		t.Error("Wrong size")
	}
}
