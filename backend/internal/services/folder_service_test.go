package services

import (
	"testing"
)

func TestParsePathIDs_EmptyRoot(t *testing.T) {
	ids := parsePathIDs(",")
	if len(ids) != 0 {
		t.Fatalf("root path should yield no ancestors, got %v", ids)
	}
}

func TestParsePathIDs_Hierarchy(t *testing.T) {
	// Two valid ObjectID hex strings.
	a := "507f1f77bcf86cd799439011"
	b := "507f191e810c19729de860ea"
	ids := parsePathIDs("," + a + "," + b + ",")
	if len(ids) != 2 {
		t.Fatalf("expected 2 ancestors, got %d", len(ids))
	}
	if ids[0].Hex() != a || ids[1].Hex() != b {
		t.Fatalf("ordering wrong: %v", ids)
	}
}

func TestParsePathIDs_SkipsInvalid(t *testing.T) {
	ids := parsePathIDs(",notanid,507f1f77bcf86cd799439011,")
	if len(ids) != 1 {
		t.Fatalf("expected 1 valid id, got %d", len(ids))
	}
}
