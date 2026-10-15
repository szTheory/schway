# Phase 37 focused acceptance evidence

Evidence date: 2026-10-15  
Code source revision: `b900dfe31abb600a82b31ab6e8cc2a8b9b1e74ca`  
Host/toolchain: macOS Darwin/arm64, Go 1.24.0, installed Clang, Node v26.10.0.  
Status: fixed-source local capture complete; parent verifier, security, review, and requirement acceptance remain pending.

This is an executor-produced local receipt, not the independent phase-verifier
or security verdict. The owner map is
[37-VALIDATION.md](37-VALIDATION.md); the exact command strings remain in its
C01-C26 registry. C01-C26 below were run with
`GOCACHE=/tmp/ai-lang-verification-gocache` for Go commands. The original
pre-addition receipts are preserved in their plan summaries: the final C01,
C02, and C25 reruns are regression receipts, not replacement baselines.

## Source, test, and fixture binding

The tracked-only guide runner copied exactly 2,691 entries from
`git ls-files -z`. Its final manifest line was
`SOURCE_MANIFEST files=2691 sha256=891462a2a2b5ecd9d996494242a0e682136ef6755d17874311576db9534e0d80 copied=git-ls-files-only`.
The copy path validates relative paths and excludes untracked/ignored inputs.
The test’s tracked-copy negative control confirmed an untracked required helper
was absent and the copied guide command reached its missing-helper failure.

| Path | SHA-256 |
| --- | --- |
| `cmd/schway/main_phase37_guide_test.go` | `7b4ce9eae5f1dc3129811b31fbc9ec346e544c0193236cc85986030cde212037` |
| `examples/crc32/README.md` | `955b6a4a28c0b63feb93cd96a931f01b9f3dec5107cb114d738aab57758de834` |
| `examples/crc32/crc32.expected.json` | `b334fd2026c1d8633b6cc405e5d84813ee73010ea8921594e412864d244a794d` |
| `examples/crc32/crc32.schway` | `8a8c7570409a6bf96258a35b5964dc899a2347871b564fb59e081beb011388c9` |
| `examples/crc32/crc_low.schway` | `797f9cd2e733c3590a593518fe671a1fda553eb6ce22ea50d70f1a26d606dccc` |
| `examples/crc32/file_bytes.bindings.json` | `c2f960ff2a92a5d9a60356711812ea60595a2f81cf6c0ff2c4b259ce6380103c` |
| `37-PROHIBITION-EVIDENCE.md` | `bc6ee7f03ef1ec0613966802e83bb1454229da06be2de21350dfe6354d39e3eb` |

## Exact owner command capture

Each Go count is top-level root RUN/PASS; nested controls are retained in the
raw log. C05 is a direct Node TAP run. C06 is the canonical producer, whose
three supported prohibition descriptors each retain bad/clean/default
subjects. Raw logs remain at the listed local paths and are SHA-256 bound.

