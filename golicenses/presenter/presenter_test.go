package presenter

import (
	"github.com/khulnasoft/go-licenses/golicenses"
	"os"
	"testing"
)

func TestGetPresenter_UnknownOption(t *testing.T) {
	ch := make(chan golicenses.LicenseResult)
	close(ch)
	p := GetPresenter(UnknownPresenter, ch)
	if p != nil {
		t.Error("expected nil presenter for unknown option")
	}
}

func TestGetPresenter_TextOption_EmptyChannel(t *testing.T) {
	f, err := os.CreateTemp("", "presenter_test_*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(f.Name())
	ch := make(chan golicenses.LicenseResult)
	close(ch)
	p := GetPresenter(TextPresenter, ch)
	if p == nil {
		t.Fatal("expected non-nil presenter for TextPresenter option")
	}
	err = p.Present(f)
	if err != nil {
		t.Errorf("expected no error for empty channel, got %v", err)
	}
}
