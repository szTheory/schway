package core

import "testing"

// TestDecodeInterfaceRequiredModeMutationKilled is 07-01 Task 2's D-07-42
// mutation-kill falsifier (QLT-08, D-07-41): with decoderRequireNonEmpty
// replaced through its unexported seam to accept "", a /1 document missing
// Mode stops being refused entirely -- proving the strict-decoding control
// actually bites, not merely appears to. Restoring the seam (via defer)
// restores the refusal. This is a same-package (white-box) test per
// D-07-42: the seam is unexported and must never be an exported
// package-level var on a production path, so only a same-package test can
// exercise it directly.
func TestDecodeInterfaceRequiredModeMutationKilled(t *testing.T) {
	const missingModeDoc = `{
		"schema": "schway.interface/1",
		"module_id": "m",
		"core_digest": "sha256:` + sixtyFourHex + `",
		"functions": [{
			"id": "f",
			"name": "f",
			"parameters": [{"id":"p","name":"p","type":"Byte","mode":"","drops":false}],
			"return": {"type":"Byte","mode":"owned","paths":[],"fresh":false},
			"abilities": [],
			"callable": false,
			"foreign": {"allocator":"","unwind":"","nonlocal_exit":""},
			"closure_digest": "sha256:` + sixtyFourHex + `"
		}]
	}`

	if _, err := DecodeInterface([]byte(missingModeDoc)); err == nil {
		t.Fatal("expected a missing-Mode document to be refused before mutation")
	}

	original := decoderRequireNonEmpty
	decoderRequireNonEmpty = func(string) bool { return true }
	defer func() { decoderRequireNonEmpty = original }()

	decoded, err := DecodeInterface([]byte(missingModeDoc))
	if err != nil {
		t.Fatalf("mutation (widening decoderRequireNonEmpty) had no observable effect: still refused with %v", err)
	}
	if decoded.V1 == nil || !decoded.Admissible {
		t.Fatalf("expected the mutated decoder to accept the missing-Mode document as admissible: %+v", decoded)
	}
}

const sixtyFourHex = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
