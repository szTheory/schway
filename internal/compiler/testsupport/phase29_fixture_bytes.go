package testsupport

import "bytes"

// Phase29FixtureBytes is the Go twin of examples/phase29/adapter.c. It is
// written independently of the C case table (D-07) so a common-mode error in
// the inputs cannot hide; TestPhase29FixtureBytesGeneratorsAgree compares them.
func Phase29FixtureBytes(caseID uint64) ([]byte, bool) {
	switch caseID {
	case 0:
		return []byte{}, true
	case 1:
		return []byte{0x00}, true
	case 2:
		return []byte("a"), true
	case 3:
		return []byte("Wikipedia"), true
	case 4:
		return make([]byte, 4096), true
	case 5:
		return bytes.Repeat([]byte{0xFF}, 5552), true
	case 6:
		return []byte{0x80, 0xFF}, true
	case 7:
		return bytes.Repeat([]byte{0xFF}, 4096), true
	case 8:
		return bytes.Repeat([]byte{0xFF}, 16384), true
	case 9:
		return bytes.Repeat([]byte{0xFF}, 65536), true
	}
	return nil, false
}

// Phase29FixtureCaseIDs lists every case id the fixture owner accepts.
func Phase29FixtureCaseIDs() []uint64 {
	return []uint64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
}

// phase29FixtureExpected holds pinned expected checksums per case id. They were
// produced by external oracles, never by any Schway implementation:
// Python 3.14.4 `zlib.adler32(data)` and Go 1.24.0 `hash/adler32.Checksum(data)`
// agree on every value below.
var phase29FixtureExpected = map[uint64]uint64{
	0: 1,
	1: 65537,
	2: 6422626,
	3: 300286872,
	4: 268435457,
	5: 4052720524,
	6: 33620352,
	7: 2170679522,
	8: 2967061426,
	9: 2006388466,
}

// Phase29FixtureExpected returns the pinned checksum for a case id.
func Phase29FixtureExpected(caseID uint64) (uint64, bool) {
	value, ok := phase29FixtureExpected[caseID]
	return value, ok
}
