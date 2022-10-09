package goskiplist_test

import (
	"testing"

	"github.com/aobt/goskiplist"
	"github.com/stretchr/testify/assert"
)

func TestSkipList(t *testing.T) {

	{
		level := 16
		sl := goskiplist.NewSkipList[int, string](4, level)
		assert.Equal(t, sl.StepSize(), 4)
		assert.Equal(t, sl.Level(), level)
		assert.Equal(t, sl.Length(), int32(0))
	}

	{
		level := 32
		sl := goskiplist.NewSkipList[int, string](4, level)

		var (
			err error
			v   *string
		)

		//----Put One----
		s0 := "hello"
		err = sl.Put(0, &s0)
		assert.NoError(t, err)
		assert.Equal(t, sl.Length(), int32(1))

		s1 := "world"
		err = sl.Put(1, &s1)
		assert.NoError(t, err)
		assert.Equal(t, sl.Length(), int32(2))

		s2 := "goskiplist"
		err = sl.Put(2, &s2)
		assert.NoError(t, err)
		assert.Equal(t, sl.Length(), int32(3))

		//----Find One----
		v, err = sl.Find(0)
		assert.NoError(t, err)
		assert.Equal(t, *v, "hello")

		v, err = sl.Find(1)
		assert.NoError(t, err)
		assert.Equal(t, *v, "world")

		v, err = sl.Find(2)
		assert.NoError(t, err)
		assert.Equal(t, *v, "goskiplist")

		_, err = sl.Find(100)
		assert.Error(t, err)

		//----Find Min----
		v, err = sl.FindMin()
		assert.NoError(t, err)
		assert.Equal(t, *v, "hello")

		//----Find Max----
		v, err = sl.FindMax()
		assert.NoError(t, err)
		assert.Equal(t, *v, "goskiplist")

		//----Pop One----
		v, err = sl.Pop(1)
		assert.NoError(t, err)
		assert.Equal(t, *v, "world")
		assert.Equal(t, sl.Length(), int32(2))

		//----Pop Min----
		v, err = sl.PopMin()
		assert.NoError(t, err)
		assert.Equal(t, *v, "hello")
		assert.Equal(t, sl.Length(), int32(1))

		//----Pop Max----
		v, err = sl.PopMax()
		assert.NoError(t, err)
		assert.Equal(t, *v, "goskiplist")
		assert.Equal(t, sl.Length(), int32(0))
	}

}
