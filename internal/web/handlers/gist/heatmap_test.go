package gist

import "testing"

func TestSplitHeatmapBands(t *testing.T) {
	weeks := make([]HeatmapWeek, 53)
	months := []HeatmapMonth{
		{Week: 0, Column: 1, Span: 5, Month: 11},
		{Week: 24, Column: 25, Span: 5, Month: 12},
		{Week: 26, Column: 27, Span: 5, Month: 1},
		{Week: 27, Column: 28, Span: 5, Month: 2},
		{Week: 52, Column: 53, Span: 5, Month: 3},
	}

	bands := splitHeatmapBands(weeks, months)
	if len(bands) != 2 {
		t.Fatalf("expected 2 bands, got %d", len(bands))
	}
	if len(bands[0].Weeks) != 27 || len(bands[1].Weeks) != 26 {
		t.Fatalf("unexpected week split: %d + %d", len(bands[0].Weeks), len(bands[1].Weeks))
	}

	first := bands[0].Months
	if len(first) != 3 {
		t.Fatalf("expected 3 month labels in first band, got %d", len(first))
	}
	if first[0].Column != 1 || first[1].Column != 25 || first[1].Span != 3 || first[2].Column != 27 || first[2].Span != 1 {
		t.Fatalf("unexpected first-band months: %+v", first)
	}

	second := bands[1].Months
	if len(second) != 2 {
		t.Fatalf("expected 2 month labels in second band, got %d", len(second))
	}
	if second[0].Column != 1 || second[0].Span != 5 || second[1].Column != 26 || second[1].Span != 1 {
		t.Fatalf("unexpected second-band months: %+v", second)
	}
}

func TestSplitHeatmapBandsEmpty(t *testing.T) {
	if bands := splitHeatmapBands(nil, nil); bands != nil {
		t.Fatalf("expected nil bands, got %+v", bands)
	}
}
