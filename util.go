package neat

import (
	"bytes"
	"fmt"
)

type Queue interface {
	Push(interface{})
	Pop() interface{}
	Size() int
	String() string
}

type squeue struct {
	q []interface{}
}

func newsqueue() *squeue {
	return &squeue{make([]interface{}, 0)}
}

func (q *squeue) Push(item interface{}) {
	q.q = append(q.q, item)
}

func (q *squeue) Pop() interface{} {
	item := q.q[0]
	q.q = q.q[1:]

	return item
}

func (q *squeue) Size() int {
	return len(q.q)
}

func (q *squeue) String() string {
	var buf bytes.Buffer

	for i := 0; i < len(q.q); i++ {

		if i == 0 {
			buf.WriteString(fmt.Sprintf("%s", q.q[i]))
		} else {
			buf.WriteString(fmt.Sprintf(" <- %s", q.q[i]))
		}
	}

	return buf.String()
}

func max(a, b int) int {
	if a < b {
		return b
	}

	return a
}

func inRange(x, lower, upper float64) bool {
	return lower <= x && x <= upper 
}
