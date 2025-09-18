package golicenses

import (
	"errors"
	"testing"
)

func TestLicenseFinder_EmptyResults(t *testing.T) {
	finder := NewLicenseFinder([]string{"./testdata/empty"}, []string{"origin"}, 0.9)
	ch, err := finder.Find()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	results := []LicenseResult{}
	for res := range ch {
		results = append(results, res)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
	if results[0].License != "" {
		t.Errorf("expected empty license, got %q", results[0].License)
	}
}

func TestLicenseFinder_MissingLicense(t *testing.T) {
	finder := NewLicenseFinder([]string{"./testdata/missing"}, []string{"origin"}, 0.9)
	ch, err := finder.Find()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for res := range ch {
		if res.License != "" {
			t.Errorf("expected empty license, got %q", res.License)
		}
	}
}

func TestLicenseFinder_MultipleRemotes(t *testing.T) {
	finder := NewLicenseFinder([]string{"./testdata/multi"}, []string{"origin", "upstream"}, 0.9)
	_, err := finder.Find()
	if err != nil && !errors.Is(err, nil) {
		t.Errorf("unexpected error: %v", err)
	}
}
