package signing

import "testing"

// supermarketSupportedSignVersions is mixlib-authentication's server-side
// SignedHeaderAuth::SUPPORTED_VERSIONS — the only mixlib signing-protocol
// versions the public Chef Supermarket's verifier accepts. Mixlib's
// ALGORITHM_FOR_VERSION table also knows about "1.3" (SHA-256), but the
// Supermarket application gates incoming requests on this smaller set, so a
// request signed with 1.3 is rejected before it ever reaches the crypto.
//
// Bumping SignVersion to "1.3" must therefore wait for a modernized
// Supermarket that advertises 1.3 in its supported set; until then a 1.3
// request earns a flat 401. The verify.rb gem test (signer_verifier_test.go)
// proves that rejection end-to-end against the real mixlib verifier.
var supermarketSupportedSignVersions = map[string]bool{
	"1.0": true,
	"1.1": true,
}

// TestSignVersionIsAcceptedBySupermarket is the always-on, Ruby-free guard
// that would have failed the v1.3 (SHA-256) regression that 401'd every
// `cinc supermarket share`: it asserts the version we actually sign with is
// one the public Supermarket verifier supports. The byte-for-byte signer
// vector tests can't catch that class of bug — a 1.3 signature is perfectly
// valid mixlib output; the problem is purely that the server won't accept
// 1.3 — so this assertion guards the one fact those tests don't.
func TestSignVersionIsAcceptedBySupermarket(t *testing.T) {
	if !supermarketSupportedSignVersions[SignVersion] {
		t.Fatalf("SignVersion = %q, which the public Supermarket verifier does NOT accept; "+
			"it only supports %v (mixlib SignedHeaderAuth::SUPPORTED_VERSIONS). "+
			"Signing with an unsupported version 401s every write request.",
			SignVersion, keysOf(supermarketSupportedSignVersions))
	}
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
