package conf

import (
	"testing"
)

func TestLoad(t *testing.T) {
	m, err := Load()
	if err != nil {
		t.Fail()
	}
	t.Log(m)
}