| ID | Reached owner roots / result | Raw local log | Log SHA-256 |
| --- | --- | --- | --- |
| C01 | 3/3 Go roots passed | `/tmp/schway37-c01-final.log` | `8ecf3bc3d74ffb85ef08118921c99e1ba4b9cb4c0b58a53173d294e0e88b55c6` |
| C02 | 1/1 Go root passed | `/tmp/schway37-c02-final.log` | `d49d517f4cdb86a277a9b67e20aa848c626a2f8c3287845b82bf95cba593e011` |
| C03 | 1/1 Go root passed | `/tmp/schway37-c03-final.log` | `56c2a05c0de3a5462725dbd378348d37063fa05e2d81e7938a2f35677aef41aa` |
| C04 | 2/2 Go roots passed | `/tmp/schway37-c04-final.log` | `679d8f96116249500386b3808ebe5cd7ef31ab3cb7f87a6adcccada61dc15dde` |
| C05 | Node TAP: 3 passed, 0 failed, 0 skipped | `/tmp/schway37-c05-final.log` | `90ae427336d40afd7fd02bf50bae2360e4c81769b12e5e44e86a2cf41966232d` |
| C06 | Canonical producer passed; three descriptors source-bound | `/tmp/schway37-c06-final-capture.log` | `b69d5d9d69e88953e5adb74a869e72253ea15a1163638198898c8621727a65a2` |
| C07 | Corrected capture: 2/2 Go roots passed | `/tmp/schway37-c07-corrected-final.log` | `ce114d58d00fa694e418b2b863394985e91339390523006471750dc76c6f44dd` |
| C08 | 4/4 Go roots passed | `/tmp/schway37-c08-final.log` | `a6f48108e0def05f84ab78ca131044d4e3e8b93adfdb66ef9773ef06d5ee1afa` |
| C09 | 4/4 Go roots passed | `/tmp/schway37-c09-final.log` | `ed116314ec18e4c7f7d3d05ee36ec4d5548b7b179aef754c0ab9aa47eafbb1a2` |
| C10 | 1/1 Go root passed | `/tmp/schway37-c10-final.log` | `e58d0204f3feef76804319bcf58f678138afb714ebc55ae5651fb34d01900afb` |
| C11 | 4/4 Go roots passed | `/tmp/schway37-c11-final.log` | `65e1e89a0ba2de6670e84f2bc2df4f02a8023e95e21e2261cd17616f9aae8f60` |
| C12 | 4/4 Go roots passed | `/tmp/schway37-c12-final.log` | `7cddcc1941bac2bf73561355ae1fd4a7d3aaa38786e3343b8803073e8c3d21fe` |
| C13 | 3/3 Go roots passed; sanitizer subcases ran, no skips | `/tmp/schway37-c13-final.log` | `2da07c40a550f3334bd7ba52628fa396184ff2b31f2d59f8d9a30d0d81c1d38b` |
| C14 | 2/2 Go roots passed | `/tmp/schway37-c14-final.log` | `fd89c929706bce63c5711a9df497cafc42f1ebab7a8ef1090a113ddc76873f0b` |
| C15 | 2/2 Go roots passed | `/tmp/schway37-c15-final.log` | `9529364c526b36469dd9fd6f7cd5e46530c6d70b7901343722cb6f5db57c501d` |
| C16 | 1/1 Go root passed | `/tmp/schway37-c16-final.log` | `2f2d5e58c69eef47b1299e15598e2cd1250183c05b0ce20805fc942e2aa416b8` |
| C17 | 1/1 Go root passed | `/tmp/schway37-c17-final.log` | `e37916e439276683f1e78a1b20d805f183f48040315dd979dda6aef11eef2953` |
| C18 | 1/1 Go root passed | `/tmp/schway37-c18-final.log` | `f37700b3c13ca607a5829bd01987128ceae82a7ecfe3aed2302e5134d3c8a01e` |
| C19 | 4/4 Go roots passed | `/tmp/schway37-c19-final.log` | `bf3f58ce45c11e40509be7e01dbdefd952f6d1f7f81cab88ba2d27660d8f289e` |
| C20 | 2/2 Go roots passed | `/tmp/schway37-c20-final.log` | `aaf138af0f454ba85a9f2dca896044368ae36c103bc2c9728942902ba570cc9a` |
| C21 | 2/2 Go roots passed | `/tmp/schway37-c21-final.log` | `d314751a0e43dc4a6d308d68f358a1a6d5d3c167447df15d62b1167b63223198` |
| C22 | 1/1 Go root passed; 78/78 guide tuples passed | `/tmp/schway37-c22-final-capture.log` | `e065574bd3f1d1eefe6f64fb356ca4f61342d1c68bc5c5f213f8999d7796baa3` |
| C23 | 5/5 Go roots passed, including all Phase32 guide roots | `/tmp/schway37-c23-final-capture.log` | `47325789da133879a2bc29c1acc149e7fd47b1e7ff7a685d4f3675fb3215a2ae` |
| C24 | 3/3 Go roots passed; 78/78 guide tuples passed | `/tmp/schway37-c24-final-capture.log` | `13d0f292d2d7d762d779d7d5af7fcbbc53153dfc0574c8ce58424a48db6460c7` |
| C25 | 4/4 Go roots passed as final regression | `/tmp/schway37-c25-final.log` | `926b2249324611868e4bf413ecdc5e2bec2ff5fea436a9c5224cb783250b4770` |
| C26 | 6/6 Go roots passed; both additive ledgers' armed controls reached | `/tmp/schway37-c26-final.log` | `9abd16a2d2de788441552eeba19c4906f22994d93542bdc521e4219ac27b56ff` |

