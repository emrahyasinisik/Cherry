package factory

import "testing"

func TestDartPackage(t *testing.T) {
	if dartPackage("kahve-siparis") != "kahve_siparis" {
		t.Fatalf("%s", dartPackage("kahve-siparis"))
	}
	if dartPackage("2go") != "app_2go" {
		t.Fatalf("%s", dartPackage("2go"))
	}
}

func TestPascalIdent(t *testing.T) {
	if pascalIdent("kahve-siparis") != "KahveSiparis" {
		t.Fatalf("%s", pascalIdent("kahve-siparis"))
	}
}
