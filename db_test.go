package ip2loc

import (
	"errors"
	"testing"
)

const (
	testDB = "./testdata/ip2location-lite-db1.ipv6.bin"

	v4Valid   = "8.8.8.8"
	v4Invalid = "404.1.2.3"

	v6Embeddedv4 = "0:0:0:0:0:ffff:808:808"
	v6Toredo     = "2001:0000:4136:e378:8000:63bf:3fff:fdd2"
)

func TestBadPath(t *testing.T) {
	_, err := New(testDB + "badpath")
	if err == nil {
		t.Fatal("expected error; got none")
	}
}

func TestVersion(t *testing.T) {
	db, err := New(testDB)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Version()
	if err != nil {
		t.Fatal(err)
	}
}

func TestValidIPv4(t *testing.T) {
	db, err := New(testDB)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	result, err := db.Query(v4Valid)
	if err != nil {
		t.Fatal(err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if result.CountryCode != "US" {
		t.Fatalf("CountryCode: expected 'US', got '%s'\n", result.CountryCode)
	}

	if result.CountryName != "United States" {
		t.Fatalf("CountryName: expected 'United States', got '%s'\n", result.CountryName)
	}
}

func TestV4EmbeddedIPv4(t *testing.T) {
	db, err := New(testDB)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	result, err := db.Query(v6Embeddedv4)
	if err != nil {
		t.Fatal(err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestToredoV6(t *testing.T) {
	db, err := New(testDB)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	result, err := db.Query(v6Toredo)
	if err != nil {
		t.Fatal(err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestInvalidV4(t *testing.T) {
	db, err := New(testDB)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	_, err = db.Query(v4Invalid)
	if !errors.Is(err, ErrInvalidIP{}) {
		t.Error("expected error, got none")
	}
}

func TestValidV6(t *testing.T) {
	db, err := New(testDB)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	result, err := db.Query("2001:4860:4860::8888")
	if err != nil {
		t.Fatal(err)
	}

	if result == nil {
		t.Fatal("result is nil")
	}

	if len(result.CountryCode) != 2 {
		t.Fatalf("CountryCode: expected 2-character country code, got '%s'\n", result.CountryCode)
	}

	if len(result.CountryName) == 0 {
		t.Fatalf("CountryName: expected non-empty country name, got '%s'\n", result.CountryName)
	}
}

func BenchmarkQuery(b *testing.B) {
	db, err := New(testDB)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := db.Query(v4Valid)
		if err != nil {
			b.Fatal(err)
		}
	}
}
