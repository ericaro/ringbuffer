package ringbuffer

import (
	"fmt"
	"testing"
)

func ExampleRing_AddAll() {
	buf := New(5)
	buf.AddAll(1, 2, 3) //fill the buffer
}
func ExampleRing_Remove() {
	buf := New(5)
	buf.AddAll(1, 2, 3) //fill the buffer
	buf.Remove(2)
	oldest, _ := buf.Get(-1)
	fmt.Printf("%v\n", oldest)
	//Output: 3
}

func ExampleRing_Get() {
	buf := New(5)
	buf.AddAll(1, 2, 3)      //fill the buffer
	latest, _ := buf.Get(0)  //get the oldest
	oldest, _ := buf.Get(-1) //get the oldest
	fmt.Printf("%v %v\n", oldest, latest)
	//Output: 1 3
}

//TestIndex because this is the main function
func TestIndex(t *testing.T) {

	var latest, length, capacity int
	capacity = 10
	length = 5
	latest = 5

	// buffer   indexes   0 1 2 3 4 5 6 7 8 9
	// circular indexes   x 4 3 2 1 0 x x x x
	// therefore the pos results

	assertPos(t, 0, 5, Index(0, latest, length, capacity))
	assertPos(t, 1, 4, Index(1, latest, length, capacity))
	assertPos(t, 2, 3, Index(2, latest, length, capacity))
	assertPos(t, 3, 2, Index(3, latest, length, capacity))
	assertPos(t, 4, 1, Index(4, latest, length, capacity))
	//also test the neg capa
	assertPos(t, -1, 1, Index(-1, latest, length, capacity))
	assertPos(t, -2, 2, Index(-2, latest, length, capacity))
	//also test the big neg capa
	assertPos(t, -1-10*length, 1, Index(-1-10*length, latest, length, capacity))
	//also test the overflow
	assertPos(t, length, 5, Index(length, latest, length, capacity))
	assertPos(t, length*10, 5, Index(length*10, latest, length, capacity))
	//

	//let's play with a more complicated example, where the circular buffer overlaps
	//
	// buffer   indexes   0 1 2 3 4 5 6 7 8 9
	// circular indexes   2 1 0 x x x x x 4 3
	latest = 2
	assertPos(t, 0, 2, Index(0, latest, length, capacity))
	assertPos(t, 1, 1, Index(1, latest, length, capacity))
	assertPos(t, 2, 0, Index(2, latest, length, capacity))
	assertPos(t, 3, 9, Index(3, latest, length, capacity))
	assertPos(t, 4, 8, Index(4, latest, length, capacity))

	//let's play with extrems
	latest = 4
	// buffer   indexes   0 1 2 3 4 5 6 7 8 9
	// circular indexes   4 3 2 1 0
	assertPos(t, 0, 4, Index(0, latest, length, capacity))
	assertPos(t, 1, 3, Index(1, latest, length, capacity))
	assertPos(t, 2, 2, Index(2, latest, length, capacity))
	assertPos(t, 3, 1, Index(3, latest, length, capacity))
	assertPos(t, 4, 0, Index(4, latest, length, capacity))

	latest = 0
	// buffer   indexes   0 1 2 3 4 5 6 7 8 9
	// circular indexes   0 x x x x x 4 3 2 1
	assertPos(t, 0, 0, Index(0, latest, length, capacity))
	assertPos(t, 1, 9, Index(1, latest, length, capacity))
	assertPos(t, 2, 8, Index(2, latest, length, capacity))
	assertPos(t, 3, 7, Index(3, latest, length, capacity))
	assertPos(t, 4, 6, Index(4, latest, length, capacity))

}

func assertPos(t *testing.T, i, j, k int) {
	if j != k {
		t.Fatalf("circular index %v should lead to absolute index %v, instead of %v", i, j, k)
	}
}

func TestAdd(t *testing.T) {
	M := 10
	b := New(M)

	for i := 0; i < M; i++ {
		err := b.Add(i)
		if err != nil {
			t.Fatal(err.Error())
		}
		p, err := b.Get(0)
		if err != nil {
			t.Fatal(err.Error())
		}
		if p != i {
			t.Fatalf("Add %v & Peek (%v). Oups", i, p)
		}
	}
	// the capacity is exhausted
	if b.Size() != b.Capacity() {
		t.Fatalf("%v Adds should have exhausted the capacity (%v). Len=%v", M, b.Capacity(), b.Size())
	}
}

