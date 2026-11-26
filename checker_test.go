package main

import (
	"strings"
	"testing"
)

func kinds(is []Issue) []string {
	var k []string
	for _, i := range is {
		k = append(k, i.Kind)
	}
	return k
}

func TestDuplicateEmail(t *testing.T) {
	csv := "email,domain\na@x.test,x.test\nA@x.test,x.test\n"
	is, n, err := Check(strings.NewReader(csv), Options{})
	if err != nil || n != 2 {
		t.Fatal(err, n)
	}
	if len(is) != 1 || is[0].Kind != "duplicate" || is[0].Row != 3 {
		t.Fatalf("got %+v", is)
	}
}

func TestDomainOnlyDuplicatesNormalised(t *testing.T) {
	csv := "domain\nhttps://www.acme.test/about\nacme.test\n"
	is, _, _ := Check(strings.NewReader(csv), Options{})
	if len(is) != 1 || is[0].Kind != "duplicate" {
		t.Fatalf("got %+v", is)
	}
}

func TestSharedDomainAllowedWhenEmailsDiffer(t *testing.T) {
	csv := "email,domain\na@x.test,x.test\nb@x.test,x.test\n"
	is, _, _ := Check(strings.NewReader(csv), Options{})
	if len(is) != 0 {
		t.Fatalf("two people at one company are not duplicates: %+v", is)
	}
}

func TestRequiredAndStage(t *testing.T) {
	csv := "email,title,stage\na@x.test,,mql\nb@x.test,CEO,customer\nc@x.test,CTO,SQL\n"
	is, _, _ := Check(strings.NewReader(csv), Options{Required: []string{"title"}})
	got := strings.Join(kinds(is), ",")
	if got != "empty_field,unknown_stage" {
		t.Fatalf("got %s", got)
	}
}

func TestBadEmailAndNoKeyColumns(t *testing.T) {
	is, _, _ := Check(strings.NewReader("email\nnot-an-email\n"), Options{})
	if len(is) != 1 || is[0].Kind != "bad_email" {
		t.Fatalf("got %+v", is)
	}
	if _, _, err := Check(strings.NewReader("name\nx\n"), Options{}); err == nil {
		t.Fatal("expected error without a match key")
	}
}
