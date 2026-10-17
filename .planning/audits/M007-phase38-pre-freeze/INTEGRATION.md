# M007 Phase 38 Pre-Freeze Integration Audit

## Verdict

**Integration paths are wired at source. The sole pending requirement is the expected EVD-17 hosted evidence run.** I found no broken cross-phase connection or additional implementation blocker. Phase 38 remains in progress; this is a pre-freeze source audit, not a milestone completion or freeze claim.

The audit is anchored to source HEAD `06e72dda0addd2744106e22ef35fec539337999b`, code digest `f21194d7244778ac073269734e0103a6e1e0f6c59335e84305176401b465aadf`, and current workspace config SHA-256 `148ccc4c56db8516788c264216ae277d9860d43e00b92de299d8ccff2fd59521`. These identify the inspected candidate; no freeze object, hosted run, or final ingest is asserted.

## Scope and evidence

I traced the six completed phases (33, 34, 35, 36, 37, and 37.1) into the Phase 38 receipt producer, manifest, CI wiring, binding checker, and planning-only final-candidate gate. The review checked imports and callers, manifest-to-test bindings, CLI-to-file flows, test event provenance, terminal-control attribution, and the canonical requirement ownership.

The requirement join uses all three required sources: `.planning/REQUIREMENTS.md` checklist and traceability, the six owner `VERIFICATION.md` tables, and all 42 available `*-SUMMARY.md` frontmatters recorded in `SUMMARY-EXTRACTION.json`. Fifteen requirement IDs have direct summary-frontmatter joins. CALL-01, CALL-02, and DX-24 are manually grounded in their owner verification evidence and named consumers because their historical summary frontmatter omits them; those summaries were left unchanged. EVD-17 is the expected unchecked Phase 38 owner and is not treated as an orphan.

This audit performed source inspection only. It ran no tests, producer, hosted workflow, or native matrix. Existing bounded receipts and verification reports were reused as historical evidence, not represented as newly executed checks. The source gates establish consistency and provenance checks; they do not authenticate GitHub observations.

## Cross-phase connections

| From → to | Status | Evidence and boundary |
|---|---|---|
| Phase 33 generic receipt/freeze gates → Phase 38 M007 binding | WIRED | `scripts/check-milestone-binding.sh` consumes the generic receipt and freeze checks, applies the M007 manifest, and binds candidate/digest/config/history constraints. It retains inherited M006 cases. |
| Phase 34 surface/refusal contract → Phase 35 dynamic-shift controls | WIRED | DX-22 assigns dynamic shift refusal to BIT-02; the current surface catalog and Phase 35 shift roots carry the named refusal. |
| Phase 35 bitwise operators and hex literals → Phase 37 CRC implementation | WIRED | `examples/crc32/lib/crc32.schway` uses hex constants and XOR/AND/right-shift in the reflected CRC loop; the public consumers invoke it. Phase 35 optimization evidence requires the complete value matrix at O0, O3, and O3-LTO. |
| Phase 36 scalar helper calls → Phase 37 CRC module and consumers | WIRED | The CRC loop calls `crc_byte`; `crc_low.schway` imports the shared CRC module. Helper and importer cases are separately bound to Phase 38 manifest rows. |
| Phase 37 public file-byte CLI → Phase 38 public receipt cases | WIRED | The Phase 37 file-run tracer exercises interpreter and native CLI behavior; Phase 38 binds the bounded public probe plus 52 CRC process rows across the two consumers and two engines. |
| Phase 37.1 terminal controls → Phase 38 control inventory | WIRED | Manifest roots preserve the execution-peer, C-generation, and session controls. CR-03/04/06 remain closed with current wiring; CR-04 uses checked-source `foreign.failed` attribution to the indexed generated OpFail, not synthetic OpReturn evidence. |
| Phase 38 producer and manifest → cross-platform CI aggregate | WIRED | CI invokes the M007 verifier and has Ubuntu/macOS checks and aggregate evidence jobs, with failure-preserving log/artifact upload. The checker requires full test/race and evidence-job success and rejects skipped or partial package/test event traces. This is inspected workflow code, not a hosted-run claim. |
| Phase 38 binding → final planning-only ingest candidate | WIRED | The final checker requires the actual candidate HEAD and `ingest_base_sha`, validates ancestry and all-parent paths, and limits subsequent changes to planning files. The final ingest has not run. |

The M007 manifest binds 106 process rows: 39 inherited rows kept verbatim, 52 CRC rows, and 15 helper rows. Its inventory names 17 roots, 480 exact events, and 119 reached-event expectations. The bitwise optimization root is the full `TestPhase35BitwiseOptimizationMatrix`, with all value vectors and O0/O3/O3-LTO entries; no reduced root cap is encoded. Structured helper and bitwise observations remain structured controls rather than invented decimal process rows. CRC and helper values, process output, causal events, and physical resource behavior remain distinct evidence types.

The producer has exact Go JSON package lifecycle checks, lossless event inventory checks, and refusal controls for missing/skipped/partial evidence. The ordinary M007 probe is bounded to CRC public behavior; only the later authorized hosted run can supply EVD-17. The source does not claim local full-suite execution or authenticated hosted observations.

## End-to-end flow status

