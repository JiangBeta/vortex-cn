package i18n

import "testing"

func TestTHitZh(t *testing.T) {
	SetLang("zh")
	if got := T("Servers"); got != "服务器" {
		t.Fatalf(`T("Servers") = %q, want 服务器`, got)
	}
}

func TestTMissFallsBackToEnglish(t *testing.T) {
	SetLang("zh")
	if got := T("NoSuchKey"); got != "NoSuchKey" {
		t.Fatalf(`T("NoSuchKey") = %q, want key itself`, got)
	}
}

func TestTEnglish(t *testing.T) {
	SetLang("en")
	if got := T("Servers"); got != "Servers" {
		t.Fatalf(`T("Servers") in en = %q, want Servers`, got)
	}
}

func TestSetLangIgnoresInvalid(t *testing.T) {
	SetLang("zh")
	SetLang("fr")
	if Lang() != "zh" {
		t.Fatalf(`Lang() = %q, want zh (invalid value must be ignored)`, Lang())
	}
}

func TestTf(t *testing.T) {
	SetLang("zh")
	if got := Tf("count=%d", 3); got != "count=3" {
		t.Fatalf(`Tf("count=%%d", 3) = %q, want count=3`, got)
	}
}
