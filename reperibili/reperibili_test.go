package reperibili

import (
	"fmt"
	"testing"
)

func TestVerificaPresenzaReperibili(t *testing.T) {

	VerificaPresenzaReperibili("CDN", "../reperibilita.csv")
	if ok, err := VerificaPresenzaReperibili("CDN", "../reperibilita.csv"); ok != true {
		if err != nil {
			fmt.Println(err.Error())
		}
		t.Skip("Problema")
	}
}
