package catalog

import "testing"

func TestGetPrefix(t *testing.T) {
	res := getPrefix(0, 10)
	correct := "($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"
	if res != correct {
		t.Errorf("expected: %s\ngot %s", correct, res)
	}

	res = getPrefix(2, 10)
	correct = "($21, $22, $23, $24, $25, $26, $27, $28, $29, $30)"
	if res != correct {
		t.Errorf("expected: %s\ngot %s", correct, res)
	}
}
