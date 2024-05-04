package slice_util

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCountElements(t *testing.T) {
	type Data []struct {
		xs       []string
		expected map[string]int
	}

	data := Data{
		{
			xs:       nil,
			expected: map[string]int{},
		},
		{
			xs:       []string{},
			expected: map[string]int{},
		},
		{
			xs: []string{"foo"},
			expected: map[string]int{
				"foo": 1,
			},
		},
		{
			xs: []string{"foo", "bar"},
			expected: map[string]int{
				"bar": 1,
				"foo": 1,
			},
		},
		{
			xs: []string{"foo", "bar", "baz"},
			expected: map[string]int{
				"bar": 1,
				"baz": 1,
				"foo": 1,
			},
		},
		{
			xs: []string{"foo", "bar", "baz", "foo"},
			expected: map[string]int{
				"bar": 1,
				"baz": 1,
				"foo": 2,
			},
		},
	}

	for _, d := range data {
		actual := CountElements(d.xs)
		diff := cmp.Diff(d.expected, actual)
		if diff != "" {
			t.Error(diff)
		}
	}
}

func TestSliceToSet(t *testing.T) {
	type Data []struct {
		xs       []string
		expected map[string]int
	}

	data := Data{
		{
			xs:       nil,
			expected: map[string]int{},
		},
		{
			xs:       []string{},
			expected: map[string]int{},
		},
		{
			xs: []string{"foo"},
			expected: map[string]int{
				"foo": 1,
			},
		},
		{
			xs: []string{"foo", "bar"},
			expected: map[string]int{
				"bar": 1,
				"foo": 1,
			},
		},
		{
			xs: []string{"foo", "bar", "baz"},
			expected: map[string]int{
				"bar": 1,
				"baz": 1,
				"foo": 1,
			},
		},
		{
			xs: []string{"foo", "bar", "baz", "foo"},
			expected: map[string]int{
				"bar": 1,
				"baz": 1,
				"foo": 2,
			},
		},
	}

	for _, d := range data {
		actual := SliceToSet(d.xs)
		diff := cmp.Diff(d.expected, actual)
		if diff != "" {
			t.Error(diff)
		}
	}
}

func TestSubtractSet(t *testing.T) {
	type Data struct {
		xs       map[string]int
		ys       map[string]int
		expected map[string]int
	}

	data := []Data{
		{
			xs:       nil,
			ys:       nil,
			expected: map[string]int{},
		},
		{
			xs: map[string]int{
				"foo": 1,
			},
			ys: nil,
			expected: map[string]int{
				"foo": 1,
			},
		},
		{
			xs: map[string]int{
				"foo": 1,
			},
			ys: map[string]int{},
			expected: map[string]int{
				"foo": 1,
			},
		},
		{
			xs: map[string]int{
				"foo": 1,
			},
			ys: map[string]int{
				"foo": 1,
			},
			expected: map[string]int{},
		},
		{
			xs: map[string]int{
				"bar": 1,
				"foo": 1,
			},
			ys: nil,
			expected: map[string]int{
				"bar": 1,
				"foo": 1,
			},
		},
		{
			xs: nil,
			ys: map[string]int{
				"bar": 1,
				"foo": 1,
			},
			expected: map[string]int{},
		},
		{
			xs: map[string]int{
				"bar": 1,
				"foo": 1,
			},
			ys: map[string]int{},
			expected: map[string]int{
				"bar": 1,
				"foo": 1,
			},
		},
		{
			xs: map[string]int{
				"bar": 1,
				"foo": 1,
			},
			ys: map[string]int{
				"foo": 1,
			},
			expected: map[string]int{
				"bar": 1,
			},
		},
		{
			xs: map[string]int{
				"bar": 1,
				"foo": 1,
			},
			ys: map[string]int{
				"bar": 1,
				"foo": 1,
			},
			expected: map[string]int{},
		},
	}

	for _, d := range data {
		actual := SubtractSet(d.xs, d.ys)
		diff := cmp.Diff(d.expected, actual)
		if diff != "" {
			t.Error(diff)
		}
	}
}

func TestSubtractSlice(t *testing.T) {
	type Data struct {
		xs       []string
		ys       []string
		expected []string
	}

	data := []Data{
		{
			xs:       nil,
			ys:       nil,
			expected: []string{},
		},
		{
			xs:       []string{"foo"},
			ys:       nil,
			expected: []string{"foo"},
		},
		{
			xs:       []string{"foo"},
			ys:       []string{},
			expected: []string{"foo"},
		},
		{
			xs:       []string{"foo"},
			ys:       []string{"foo"},
			expected: []string{},
		},
		{
			xs:       []string{"foo", "bar"},
			ys:       nil,
			expected: []string{"foo", "bar"},
		},
		{
			xs:       nil,
			ys:       []string{"foo", "bar"},
			expected: []string{},
		},
		{
			xs:       []string{"foo", "bar"},
			ys:       []string{},
			expected: []string{"foo", "bar"},
		},
		{
			xs:       []string{"foo", "bar"},
			ys:       []string{"foo"},
			expected: []string{"bar"},
		},
		{
			xs: []string{"foo", "bar"},
			ys: []string{"foo", "bar"},
			expected: []string{},
		},
	}

	for _, d := range data {
		actual := SubtractSlice(d.xs, d.ys)
		diff := cmp.Diff(d.expected, actual)
		if diff != "" {
			t.Error(diff)
		}
	}
}