The corrected map has 63 passing Go root invocations across its repeated
owner commands. There were no selected Go root skips, no sanitizer skips, no
compile/tool failures, and no failing roots in the accepted capture. The
initial C07 attempt is excluded and preserved at
`/tmp/schway37-c07-final.log` (SHA-256
`2a477348cddb4419fa8d9014bb6439e9aa0b8507fc9604eb3a6d0322c8d77a12`): its selector named nonexistent
`TestPhase37CaptureDisabledRetention`, yielding “no tests to run” in the
interpreter package while the session root passed. The map now selects the
actual `TestPhase37CaptureRetentionControls`; the corrected C07 log above
shows both package owners reached and passing.

## Guide behavior proved

C22 and C24 ran the marked commands from the actual tracked-only copy. Thirteen
fixture cases were checked through each of six combinations: `crc32` and
`crc_low`, each on interpreter, one-shot native, and retained native. The
78 exact tuples include empty bytes, zero bytes, `123456789`, high-bit and
all-byte vectors, cap-minus-one/cap/cap-plus-one, missing path, directory,
flag-shaped path, space path, and shell-metacharacter path. Each tuple checked
exit code, stdout, and stderr against `crc32.expected.json`; retained builds
also required executable artifacts. The guide creates binary fixtures with
Python only as setup, in ignored `.schway/crc32` paths. The Python/zlib values
are independent oracle provenance, not a compiler/runtime dependency.

The guide distinguishes CRC-32/ISO-HDLC from POSIX `cksum`/CRC-32/CKSUM and
CRC32C, prints unsigned decimal with a newline, uses fixed interpreter
FileError outcomes and the 65,536-byte reader limit, and describes the
native binding-specific limit. It preserves the empty-file result
`(0, "0\\n", "")` and the empty-path result `(65, "", "")`. It limits
claims about capture growth, formatter canonicality, arbitrary timeout parity,
and physical cleanup to their named evidence.

## Independent proof and evidence boundaries

The fresh canonical proof in C06 updated
[37-PROHIBITION-EVIDENCE.md](37-PROHIBITION-EVIDENCE.md) at the code revision
above. Its three descriptor-backed claims cover native capture exhaustion,
causal evidence not being sold as numeric CRC truth, and interpreter
FileBytes capture exhaustion. Each keeps bad/clean/default subjects; bad
subjects reach their intended assertion and clean/default controls pass. The
raw descriptorless/unclassified projection remains historical and unchanged.
Canonical proof does not establish independent phase-verifier, security, or
review acceptance.

The fixed-source capture preserves earlier Phase37 distinctions:

- The public expected-first matrix and five semantic mutants remain distinct
  evidence lanes; a causal witness does not substitute for independent
  numerical answers.
- No-edit module rows and InputID stay stable. Native executable bytes and
  full BuildID may vary under the existing incomplete/non-cacheable receipt
  contract. This is not a deterministic native-binary claim.
- The additive Phase37 surface witness preserves the historical Phase34
  pinned bytes and does not change the ten-field public wire schema.
- Reader progress behavior and exhausted-capture defect classification fixes
  remain in the earlier plan summaries and current owner tests.
- C13 ran the focused physical cleanup and ASan/UBSan controls on Darwin.
  This does not claim Darwin leak detection or Linux execution.
- The maximum-size interpreter reader completed in 2.46 seconds in one C07
  observation on Darwin/arm64. No latency distribution or arbitrary
  cancellation/timeout parity is claimed.

## Historical and pending evidence

The pre-addition and original tracer baselines remain at the source revisions
recorded by Plans 37-01 through 37-07. Existing Phase35/36 and M006 hosted
receipts remain historical. The recurring `.github/workflows/ci.yml` is
source-inspected integration evidence; it was not dispatched or run here.
There is no Phase37 hosted receipt, local Linux execution, whole-package or
project-wide suite, broad native matrix, human readability result, or
cold/warm full-loop distribution. The four Phase37 requirements remain
pending parent acceptance.
