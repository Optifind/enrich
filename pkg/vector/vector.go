package vector

import (
	"fmt"
	"strings"
)

// GetAverageVec returns the average vector of the given vectors.
func GetAverageVec(vectors [][]float32) []float32 {
	sumVec := GetSumVec(vectors)
	for i := range sumVec {
		sumVec[i] = sumVec[i] / float32(len(vectors))
	}
	return sumVec
}

func GetSumVec(vectors [][]float32) []float32 {
	sumVec := make([]float32, len(vectors[0]))
	for i := 0; i < len(vectors); i++ {
		for j, v := range vectors[i] {
			sumVec[j] += v
		}
	}
	return sumVec
}

func NewPGVector(vector []float32) string {
	strVec := vecToString(vector)
	return fmt.Sprintf("[%s]", strVec)
}

// Concatenates vector to a string. Vector components are separated by comma.
func vecToString(vec []float32) string {
	texts := make([]string, len(vec))
	for i := range vec {
		texts[i] = fmt.Sprint(vec[i])
	}
	return strings.Join(texts, ",")
}
