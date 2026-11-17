package helper

import "testing"

func TestTruncate(t *testing.T) {
	truncate := funcMap["truncate"].(func(string, int) string)

	if got := truncate("Bonjour", 10); got != "Bonjour" {
		t.Errorf("texte court modifie : %q", got)
	}
	if got := truncate("Bonjour tout le monde", 7); got != "Bonjour..." {
		t.Errorf("texte long mal tronque : %q", got)
	}
}