func TestAddAll(t *testing.T) {
	M := 10
	b := New(M)

	err := b.AddAll(0, 1, 2, 3)
	if err != nil {
		t.Fatal(err.Error())
	}
	if b.Size() != 4 {
		t.Fatalf("Invalid length %v, expecting %v", b.Size(), 4)
	}

	p, err := b.Get(0)
	if err != nil {
		t.Fatal(err.Error())
	}
	if p != 3 {
		t.Fatalf("Latest value should be 3 instead of %v", p)
	}

	//let's add four more values, but this time we are not at the begining of the capacity
	// meaning that we are going to add in two times
	b.head = 8
	err = b.AddAll(0, 1, 2, 3)
	if err != nil {
		t.Fatal(err.Error())
	}
	if b.Size() != 8 {
		t.Fatalf("Invalid length %v, expecting %v", b.Size(), 8)
	}

	p, err = b.Get(0)
	if err != nil {
		t.Fatal(err.Error())
	}
	if p != 3 {
		t.Fatalf("Latest value should be 3 instead of %v", p)
	}

	// and now fill it up exactly
	b = New(4)

	err = b.AddAll(0, 1, 2, 3)
	if err != nil {
		t.Fatal(err.Error())
	}
	if b.Size() != 4 {
		t.Fatalf("Invalid length %v, expecting %v", b.Size(), 4)
	}

	// and now fill it too much
	//
	b = New(3)

	err = b.AddAll(0, 1, 2, 3)
	if err != ErrFull {
		t.Fatalf("should have failed with FullError, got %v", err)
	}
	if b.Size() != 0 {
		t.Fatalf("Invalid length %v, expecting %v", b.Size(), 0)
	}

}

func TestIncrease(t *testing.T) {

	x := New(5)
	x.AddAll(1, 2, 3, 4)

	// basic increase
	b := New(5)
	b.AddAll(1, 2, 3, 4)
	t.Logf("before %s", print(b))
	b.SetCapacity(10)
	t.Logf("after  %s", print(b))
	if !equals(b, x) {
		t.Errorf("increase failed. Different before: %s\nafter    %s", print(b), print(x))
	}

	b = New(6)
	b.head = 0 //fake an offset
	b.AddAll(1, 2, 3, 4)
	t.Logf("before %s", print(b))
	b.SetCapacity(10)
	t.Logf("after  %s", print(b))
	if !equals(b, x) {
		t.Errorf("increase failed. Different before: %s\nafter    %s", print(b), print(x))
	}

	b = New(6)
	b.head = 1 // values are all stick at the end
	b.AddAll(1, 2, 3, 4)
	t.Logf("before %s", print(b))
	b.SetCapacity(10)
	t.Logf("after  %s", print(b))
	if !equals(b, x) {
		t.Errorf("increase failed. Different before: %s\nafter    %s", print(b), print(x))
	}

	b = New(6)
	b.head = 3 //values overlap the end
	b.AddAll(1, 2, 3, 4)
	t.Logf("before %s", print(b))
	b.SetCapacity(10)
	t.Logf("after  %s", print(b))
}

func equals(b, c *Ring) bool {
	if b.Size() != c.Size() {
		return false
	}
	for i := 0; i < b.Size(); i++ {
		x, xerr := b.Get(i)
		y, yerr := c.Get(i)
		if xerr != nil {
			panic(xerr)
		}
		if yerr != nil {
			panic(yerr)
		}
		if x != y {
			return false
		}
	}
	return true
}

func print(b *Ring) string {
	latest := b.head
	end := Index(-1, latest, b.size, b.Capacity())
	if end < latest { // one piece
		switch {
		case end == 0:
			return fmt.Sprintf("*%v*   %v", b.buf[end:latest+1], b.buf[latest+1:])
		case latest+1 >= b.Capacity():
			return fmt.Sprintf("%v   *%v*", b.buf[:end], b.buf[end:latest+1])
		default:
			return fmt.Sprintf("%v  *%v*   %v", b.buf[:end], b.buf[end:latest+1], b.buf[latest+1:])

		}
	} else { //two pieces
		return fmt.Sprintf("*%v*  %v   *%v*", b.buf[:latest+1], b.buf[latest+1:end], b.buf[end:])
	}
	return ""
}
