package vector

import (
	"math"
	"reflect"
	"testing"
)

func TestGetSumVec(t *testing.T) {
	tests := []struct {
		vectors  [][]float32
		expected []float32
	}{
		{
			vectors: [][]float32{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			expected: []float32{12, 15, 18},
		},
		{
			vectors: [][]float32{
				{1, 1, 1},
				{1, 1, 1},
				{1, 1, 1},
			},
			expected: []float32{3, 3, 3},
		},
		{
			vectors: [][]float32{
				{0, 0, 0},
				{0, 0, 0},
			},
			expected: []float32{0, 0, 0},
		},
		{
			vectors: [][]float32{
				{-1, -2, -3},
				{-4, -5, -6},
			},
			expected: []float32{-5, -7, -9},
		},
		{
			vectors: [][]float32{
				{1.5, 2.5, 3.5},
				{4.5, 5.5, 6.5},
			},
			expected: []float32{6, 8, 10},
		},
	}

	for _, test := range tests {
		result := GetSumVec(test.vectors)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("GetSumVec(%v) = %v; want %v", test.vectors, result, test.expected)
		}
	}
}

func TestGetAverageVec(t *testing.T) {
	tests := []struct {
		vectors  [][]float32
		expected []float32
	}{
		{
			vectors: [][]float32{
				{2, 2, 2},
				{2, 2, 2},
				{2, 2, 2},
			},
			expected: []float32{2, 2, 2},
		},
		{
			vectors: [][]float32{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			expected: []float32{4, 5, 6},
		},
		{
			vectors: [][]float32{
				{0, 0, 0},
				{0, 0, 0},
			},
			expected: []float32{0, 0, 0},
		},
		{
			vectors: [][]float32{
				{-1, -2, -3},
				{-4, -5, -6},
			},
			expected: []float32{-2.5, -3.5, -4.5},
		},
		{
			vectors: [][]float32{
				{1.5, 2.5, 3.5},
				{4.5, 5.5, 6.5},
			},
			expected: []float32{3, 4, 5},
		},
	}

	for _, test := range tests {
		result := GetAverageVec(test.vectors)
		for i := range result {
			if math.Abs(float64(result[i]-test.expected[i])) > 1e-6 {
				t.Errorf("GetAverageVec(%v) = %v; want %v", test.vectors, result, test.expected)
			}
		}
	}
}
