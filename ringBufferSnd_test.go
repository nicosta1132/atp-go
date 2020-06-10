package atp

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	r := NewRingBufferSnd(10, 3)
	seg := makeSegment(0)
	_, err := r.insertSequence(seg)
	assert.NoError(t, err)
}

func TestInsertNotOrdered(t *testing.T) {
	r := NewRingBufferSnd(10, 3)
	seg := makeSegment(0)
	_, err := r.insertSequence(seg)
	assert.NoError(t, err)
	seg = makeSegment(2)
	_, err = r.insertSequence(seg)
	assert.Error(t, err)
}

func TestNotOrdered(t *testing.T) {
	r := NewRingBufferSnd(10, 3)
	seg := makeSegment(1)
	_, err := r.insertSequence(seg)
	assert.Error(t, err)
}

func TestFullSnd(t *testing.T) {
	r := NewRingBufferSnd(10, 3)
	for i := 0; i < 10; i++ {
		seg := makeSegment(uint32(i))
		_, err := r.insertSequence(seg)
		assert.NoError(t, err)
	}
	seg := makeSegment(11)
	_, err := r.insertSequence(seg)
	assert.Error(t, err)
}

func TestRemove(t *testing.T) {
	r := NewRingBufferSnd(10, 3)
	for i := 0; i < 10; i++ {
		seg := makeSegment(uint32(i))
		_, err := r.insertSequence(seg)
		assert.NoError(t, err)
	}
	r.remove(5)
	s := r.getTimedout(timeZero.Add(time.Second*time.Duration(r.timoutSec()) + 1))
	assert.Equal(t, 9, len(s))
}

func TestRemove5(t *testing.T) {
	r := NewRingBufferSnd(10, 3)
	for i := 0; i < 5; i++ {
		seg := makeSegment(uint32(i))
		_, err := r.insertSequence(seg)
		assert.NoError(t, err)
	}
	r.remove(4)
	s := r.getTimedout(timeZero.Add(time.Second*time.Duration(r.timoutSec()) + 1))
	assert.Equal(t, 4, len(s))
}

func TestNoRemove(t *testing.T) {
	r := NewRingBufferSnd(10, 3)
	for i := 0; i < 5; i++ {
		seg := makeSegment(uint32(i))
		_, err := r.insertSequence(seg)
		assert.NoError(t, err)
	}
	r.remove(4)
	//no timeout yet
	s := r.getTimedout(timeZero.Add(time.Second * time.Duration(r.timoutSec())))
	assert.Equal(t, 0, len(s))
}

func TestInsertRemove(t *testing.T) {
	r := NewRingBufferSnd(10, 3)

	for i := 0; i < 5; i++ {
		seg := makeSegment(uint32(i))
		_, err := r.insertSequence(seg)
		assert.NoError(t, err)
	}
	_, _, err := r.remove(3)
	assert.NoError(t, err)
	_, _, err = r.remove(1)
	assert.NoError(t, err)

	s := r.getTimedout(timeZero)
	assert.Equal(t, 0, len(s))
	s = r.getTimedout(timeZero.Add(time.Second*time.Duration(r.timoutSec()) + 1))
	assert.Equal(t, 3, len(s))
}

func TestInsertRemove2(t *testing.T) {
	r := NewRingBufferSnd(10, 3)
	seg := makeSegment(0)
	_, err := r.insertSequence(seg)
	assert.NoError(t, err)
	_, _, err = r.remove(0)
	assert.NoError(t, err)
	_, err = r.insertSequence(seg)
	assert.Error(t, err)
	s := r.getTimedout(timeZero)
	assert.Equal(t, 0, len(s))
	s = r.getTimedout(timeZero.Add(time.Second*time.Duration(r.timoutSec()) + 1))
	assert.Equal(t, 0, len(s))
}

func TestAlmostFull(t *testing.T) {
	r := NewRingBufferSnd(10, 3)
	for i := 0; i < 10; i++ {
		seg := makeSegment(uint32(i))
		_, err := r.insertSequence(seg)
		assert.NoError(t, err)
	}
	r.remove(4)
	seg := makeSegment(10)
	_, err := r.insertSequence(seg)
	assert.Error(t, err)
	r.remove(0)

	seg = makeSegment(10)
	_, err = r.insertSequence(seg)
	assert.NoError(t, err)
}

func TestFuzz(t *testing.T) {
	r := NewRingBufferSnd(10, 3)

	seqIns := 0
	seqRem := 0
	rand.Seed(42)

	for j := 0; j < 1000000; j++ {
		rnd := rand.Intn(10) + 1

		for i := 0; i < rnd; i++ {
			if seqIns == 32 {
				print("aue")
			}
			seg := makeSegment(uint32(seqIns))

			ins, err := r.insertSequence(seg)
			if err != nil {
				assert.NoError(t, err)
			}
			if !ins {
				rnd = i + 1
				break
			}
			seqIns++
		}

		rnd2 := rand.Intn(rnd) + 1
		if rand.Intn(2) == 0 {
			rnd2 = rand.Intn(seqIns-seqRem) + 1
		}

		if j == 5 {
			print("aeu")
		}

		for i := 0; i < rnd2; i++ {
			_, _, err := r.remove(uint32(seqRem))
			if err != nil {
				assert.NoError(t, err)
				_, _, err = r.remove(uint32(seqRem))
			}
			seqRem++
		}

		if rand.Intn(3) == 0 {
			r = r.resize(r.size() + 1)
		}

		s := r.getTimedout(timeZero.Add(time.Hour))
		fmt.Printf("size: %v\n", len(s))

	}
	fmt.Printf("send %v, recv %v", seqIns, seqRem)
}