| Flow | Status | Result |
|---|---|---|
| Shared CRC source → interpreter/native file CLI → pinned output | WIRED | The Phase 37 source tracer and Phase 38 manifest consumer bind this public route. Historical bounded verification is reused; no new execution here. |
| M007 fixture matrices → manifest rows → exact Go test events | WIRED | Fixture tuples are matched to distinct case/reached rows; package lifecycle, root and test events are checked for exact run/pass evidence. |
| Terminal failure controls → checked-source causal attribution | WIRED | CR-03, CR-04, and CR-06 dispositions remain attached to the named roots and exact terminal behaviors. |
| Manifest and producer → CI jobs → host-specific evidence artifacts | WIRED IN SOURCE; EXECUTION PENDING | Workflow/checker links are present. No actual Ubuntu/macOS run or artifact authentication was performed or claimed. |
| Frozen candidate → final planning-only ingest | WIRED IN SOURCE; EXECUTION PENDING | Candidate and ancestry gates are implemented. Phase 38 and the actual freeze/final ingest remain unfinished. |

There are no HTTP routes or authenticated web areas in this compiler/CLI evidence flow, so API-consumer coverage and web-auth protection are not applicable.

## Requirement integration map

| Requirement | Integration path / evidence | Status | Issue |
|---|---|---|---|
| QLT-13 | Phase 33 bounded native capture and capacity/mutant evidence → generic evidence gate | WIRED | Summary join direct. |
| QLT-14 | Phase 33 warning dispositions and tests → current source/gate refresh | WIRED | Summary join direct. |
| QLT-15 | Phase 33 prior-warning and documentation dispositions → current milestone evidence chain | WIRED | Summary join direct. |
| EVD-16 | Phase 33 milestone-agnostic gate → Phase 38 M007 manifest/binding | WIRED | Summary join direct; current M007 mode uses the generic gate. |
| DX-20 | Phase 34 text/JSON location diagnostics → downstream authoring interface | WIRED (self-contained owner surface) | Summary join direct. |
| DX-21 | Phase 34 recovery-point diagnostic behavior → authoring diagnostics | WIRED (self-contained owner surface) | Summary join direct. |
| DX-22 | Phase 34 unsupported-surface manifest → Phase 35 dynamic-shift refusal (BIT-02) | WIRED | Summary join direct; ownership remains Phase 34. |
| DX-23 | Phase 34 deterministic surface manifest → downstream refusal/consumer contract | WIRED (self-contained owner surface) | Summary join direct. |
| BIT-01 | Phase 35 bitwise semantic/optimizer matrix → Phase 37 CRC source and Phase 38 complete matrix binding | WIRED | Summary join direct. |
| BIT-02 | Phase 35 literal shift admission/refusal → Phase 34 DX-22 and Phase 37 source semantics | WIRED | Summary join direct. |
| BIT-03 | Phase 35 hex literals → Phase 37 CRC constants | WIRED | Summary join direct. |
| BIT-04 | Phase 35 checksum shift optimization → existing checksum/adler CLI receipt rows | WIRED | Summary join direct. |
| CALL-01 | Phase 36 scalar-call semantics → CRC byte helper and imported public CRC consumer | WIRED (manual history join) | Historical Phase 36 summary frontmatter omits this ID; owner verification and consumer evidence ground the join. |
| CALL-02 | Phase 36 independent call derivation and fault controls → Phase 37 helper consumer and Phase 38 helper rows | WIRED (manual history join) | Historical Phase 36 summary frontmatter omits this ID; owner verification and consumer evidence ground the join. |
| APP-13 | Phase 37 CRC source/file CLI → Phase 38 CRC cases and bounded public probe | WIRED | Summary join direct. |
| MOD-07 | Shared CRC module → second importer and receipt case | WIRED | Summary join direct. |
| APP-14 | Phase 37 interpreter/native file input parity → Phase 38 public CLI binding | WIRED | Summary join direct. |
| DX-24 | Executable CRC guide → clean-checkout guide test and Phase 38 `guide` root | WIRED (manual history join) | Historical Phase 37 summary frontmatter omits this ID; owner verification and named test ground the join. |
| EVD-17 | Phase 38 manifest/producer/checker/CI → final same-digest Ubuntu/macOS hosted receipts | PENDING — EXPECTED OWNER | The source path is prepared; no authenticated two-host run bound to the final M007 digest or final ingest exists yet. This is the sole pending requirement. |

Self-contained owner surfaces with no separate consumer phase are DX-20, DX-21, and DX-23; their absence of cross-phase wiring is expected, not an orphan. All 19 canonical requirement IDs are mapped, with 18 completed and the planned EVD-17 owner pending.

## Findings

**BLOCKER:** none found in cross-phase source wiring.

**WARNING:** EVD-17 execution remains pending by design. Phase 38 is not complete and cannot be reported frozen or milestone-complete from this audit. No additional implementation gap was established.

## Phase 38 seed dispositions

The Phase 38 transition review retains `z1g`, `trz`, and `z1s` as dormant. Current named witnesses/consumers are `QueryCommandFile` and ordinary source search (no measured navigation gap); `readApplicationCapture` and application replay (no unexplained failure); and `syntax.Format` plus checked CRC/helper source (no contested source form). No experiment or roadmap promotion is justified by the current evidence. The next useful capability priorities remain the same-source hosted M007 receipts, then the M008 `wc`/`longest_line` consumer pair, then the bounded fixed-schema key=value checker; no priority-changing observation was found in this integration trace.
