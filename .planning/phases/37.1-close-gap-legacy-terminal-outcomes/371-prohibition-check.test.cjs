"use strict";

const assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { test } = require("node:test");

const ROOT = path.resolve(__dirname, "../../..");
const PHASE_REL = ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/";
const PLAN_REL = `${PHASE_REL}37.1-01-PLAN.md`;
const FAILURE_PLAN_REL = `${PHASE_REL}37.1-02-PLAN.md`;
const REASON_PLAN_REL = `${PHASE_REL}37.1-03-PLAN.md`;
const BASELINE_PLAN_REL = `${PHASE_REL}37.1-04-PLAN.md`;
const PROJECTION_REL = `${PHASE_REL}371-PROHIBITION-PROJECTION.json`;
const RESOLUTION_REL = `${PHASE_REL}371-PROHIBITION-RESOLUTION.json`;
const ADAPTER_REL = `${PHASE_REL}371-prohibition-check.test.cjs`;
const CLEAN_REL = `${PHASE_REL}prohibition-clean.json`;
const DEFECT_REL = `${PHASE_REL}prohibition-defect-violation.json`;
const FAILURE_REL = `${PHASE_REL}prohibition-failure-violation.json`;
const REASON_REL = `${PHASE_REL}prohibition-reason-violation.json`;
const BASELINE_REL = `${PHASE_REL}prohibition-baseline-violation.json`;
const SOURCE_REL = "internal/compiler/executionpeer/executionpeer.go";
const CGEN_REL = "internal/compiler/cgen/cgen_program.go";
const PARSER_REL = "internal/compiler/syntax/parser.go";
const COREVALIDATE_REL = "internal/compiler/corevalidate/corevalidate.go";
const REASON_OVERLAY_SOURCE_RELATIVES = [CGEN_REL, PARSER_REL, COREVALIDATE_REL];
const REASON_TEST_REL = "internal/compiler/session/session_phase371_defect_reason_test.go";
const CGEN_TEST_REL = "internal/compiler/cgen/cgen_phase371_legacy_defect_test.go";
const COMPATIBILITY_TEST_REL = "internal/compiler/session/session_phase371_compatibility_test.go";
const STORAGE_EVOLUTION_PATHS = [ADAPTER_REL, CGEN_REL, CGEN_TEST_REL].sort();
const PROOF_BINDING_EVOLUTION_PATHS = [ADAPTER_REL, PARSER_REL, COREVALIDATE_REL].sort();
const TEST_REL = "internal/compiler/executionpeer/executionpeer_phase371_test.go";
const MODEL_REL = "internal/compiler/session/session_phase24_model_test.go";
const FAILURE_BASE_HEAD = "8fbd1e51bb6fd97b1a46d5d1015ce5e39d69bfe0";
const REASON_BASE_HEAD = "8f5b5795c8c31488257567511e807a276b78a10c";
const REASON_ADMISSION_BASE_HEAD = "9d09df995f1ab68c7044cda0915608d5d0a3209f";
const PACKAGE = "./internal/compiler/executionpeer";
const TMP_ROOT = "/tmp";
const GOCACHE = "/tmp/ai-lang-verification-gocache";
const MAX_OUTPUT = 8 * 1024 * 1024;
const CURRENT_PROOF_FORMAT_VERSION = 2;
const GO_TIMEOUT_MS = 120_000;
const TEST_NAME = "legacy defect closure rejects whole-execution defects";
const FAILURE_STATEMENT = "MUST NOT classify arbitrary OpReturn as propagated failure or treat causal acceptance as payload-value or actual CFG-path proof.";
const STATEMENT = "MUST NOT accept caller activity after an accepted whole-execution defect or synthesize caller unwind records.";
const CURRENT_GUARD = "\t\t\tif event.Kind == \"function.defected\" {";
const ORIGINAL_GUARD = "\t\t\tif i.dynamicCallIdentity && event.Kind == \"function.defected\" && document.Outcome.Kind == execution.OutcomeDefect {";
const ROOTS = [
  {
    name: "TestPhase371LegacyDefectClosesAllActiveFrames",
    subtests: ["legacy-static", "ordinary-nested-return", "dynamic-scalar"],
    failsOnDefect: true,
  },
  {
    name: "TestPhase371DefectTerminalControls",
    subtests: [
      "legacy-static",
      "legacy-static/source",
      "legacy-static/type",
      "legacy-static/function",
      "legacy-static/invocation",
      "legacy-static/reason",
      "legacy-static/outcome-kind",
      "legacy-static/outcome-value",
      "legacy-static/missing-terminal",
      "legacy-static/duplicate-terminal",
      "dynamic-scalar",
      "dynamic-scalar/source",
      "dynamic-scalar/type",
      "dynamic-scalar/function",
      "dynamic-scalar/invocation",
      "dynamic-scalar/reason",
      "dynamic-scalar/outcome-kind",
      "dynamic-scalar/outcome-value",
      "dynamic-scalar/missing-terminal",
      "dynamic-scalar/duplicate-terminal",
    ],
    failsOnDefect: true,
  },
  {
    name: "TestPhase33TerminalAttribution",
    subtests: [
      "return", "return/wrong_source", "return/wrong_type", "return/target_on_terminal",
      "return/missing_outcome", "return/unknown_outcome", "return/event_after_root_terminal",
      "return/outcome_typed_failure", "return/outcome_defect", "return/outcome_cancelled",
      "fail", "fail/wrong_source", "fail/wrong_type", "fail/target_on_terminal",
      "fail/missing_outcome", "fail/unknown_outcome", "fail/event_after_root_terminal",
      "fail/outcome_returned", "fail/outcome_defect", "fail/outcome_cancelled",
      "defect", "defect/wrong_source", "defect/wrong_type", "defect/target_on_terminal",
      "defect/missing_outcome", "defect/unknown_outcome", "defect/event_after_root_terminal",
      "defect/outcome_returned", "defect/outcome_typed_failure", "defect/outcome_cancelled",
      "defect/wrong_reason", "defect/defect_value",
    ],
    failsOnDefect: false,
  },
  {
    name: "TestPhase36CallCausalAttributionControls",
    subtests: [
      "ordinal-gap", "ordinal-duplicate", "ordinal-reordered", "wrong-parent", "wrong-static-site",
      "wrong-callee", "wrong-source-place", "wrong-target-place", "wrong-result-type", "missing-call",
      "missing-return", "repeated-terminal",
    ],
    failsOnDefect: false,
  },
];
const EXACT_RUN = "^(TestPhase371LegacyDefectClosesAllActiveFrames|TestPhase371DefectTerminalControls|TestPhase33TerminalAttribution|TestPhase36CallCausalAttributionControls)$";
const FAILURE_ROOTS = [
  {
    name: "TestPhase371ForeignFailureOpFailAttribution",
    subtests: ["AcquireError.AcquireFailed", "UseError.UnsupportedByte"],
    failsOnFailure: true,
  },
  {
    name: "TestPhase371ForeignFailureAttributionControls",
    subtests: [
      "nonexistent-foreign-operation", "different-foreign-operation", "wrong-function", "wrong-invocation",
      "wrong-source", "wrong-target", "wrong-type", "failure-from-different-call", "wrong-foreign-contract",
      "wrong-foreign-error-edge", "wrong-caller-source", "wrong-caller-target", "wrong-caller-type",
      "wrong-caller-error-target", "wrong-caller-error-edge", "wrong-terminal-source", "wrong-terminal-constructor",
      "wrong-terminal-operation", "wrong-root-kind", "missing-terminal", "duplicate-terminal",
      "stale-cross-activation-failure", "clean-return-relabeled-as-failure", "failure-payload-is-outside-peer-knowledge",
    ],
    failsOnFailure: true,
  },
  {
    name: "TestPhase371LegacyDefectClosesAllActiveFrames",
    subtests: ["legacy-static", "ordinary-nested-return", "dynamic-scalar"],
    failsOnFailure: false,
  },
  {
    name: "TestPhase33TerminalAttribution",
    subtests: [
      "return", "return/wrong_source", "return/wrong_type", "return/target_on_terminal",
      "return/missing_outcome", "return/unknown_outcome", "return/event_after_root_terminal",
      "return/outcome_typed_failure", "return/outcome_defect", "return/outcome_cancelled",
      "fail", "fail/wrong_source", "fail/wrong_type", "fail/target_on_terminal",
      "fail/missing_outcome", "fail/unknown_outcome", "fail/event_after_root_terminal",
      "fail/outcome_returned", "fail/outcome_defect", "fail/outcome_cancelled",
      "defect", "defect/wrong_source", "defect/wrong_type", "defect/target_on_terminal",
      "defect/missing_outcome", "defect/unknown_outcome", "defect/event_after_root_terminal",
      "defect/outcome_returned", "defect/outcome_typed_failure", "defect/outcome_cancelled",
      "defect/wrong_reason", "defect/defect_value",
    ],
    failsOnFailure: false,
  },
  {
    name: "TestPhase36CallCausalAttributionControls",
    subtests: [
      "ordinal-gap", "ordinal-duplicate", "ordinal-reordered", "wrong-parent", "wrong-static-site",
      "wrong-callee", "wrong-source-place", "wrong-target-place", "wrong-result-type", "missing-call",
      "missing-return", "repeated-terminal",
    ],
    failsOnFailure: false,
  },
];
const FAILURE_EXACT_RUN = "^(TestPhase371ForeignFailureOpFailAttribution|TestPhase371ForeignFailureAttributionControls|TestPhase371LegacyDefectClosesAllActiveFrames|TestPhase33TerminalAttribution|TestPhase36CallCausalAttributionControls)$";
const REASON_EXACT_RUN = "^TestPhase371CountedDefectReasonBytes$";
const REASON_ROOT = "TestPhase371CountedDefectReasonBytes";
const REASON_ADMISSION_ROOT = "TestPhase371DefectReasonUTF8Admission";
const REASON_STORAGE_SUBTEST = `${REASON_ROOT}/static_readonly_counted_storage`;
const BASELINE_ROOT = "TestPhase371CompatibilityAmendmentControls";
const EXECUTIONPEER_ROOTS = [
  "TestPhase371LegacyDefectClosesAllActiveFrames",
  "TestPhase371DefectTerminalControls",
  "TestPhase33TerminalAttribution",
  "TestPhase36CallCausalAttributionControls",
  "TestPhase371ForeignFailureOpFailAttribution",
  "TestPhase371ForeignFailureAttributionControls",
];
const SESSION_ROOTS = [
  "TestPhase24ModelReplayIsDeterministicAndModelOnly",
  "TestPhase371DefectReasonUTF8Admission",
  "TestPhase371PreEmitterRouteCensus",
  BASELINE_ROOT,
  "TestPhase34PreChangeCProgramGolden",
  "TestPhase34PreChangeCGoldenIsNotInert",
  "TestPhase36NoCallGeneratedCByteBaseline",
  "TestPhase36NoCallGeneratedCByteBaselineControls",
  "TestPhase37Phase34AdditionLedgerControls",
  "TestPhase37Phase36AdditionLedgerControls",
];
const CGEN_ROOTS = [
  "TestPhase371NoCallScalarDefectNativeConformance",
  REASON_ROOT,
  "TestPhase33ScalarExplicitDefectShells",
];
const MUTATED_FAILURE_NAMES = [
  "TestPhase371LegacyDefectClosesAllActiveFrames",
  "TestPhase371LegacyDefectClosesAllActiveFrames/legacy-static",
  "TestPhase371DefectTerminalControls",
  "TestPhase371DefectTerminalControls/legacy-static",
];
const LEGACY_ASSERTION = "peer rejected actual legacy child defect trace";
const FAILURE_FAILURE_NAMES = [
  "TestPhase371ForeignFailureOpFailAttribution",
  "TestPhase371ForeignFailureOpFailAttribution/AcquireError.AcquireFailed",
  "TestPhase371ForeignFailureOpFailAttribution/UseError.UnsupportedByte",
  "TestPhase371ForeignFailureAttributionControls",
];

const FIXTURES = {
  none: { relative: CLEAN_REL, data: { version: 1, subject: "none" } },
  defect: { relative: DEFECT_REL, data: { version: 1, subject: "defect" } },
  failure: { relative: FAILURE_REL, data: { version: 1, subject: "failure" } },
  reason: { relative: REASON_REL, data: { version: 1, subject: "reason" } },
  baseline: { relative: BASELINE_REL, data: { version: 1, subject: "baseline" } },
};

function digest(bytes) {
  return crypto.createHash("sha256").update(bytes).digest("hex");
}

function exactCount(text, needle) {
  return text.split(needle).length - 1;
}

function escapeRegex(text) {
  return text.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function selectSubject() {
  const raw = process.env.GSD_PROHIB_SUBJECT;
  const mode = process.env.GSD_PROHIB_MODE || "defect";
  if (!new Set(["defect", "failure", "reason", "baseline", "native", "all"]).has(mode)) return { error: `unsupported closed sample mode: ${mode}` };
  if (raw === undefined) return { subject: "none", source: "default", mode };
  for (const [subject, fixture] of Object.entries(FIXTURES)) {
    const absolute = path.resolve(ROOT, fixture.relative);
    if (raw !== fixture.relative && raw !== absolute) continue;
    try {
      const bytes = fs.readFileSync(absolute);
      if (bytes.length > 4096) return { error: `fixture exceeds 4 KiB: ${fixture.relative}` };
      const parsed = JSON.parse(bytes.toString("utf8"));
      if (JSON.stringify(parsed) !== JSON.stringify(fixture.data)) {
        return { error: `fixture differs from its closed schema: ${fixture.relative}` };
      }
      return { subject, source: fixture.relative, mode };
    } catch (error) {
      return { error: `fixture unavailable or malformed (${fixture.relative}): ${error.message}` };
    }
  }
  return { error: "GSD_PROHIB_SUBJECT is outside the closed fixture allowlist" };
}

function originalModeOverlay(sourcePath, tempDir) {
  const originalBytes = fs.readFileSync(sourcePath);
  const originalText = originalBytes.toString("utf8");
  if (exactCount(originalText, CURRENT_GUARD) !== 1) {
    throw new Error("repaired defect-closure guard count is not exactly one");
  }
  const replacementText = originalText.replace(CURRENT_GUARD, ORIGINAL_GUARD);
  const replacementPath = path.join(tempDir, "original-mode-guard.go");
  const overlayPath = path.join(tempDir, "overlay.json");
  fs.writeFileSync(replacementPath, replacementText, { flag: "wx" });
  fs.writeFileSync(overlayPath, JSON.stringify({ Replace: { [sourcePath]: replacementPath } }), { flag: "wx" });
  return {
    overlayPath,
    originalBytes,
    replacementSha256: digest(Buffer.from(replacementText)),
    overlaySha256: digest(fs.readFileSync(overlayPath)),
  };
}

function originalFailureModeOverlay(sourcePath, tempDir) {
  const currentBytes = fs.readFileSync(sourcePath);
  const baseline = spawnSync("git", ["show", `${FAILURE_BASE_HEAD}:${SOURCE_REL}`], {
    cwd: ROOT,
    encoding: "buffer",
    timeout: 10_000,
    shell: false,
  });
  if (baseline.error || baseline.status !== 0 || !Buffer.isBuffer(baseline.stdout) || baseline.stdout.length === 0) {
    throw new Error(`read fixed pre-failure peer source ${FAILURE_BASE_HEAD}: ${baseline.error || baseline.stderr || baseline.status}`);
  }
  const replacementText = baseline.stdout.toString("utf8");
  if (replacementText.includes('event.ID == id+":event" && event.Kind == "foreign.failed"')) {
    throw new Error("failure baseline unexpectedly contains the foreign.failed classifier repair");
  }
  if (!replacementText.includes('if event.Kind == "function.defected" {')) {
    throw new Error("failure baseline does not retain the completed Plan01 defect-closure guard");
  }
  const replacementPath = path.join(tempDir, "pre-failure-executionpeer.go");
  const overlayPath = path.join(tempDir, "overlay.json");
  fs.writeFileSync(replacementPath, baseline.stdout, { flag: "wx" });
  fs.writeFileSync(overlayPath, JSON.stringify({ Replace: { [sourcePath]: replacementPath } }), { flag: "wx" });
  return {
    overlayPath,
    originalBytes: currentBytes,
    replacementSha256: digest(baseline.stdout),
    overlaySha256: digest(fs.readFileSync(overlayPath)),
    baselineHead: FAILURE_BASE_HEAD,
  };
}

function originalReasonModeOverlay(tempDir) {
  const targets = [
    { relative: CGEN_REL, head: REASON_BASE_HEAD, name: "pre-reason-cgen-program.go" },
    { relative: "internal/compiler/syntax/parser.go", head: REASON_ADMISSION_BASE_HEAD, name: "pre-reason-parser.go" },
    { relative: "internal/compiler/corevalidate/corevalidate.go", head: REASON_ADMISSION_BASE_HEAD, name: "pre-reason-corevalidate.go" },
  ];
  const replacements = {};
  const sources = [];
  for (const target of targets) {
    const sourcePath = path.resolve(ROOT, target.relative);
    const currentBytes = fs.readFileSync(sourcePath);
    const baseline = spawnSync("git", ["show", `${target.head}:${target.relative}`], {
      cwd: ROOT,
      encoding: "buffer",
      timeout: 10_000,
      shell: false,
    });
    if (baseline.error || baseline.status !== 0 || !Buffer.isBuffer(baseline.stdout) || baseline.stdout.length === 0) {
      throw new Error(`read original ${target.relative} from ${target.head}: ${baseline.error || baseline.stderr || baseline.status}`);
    }
    if (target.relative === CGEN_REL) {
      const priorCgen = baseline.stdout.toString("utf8");
      if (!priorCgen.includes("if shell == programApplicationShell {")
          || !priorCgen.includes("schway_write_json_string(%s)") || !priorCgen.includes("fputs(%s, stderr)")) {
        throw new Error("reason emitter baseline does not preserve corrected Task01 shell routing and original reason helpers");
      }
    }
    const replacementPath = path.join(tempDir, target.name);
    fs.writeFileSync(replacementPath, baseline.stdout, { flag: "wx" });
    replacements[sourcePath] = replacementPath;
    sources.push({
      path: target.relative,
      originalSourceSha256: digest(currentBytes),
      replacementSourceSha256: digest(baseline.stdout),
      baselineHead: target.head,
    });
  }
  const overlayPath = path.join(tempDir, "reason-overlay.json");
  fs.writeFileSync(overlayPath, JSON.stringify({ Replace: replacements }), { flag: "wx" });
  return {
    overlayPath,
    sources,
    overlaySha256: digest(fs.readFileSync(overlayPath)),
  };
}

function originalBaselineModeOverlay(sourcePath, tempDir) {
  const originalBytes = fs.readFileSync(sourcePath);
  const originalText = originalBytes.toString("utf8");
  const anchor = "\tif !bytes.Equal(base, got) {";
  if (exactCount(originalText, anchor) !== 1) {
    throw new Error("actual old-C comparator byte check is not a unique overlay anchor");
  }
  const replacementText = originalText.replace(anchor, "\tif false && !bytes.Equal(base, got) {");
  const replacementPath = path.join(tempDir, "baseline-comparator-no-op.go");
  const overlayPath = path.join(tempDir, "overlay.json");
  fs.writeFileSync(replacementPath, replacementText, { flag: "wx" });
  fs.writeFileSync(overlayPath, JSON.stringify({ Replace: { [sourcePath]: replacementPath } }), { flag: "wx" });
  return {
    overlayPath,
    originalBytes,
    replacementSha256: digest(Buffer.from(replacementText)),
    overlaySha256: digest(fs.readFileSync(overlayPath)),
    anchorSha256: digest(Buffer.from(anchor)),
  };
}

function expectedSubtests(root, mode, original) {
  if (original && mode === "failure" && root.name === "TestPhase371ForeignFailureAttributionControls") return [];
  if (original && mode === "defect" && root.name === "TestPhase371DefectTerminalControls") {
    return root.subtests.filter((name) => !name.startsWith("legacy-static/"));
  }
  return root.subtests;
}

function exactRootNamesForMode(mode) {
  const names = mode === "failure" ? FAILURE_ROOTS.map(({ name }) => name)
    : mode === "reason" ? [REASON_ROOT, REASON_ADMISSION_ROOT]
      : mode === "baseline" ? [BASELINE_ROOT]
        : mode === "defect" ? ROOTS.map(({ name }) => name)
          : null;
  if (!names) throw new Error(`unsupported canonical producer suite: ${mode}`);
  return names.sort();
}

function canonicalControlLane(selection) {
  if (selection.source === "default") return "default";
  if (selection.source === CLEAN_REL && selection.subject === "none") return "clean";
  if (selection.subject !== "none" && FIXTURES[selection.subject]?.relative === selection.source) return "bad";
  throw new Error(`canonical producer used an unknown control subject: ${selection.source}`);
}

function recordCanonicalProducerSuiteSelection(selection, result) {
  const directory = process.env.GSD_PROHIB_PRODUCER_RECEIPT_DIR;
  if (!directory) return;

  const evidenceDir = fs.realpathSync(path.resolve(directory));
  const tempRoot = fs.realpathSync(TMP_ROOT);
  assert.ok(evidenceDir.startsWith(`${tempRoot}${path.sep}`), "canonical producer receipt directory must be under the bounded temp root");
  assert.ok(path.basename(evidenceDir).startsWith("schway371-prohibition-proof-"), "canonical producer receipt directory must use the phase proof prefix");

  const lane = canonicalControlLane(selection);
  const expectedRoots = exactRootNamesForMode(selection.mode);
  const observedRoots = Object.keys(result.roots || {}).sort();
  assert.deepEqual(observedRoots, expectedRoots, `${lane}: canonical producer must select the exact ${selection.mode} root inventory`);
  assert.equal(process.env.GSD_PROHIB_MODE, selection.mode, `${lane}: canonical producer mode must be explicit in its child environment`);

  const receipt = {
    version: 1,
    control: lane,
    subject: selection.subject,
    source: selection.source,
    mode: selection.mode,
    exact_root_inventory: observedRoots,
    semantic_result: result.kind,
    ...(selection.mode === "reason" ? { reason_overlay_binding_controls: runReasonOverlayBindingNegativeControls() } : {}),
  };
  const recordPath = path.join(evidenceDir, `canonical-suite-${lane}.json`);
  const bytes = Buffer.from(`${JSON.stringify(receipt, null, 2)}\n`);
  fs.writeFileSync(recordPath, bytes, { flag: "wx" });
}

function parseFailureName(line) {
  const match = line.match(/^--- FAIL: (.+) \([0-9]+(?:\.[0-9]+)?s\)$/);
  return match ? match[1] : null;
}

function validateGoResult(result, subject, requestedMode) {
  const original = subject !== "none";
  const mode = original ? subject : requestedMode;
  if (mode === "reason") return validateReasonGoResult(result, subject);
  const rootsForMode = mode === "failure" ? FAILURE_ROOTS : ROOTS;
  const stdout = typeof result.stdout === "string" ? result.stdout : "";
  const stderr = typeof result.stderr === "string" ? result.stderr : "";
  const output = `${stdout}\n${stderr}`;
  if (result.error) return { kind: "invalid", reason: `go spawn failed: ${result.error.code || result.error.message}`, output };
  if (result.signal) return { kind: "invalid", reason: `go was interrupted by ${result.signal}`, output };
  if (Buffer.byteLength(output) > MAX_OUTPUT) return { kind: "invalid", reason: "Go output exceeded 8 MiB", output };

  const lines = output.split(/\r?\n/);
  const failures = lines.map((line) => line.trimStart()).filter((line) => line.startsWith("--- FAIL:"));
  const failureNames = failures.map(parseFailureName);
  const roots = {};
  const expectedFailureNames = original ? (mode === "failure" ? FAILURE_FAILURE_NAMES : MUTATED_FAILURE_NAMES) : [];
  for (const root of rootsForMode) {
    const escaped = escapeRegex(root.name);
    const run = new RegExp(`^=== RUN   ${escaped}$`);
    const pass = new RegExp(`^--- PASS: ${escaped} \\([0-9]+(?:\\.[0-9]+)?s\\)$`);
    const fail = new RegExp(`^--- FAIL: ${escaped} \\([0-9]+(?:\\.[0-9]+)?s\\)$`);
    const skip = new RegExp(`^--- SKIP: ${escaped}(?:\\s|$)`);
    const runCount = lines.filter((line) => run.test(line)).length;
    const passCount = lines.filter((line) => pass.test(line.trimStart())).length;
    const failCount = lines.filter((line) => fail.test(line.trimStart())).length;
    const skipCount = lines.filter((line) => skip.test(line.trimStart())).length;
    const shouldFail = original && (mode === "failure" ? root.failsOnFailure : root.failsOnDefect);
    if (runCount !== 1 || passCount !== (shouldFail ? 0 : 1) || failCount !== (shouldFail ? 1 : 0) || skipCount !== 0) {
      return { kind: "invalid", reason: `${root.name} exact root count/result mismatch`, output };
    }
    const subtests = {};
    for (const subtest of expectedSubtests(root, mode, original)) {
      const fullName = `${root.name}/${subtest}`;
      const subtestRun = new RegExp(`^=== RUN   ${escapeRegex(fullName)}$`);
      const subtestPass = new RegExp(`^--- PASS: ${escapeRegex(fullName)} \\([0-9]+(?:\\.[0-9]+)?s\\)$`);
      const subtestFail = new RegExp(`^--- FAIL: ${escapeRegex(fullName)} \\([0-9]+(?:\\.[0-9]+)?s\\)$`);
      const subtestSkip = new RegExp(`^--- SKIP: ${escapeRegex(fullName)}(?:\\s|$)`);
      const shouldSubtestFail = original && (mode === "defect"
        ? fullName === MUTATED_FAILURE_NAMES[1] || fullName === MUTATED_FAILURE_NAMES[3]
        : fullName === FAILURE_FAILURE_NAMES[1] || fullName === FAILURE_FAILURE_NAMES[2]);
      const runCountSubtest = lines.filter((line) => subtestRun.test(line)).length;
      const passCountSubtest = lines.filter((line) => subtestPass.test(line.trimStart())).length;
      const failCountSubtest = lines.filter((line) => subtestFail.test(line.trimStart())).length;
      const skipCountSubtest = lines.filter((line) => subtestSkip.test(line.trimStart())).length;
      if (runCountSubtest !== 1 || passCountSubtest !== (shouldSubtestFail ? 0 : 1)
          || failCountSubtest !== (shouldSubtestFail ? 1 : 0) || skipCountSubtest !== 0) {
        return { kind: "invalid", reason: `${fullName} subtest count/result mismatch`, output };
      }
      subtests[subtest] = { run: runCountSubtest, pass: passCountSubtest, fail: failCountSubtest, skip: skipCountSubtest };
    }
    roots[root.name] = {
      run: runCount,
      pass: passCount,
      fail: failCount,
      skip: skipCount,
      subtestCount: Object.keys(subtests).length,
      subtests,
    };
  }
  if (/^\s*--- SKIP:/m.test(output)) return { kind: "invalid", reason: "a required Go root or subtest was skipped", output };
  if (failureNames.some((name) => name == null)
      || failureNames.length !== expectedFailureNames.length
      || expectedFailureNames.some((name) => failureNames.filter((found) => found === name).length !== 1)) {
    return { kind: "invalid", reason: "Go run failed outside the exact original-mode semantic assertions", output };
  }

  const expectedStatus = original ? 1 : 0;
  if (result.status !== expectedStatus) return { kind: "invalid", reason: `Go exit status ${result.status}, expected ${expectedStatus}`, output };
  if (original && mode === "defect") {
    if (exactCount(output, LEGACY_ASSERTION) !== 2 || exactCount(output, "activation has no terminal event") !== 2) {
      return { kind: "invalid", reason: "original-mode failures missed their exact child-defect semantic assertions", output };
    }
    return { kind: "violation", output, roots, failureNames };
  }
  if (original && mode === "failure") {
    if (exactCount(output, "peer rejected checked") !== 2 || exactCount(output, "executionpeer.unknown_kind") !== 3
        || exactCount(output, "fresh UseError child-to-root failure rejected") !== 1) {
      return { kind: "invalid", reason: "pre-failure peer run missed its exact foreign-failure semantic assertions", output };
    }
    return { kind: "violation", output, roots, failureNames };
  }
  if (/^--- FAIL:|^FAIL(?:\s|$)/m.test(output)) return { kind: "invalid", reason: "unmutated exact-root command failed", output };
  return { kind: "green", output, roots, failureNames };
}

function validateReasonGoResult(result, subject) {
  const output = `${typeof result.stdout === "string" ? result.stdout : ""}\n${typeof result.stderr === "string" ? result.stderr : ""}`;
  if (result.error) return { kind: "invalid", reason: `go spawn failed: ${result.error.code || result.error.message}`, output };
  if (result.signal) return { kind: "invalid", reason: `go was interrupted by ${result.signal}`, output };
  if (Buffer.byteLength(output) > MAX_OUTPUT) return { kind: "invalid", reason: "Go output exceeded 8 MiB", output };
  const runCount = exactCount(output, `=== RUN   ${REASON_ROOT}`);
  const passCount = exactCount(output, `--- PASS: ${REASON_ROOT} (`);
  const failCount = exactCount(output, `--- FAIL: ${REASON_ROOT} (`);
  if (runCount !== 1 || /^--- SKIP:/m.test(output)) return { kind: "invalid", reason: "reason root missing, duplicated or skipped", output };
  if (subject === "reason") {
    if (result.status !== 1 || passCount !== 0 || failCount !== 1
        || exactCount(output, `=== RUN   ${REASON_STORAGE_SUBTEST}`) !== 1
        || exactCount(output, `--- FAIL: ${REASON_STORAGE_SUBTEST} (`) !== 1
        || !output.includes("generated native and application counted reason buffers must be static readonly without automatic compound literals")) {
      return { kind: "invalid", reason: "original reason emitter missed the exact static-storage invariant", output };
    }
    if (/syntax\.|undefined reference|no tests to run|timed out|^FAIL\s/m.test(output)) {
      return { kind: "invalid", reason: "reason RED was replaced by a syntax/tool/timeout or missing-root failure", output };
    }
    return { kind: "violation", output, roots: { [REASON_ROOT]: { run: 1, pass: 0, fail: 1, skip: 0 } }, failureNames: [REASON_ROOT] };
  }
  if (result.status !== 0 || passCount !== 1 || failCount !== 0 || /^--- FAIL:|^FAIL(?:\s|$)/m.test(output)) {
    return { kind: "invalid", reason: "clean reason test root did not pass exactly once", output };
  }
  return { kind: "green", output, roots: { [REASON_ROOT]: { run: 1, pass: 1, fail: 0, skip: 0 } }, failureNames: [] };
}

function runReasonExactRoots(subject) {
  const original = subject === "reason";
  const roots = [
    { name: REASON_ROOT, package: "./internal/compiler/cgen", run: REASON_EXACT_RUN,
      requiredOnRed: [`--- FAIL: ${REASON_STORAGE_SUBTEST} (`, "generated native and application counted reason buffers must be static readonly without automatic compound literals"] },
    { name: "TestPhase371DefectReasonUTF8Admission", package: "./internal/compiler/session",
      run: "^TestPhase371DefectReasonUTF8Admission$",
      requiredOnRed: ["missing stable syntax.invalid_utf8 byte span", "raw-core malformed reason was not independently refused as core.final_claim_mismatch"] },
  ];
  let tempDir = null;
  let overlay = null;
  const outputParts = [];
  const observed = {};
  try {
    if (original) {
      tempDir = fs.mkdtempSync(path.join(TMP_ROOT, "schway371-reason-overlay-"));
      overlay = originalReasonModeOverlay(tempDir);
    }
    const env = { ...process.env, GOCACHE, GOFLAGS: "" };
    delete env.GSD_PROHIB_SUBJECT;
    for (const root of roots) {
      const args = ["test"];
      if (overlay) args.push("-overlay", overlay.overlayPath);
      args.push(root.package, "-run", root.run, "-count=1", "-v", "-timeout=120s");
      const result = spawnSync("go", args, {
        cwd: ROOT,
        env,
        encoding: "utf8",
        timeout: GO_TIMEOUT_MS,
        maxBuffer: MAX_OUTPUT,
        stdio: ["ignore", "pipe", "pipe"],
        shell: false,
        windowsHide: true,
      });
      const output = `${typeof result.stdout === "string" ? result.stdout : ""}${typeof result.stderr === "string" ? result.stderr : ""}`;
      outputParts.push(`=== EXACT ROOT ${root.name} ===\n${output}`);
      if (result.error || result.signal || Buffer.byteLength(output) > MAX_OUTPUT || /^--- SKIP:/m.test(output)) {
        return { kind: "invalid", reason: `${root.name}: Go spawn, timeout, output bound or skip failure`, output: outputParts.join("\n") };
      }
      const runCount = [...output.matchAll(new RegExp(`^=== RUN   ${escapeRegex(root.name)}$`, "gm"))].length;
      const passCount = [...output.matchAll(new RegExp(`^--- PASS: ${escapeRegex(root.name)} \\([0-9]+(?:\\.[0-9]+)?s\\)$`, "gm"))].length;
      const failCount = [...output.matchAll(new RegExp(`^--- FAIL: ${escapeRegex(root.name)} \\([0-9]+(?:\\.[0-9]+)?s\\)$`, "gm"))].length;
      if (runCount !== 1) return { kind: "invalid", reason: `${root.name}: exact RUN count was ${runCount}`, output: outputParts.join("\n") };
      if (original) {
        if (result.status !== 1 || passCount !== 0 || failCount !== 1 || root.requiredOnRed.some((needle) => !output.includes(needle))) {
          return { kind: "invalid", reason: `${root.name}: original source missed its named semantic RED`, output: outputParts.join("\n") };
        }
        if (/\[build failed\]|undefined reference|no tests to run|timed out|syntax\.unterminated/m.test(output)) {
          return { kind: "invalid", reason: `${root.name}: semantic RED was replaced by a build/tool/syntax failure`, output: outputParts.join("\n") };
        }
        observed[root.name] = { run: 1, pass: 0, fail: 1, skip: 0 };
      } else {
        if (result.status !== 0 || passCount !== 1 || failCount !== 0 || /^--- FAIL:|^FAIL(?:\s|$)/m.test(output)) {
          return { kind: "invalid", reason: `${root.name}: clean exact root did not pass once`, output: outputParts.join("\n") };
        }
        observed[root.name] = { run: 1, pass: 1, fail: 0, skip: 0 };
      }
    }
    if (overlay) {
      for (const source of overlay.sources) {
        const current = fs.readFileSync(path.join(ROOT, source.path));
        if (digest(current) !== source.originalSourceSha256) {
          return { kind: "invalid", reason: `tracked ${source.path} changed during exact-source overlay run`, output: outputParts.join("\n") };
        }
      }
    }
    return {
      kind: original ? "violation" : "green",
      output: outputParts.join("\n"),
      roots: observed,
      failureNames: original ? Object.keys(observed) : [],
      ...(overlay ? { overlay: {
        originalSourceSha256: digest(Buffer.from(JSON.stringify(overlay.sources.map(({ path: name, originalSourceSha256 }) => ({ path: name, sha256: originalSourceSha256 }))))),
        replacementSourceSha256: digest(Buffer.from(JSON.stringify(overlay.sources.map(({ path: name, replacementSourceSha256 }) => ({ path: name, sha256: replacementSourceSha256 }))))),
        overlayJsonSha256: overlay.overlaySha256,
        sourceUnchanged: true,
        baselineHead: overlay.sources.map((source) => `${source.path}@${source.baselineHead}`).join(","),
        sources: overlay.sources,
      } } : {}),
    };
  } catch (error) {
    return { kind: "invalid", reason: `reason exact-root setup failed: ${error.message}`, output: outputParts.join("\n") };
  } finally {
    if (tempDir) fs.rmSync(tempDir, { recursive: true, force: true });
  }
}

function runBaselineExactRoot(subject) {
  const original = subject === "baseline";
  const sourcePath = path.resolve(ROOT, COMPATIBILITY_TEST_REL);
  const originalBytes = fs.readFileSync(sourcePath);
  let tempDir = null;
  let overlay = null;
  try {
    if (original) {
      tempDir = fs.mkdtempSync(path.join(TMP_ROOT, "schway371-baseline-overlay-"));
      overlay = originalBaselineModeOverlay(sourcePath, tempDir);
    }
    const args = ["test"];
    if (overlay) args.push("-overlay", overlay.overlayPath);
    args.push("./internal/compiler/session", "-run", `^${BASELINE_ROOT}$`, "-count=1", "-v", "-timeout=120s");
    const result = spawnSync("go", args, {
      cwd: ROOT,
      env: { ...process.env, GOCACHE, GOFLAGS: "" },
      encoding: "utf8",
      timeout: GO_TIMEOUT_MS,
      maxBuffer: MAX_OUTPUT,
      stdio: ["ignore", "pipe", "pipe"],
      shell: false,
      windowsHide: true,
    });
    const output = `${typeof result.stdout === "string" ? result.stdout : ""}${typeof result.stderr === "string" ? result.stderr : ""}`;
    if (result.error || result.signal || Buffer.byteLength(output) > MAX_OUTPUT || /^--- SKIP:/m.test(output)) {
      return { kind: "invalid", reason: "baseline exact-root Go spawn, timeout, output bound or skip failure", output };
    }
    const rootRun = [...output.matchAll(new RegExp(`^=== RUN   ${escapeRegex(BASELINE_ROOT)}$`, "gm"))].length;
    const rootPass = [...output.matchAll(new RegExp(`^--- PASS: ${escapeRegex(BASELINE_ROOT)} \\([0-9]+(?:\\.[0-9]+)?s\\)$`, "gm"))].length;
    const rootFail = [...output.matchAll(new RegExp(`^--- FAIL: ${escapeRegex(BASELINE_ROOT)} \\([0-9]+(?:\\.[0-9]+)?s\\)$`, "gm"))].length;
    const lines = output.split(/\r?\n/);
    const subRunPrefix = `=== RUN   ${BASELINE_ROOT}/`;
    const subPassPrefix = `--- PASS: ${BASELINE_ROOT}/`;
    const subFailPrefix = `--- FAIL: ${BASELINE_ROOT}/`;
    const stripDuration = (value) => value.replace(/ \([0-9]+(?:\.[0-9]+)?s\)$/, "");
    const subRuns = lines.filter((line) => line.startsWith(subRunPrefix)).map((line) => line.slice(subRunPrefix.length));
    const subPasses = lines.map((line) => line.trimStart()).filter((line) => line.startsWith(subPassPrefix)).map((line) => stripDuration(line.slice(subPassPrefix.length)));
    const subFails = lines.map((line) => line.trimStart()).filter((line) => line.startsWith(subFailPrefix)).map((line) => stripDuration(line.slice(subFailPrefix.length)));
    if (rootRun !== 1 || subRuns.length !== new Set(subRuns).size || subRuns.length !== subPasses.length + subFails.length) {
      return { kind: "invalid", reason: `baseline root/subtest counts: root=${rootRun} subrun=${subRuns.length}/${new Set(subRuns).size} subpass=${subPasses.length} subfail=${subFails.length}; tail=${output.slice(-1000)}`, output };
    }
    if (original) {
      if (result.status !== 1 || rootPass !== 0 || rootFail !== 1 || subFails.length !== 0
          || !output.includes("unrelated extra C-byte mutation was accepted by the actual comparator")) {
        return { kind: "invalid", reason: "no-op comparator overlay missed the exact unrelated-C-byte semantic RED", output };
      }
      if (/\\[build failed\\]|undefined reference|no tests to run|timed out/m.test(output)) {
        return { kind: "invalid", reason: "baseline RED was replaced by a build/tool/timeout failure", output };
      }
    } else if (result.status !== 0 || rootPass !== 1 || rootFail !== 0 || subFails.length !== 0 || /^--- FAIL:|^FAIL(?:\\s|$)/m.test(output)) {
      return { kind: "invalid", reason: "clean baseline root or a required subtest failed", output };
    }
    const afterBytes = fs.readFileSync(sourcePath);
    if (!afterBytes.equals(originalBytes)) return { kind: "invalid", reason: "tracked amendment comparator source changed during overlay run", output };
    const counts = { run: 1, pass: rootPass, fail: rootFail, skip: 0, subtestCount: subRuns.length, subtestPass: subPasses.length, subtestFail: subFails.length };
    return {
      kind: original ? "violation" : "green",
      output,
      roots: { [BASELINE_ROOT]: counts },
      failureNames: original ? [BASELINE_ROOT] : [],
      ...(overlay ? { overlay: {
        originalSourceSha256: digest(originalBytes),
        replacementSourceSha256: overlay.replacementSha256,
        overlayJsonSha256: overlay.overlaySha256,
        anchorSha256: overlay.anchorSha256,
        sourceUnchanged: true,
      } } : {}),
    };
  } catch (error) {
    return { kind: "invalid", reason: `baseline exact-root setup failed: ${error.message}` };
  } finally {
    if (tempDir) fs.rmSync(tempDir, { recursive: true, force: true });
  }
}

function runRootInventory(kind) {
  const selectedPackages = kind === "native"
    ? [{ name: "./internal/compiler/cgen", roots: CGEN_ROOTS, timeout: GO_TIMEOUT_MS }]
    : [
      { name: PACKAGE, roots: EXECUTIONPEER_ROOTS, timeout: GO_TIMEOUT_MS },
      { name: "./internal/compiler/session", roots: SESSION_ROOTS, timeout: 600_000 },
      { name: "./internal/compiler/cgen", roots: CGEN_ROOTS, timeout: GO_TIMEOUT_MS },
    ];
  const expectedRootCount = kind === "native" ? 3 : 19;
  const allRootNames = selectedPackages.flatMap((item) => item.roots);
  if (allRootNames.length !== expectedRootCount || new Set(allRootNames).size !== expectedRootCount) {
    return { kind: "invalid", reason: `${kind} declared root union is not a unique ${expectedRootCount}-root inventory` };
  }
  const outputParts = [];
  const observed = {};
  for (const selected of selectedPackages) {
    const runPattern = `^(${selected.roots.map(escapeRegex).join("|")})$`;
    const result = spawnSync("go", ["test", selected.name, "-run", runPattern, "-count=1", "-v", `-timeout=${Math.floor(selected.timeout / 1000)}s`], {
      cwd: ROOT,
      env: { ...process.env, GOCACHE, GOFLAGS: "" },
      encoding: "utf8",
      timeout: selected.timeout,
      maxBuffer: MAX_OUTPUT,
      stdio: ["ignore", "pipe", "pipe"],
      shell: false,
      windowsHide: true,
    });
    const output = `${typeof result.stdout === "string" ? result.stdout : ""}${typeof result.stderr === "string" ? result.stderr : ""}`;
    outputParts.push(`=== PACKAGE ${selected.name} ===\n${output}`);
    if (result.error || result.signal || result.status !== 0 || Buffer.byteLength(output) > MAX_OUTPUT
        || /^--- SKIP:/m.test(output) || /no tests to run|timed out/i.test(output)) {
      return { kind: "invalid", reason: `${selected.name} bounded root command failed, skipped or exceeded its limits`, output: outputParts.join("\n") };
    }
    const lines = output.split(/\r?\n/);
    const topRuns = [...output.matchAll(/^=== RUN   ([^/\s]+)$/gm)].map((match) => match[1]);
    const expectedTop = new Set(selected.roots);
    if (topRuns.length !== expectedRootCountForPackage(selected) || new Set(topRuns).size !== topRuns.length
        || topRuns.some((name) => !expectedTop.has(name))) {
      return { kind: "invalid", reason: `${selected.name} executed a missing, duplicate or unregistered root`, output: outputParts.join("\n") };
    }
    for (const root of selected.roots) {
      const rootRun = [...output.matchAll(new RegExp(`^=== RUN   ${escapeRegex(root)}$`, "gm"))].length;
      const rootPass = [...output.matchAll(new RegExp(`^--- PASS: ${escapeRegex(root)} \\([0-9]+(?:\\.[0-9]+)?s\\)$`, "gm"))].length;
      const rootFail = [...output.matchAll(new RegExp(`^--- FAIL: ${escapeRegex(root)} \\([0-9]+(?:\\.[0-9]+)?s\\)$`, "gm"))].length;
      const subRunPrefix = `=== RUN   ${root}/`;
      const subPassPrefix = `--- PASS: ${root}/`;
      const subFailPrefix = `--- FAIL: ${root}/`;
      const subSkipPrefix = `--- SKIP: ${root}/`;
      const stripDuration = (value) => value.replace(/ \([0-9]+(?:\.[0-9]+)?s\)$/, "");
      const subRuns = lines.filter((line) => line.startsWith(subRunPrefix)).map((line) => line.slice(subRunPrefix.length));
      const subPasses = lines.map((line) => line.trimStart()).filter((line) => line.startsWith(subPassPrefix)).map((line) => stripDuration(line.slice(subPassPrefix.length)));
      const subFails = lines.map((line) => line.trimStart()).filter((line) => line.startsWith(subFailPrefix)).map((line) => stripDuration(line.slice(subFailPrefix.length)));
      const subSkips = lines.map((line) => line.trimStart()).filter((line) => line.startsWith(subSkipPrefix));
      if (rootRun !== 1 || rootPass !== 1 || rootFail !== 0
          || new Set(subRuns).size !== subRuns.length || subRuns.length !== subPasses.length + subFails.length
          || subFails.length !== 0 || subSkips.length !== 0) {
        return { kind: "invalid", reason: `${root} exact root/subtest run/pass counts did not match`, output: outputParts.join("\n") };
      }
      observed[root] = {
        package: selected.name, run: rootRun, pass: rootPass, fail: rootFail, skip: 0,
        subtestCount: subRuns.length, subtestPass: subPasses.length, subtestFail: subFails.length, subtestSkip: 0,
      };
    }
    if (/^--- FAIL:|^FAIL(?:\s|$)/m.test(output)) {
      return { kind: "invalid", reason: `${selected.name} emitted an unexpected failure despite a zero exit`, output: outputParts.join("\n") };
    }
  }
  return { kind: "green", output: outputParts.join("\n"), roots: observed, failureNames: [] };
}

function expectedRootCountForPackage(selected) {
  return selected.roots.length;
}

function runExactRoots(subject, requestedMode = "defect") {
  const mode = subject === "none" ? requestedMode : subject;
  if (!new Set(["defect", "failure", "reason", "baseline"]).has(mode)) throw new Error(`unsupported exact-root suite: ${mode}`);
  if (mode === "reason") return runReasonExactRoots(subject);
  if (mode === "baseline") return runBaselineExactRoot(subject);
  const original = subject !== "none";
  const sourcePath = path.resolve(ROOT, mode === "reason" ? CGEN_REL : SOURCE_REL);
  const originalBytes = fs.readFileSync(sourcePath);
  let tempDir = null;
  let overlay = null;
  let captured;
  let cleanupError = "";
  try {
    const args = ["test"];
    if (original) {
      tempDir = fs.mkdtempSync(path.join(TMP_ROOT, `schway371-${mode}-overlay-`));
      overlay = mode === "defect" ? originalModeOverlay(sourcePath, tempDir)
        : mode === "failure" ? originalFailureModeOverlay(sourcePath, tempDir)
          : originalReasonModeOverlay(sourcePath, tempDir);
      args.push("-overlay", overlay.overlayPath);
    }
    const packageName = mode === "reason" ? "./internal/compiler/cgen" : PACKAGE;
    const exactRun = mode === "failure" ? FAILURE_EXACT_RUN : mode === "reason" ? REASON_EXACT_RUN : EXACT_RUN;
    args.push(packageName, "-run", exactRun, "-count=1", "-v", "-timeout=120s");
    const env = { ...process.env, GOCACHE, GOFLAGS: "" };
    delete env.GSD_PROHIB_SUBJECT;
    const result = spawnSync("go", args, {
      cwd: ROOT,
      env,
      encoding: "utf8",
      timeout: GO_TIMEOUT_MS,
      maxBuffer: MAX_OUTPUT,
      stdio: ["ignore", "pipe", "pipe"],
      shell: false,
      windowsHide: true,
    });
    captured = validateGoResult(result, subject, mode);
    const afterBytes = fs.readFileSync(sourcePath);
    if (!afterBytes.equals(originalBytes)) captured = { kind: "invalid", reason: `tracked ${mode === "reason" ? CGEN_REL : SOURCE_REL} changed during overlay run`, output: captured.output };
    if (original && overlay) {
      captured.overlay = {
        originalSourceSha256: digest(originalBytes),
        replacementSourceSha256: overlay.replacementSha256,
        overlayJsonSha256: overlay.overlaySha256,
        sourceUnchanged: afterBytes.equals(originalBytes),
        ...(overlay.baselineHead ? { baselineHead: overlay.baselineHead } : {}),
      };
    }
  } catch (error) {
    captured = { kind: "invalid", reason: `exact-root execution or overlay setup failed: ${error.message}` };
  } finally {
    if (tempDir) {
      try {
        fs.rmSync(tempDir, { recursive: true, force: true });
      } catch (error) {
        cleanupError = error.message;
      }
    }
  }
  if (cleanupError) return { kind: "invalid", reason: `temporary overlay cleanup failed: ${cleanupError}`, output: captured && captured.output };
  return captured || { kind: "invalid", reason: "Go runner produced no result" };
}

function loadGsdProjectionApi() {
  const lib = path.join(os.homedir(), ".codex", "gsd-core", "bin", "lib");
  const { projectProhibitions } = require(path.join(lib, "probe-core.cjs"));
  const { descriptorFromProjection } = require(path.join(lib, "prohibition-enforcement.cjs"));
  assert.equal(typeof projectProhibitions, "function", "installed projectProhibitions API must be callable");
  assert.equal(typeof descriptorFromProjection, "function", "installed descriptorFromProjection API must be callable");
  return { projectProhibitions, descriptorFromProjection };
}

function parsePlanProhibition(planRelative = PLAN_REL, expectedStatement = STATEMENT, sourceText = undefined) {
  const planPath = path.join(ROOT, planRelative);
  const planText = sourceText === undefined ? fs.readFileSync(planPath, "utf8") : sourceText;
  const match = planText.match(/^  prohibitions:\n([\s\S]*?)(?=^  artifacts:)/m);
  assert.ok(match, `${planRelative} must retain its prohibition block`);
  assert.ok(!/^\s{6}check:\s*$/m.test(match[1]), "plan descriptors must remain flat check_* scalars");
  let current = null;
  const items = [];
  for (const line of match[1].split(/\r?\n/)) {
    const start = line.match(/^    - statement: "(.*)"$/);
    if (start) {
      if (current) items.push(current);
      current = { statement: start[1] };
      continue;
    }
    const field = line.match(/^      ([a-z_]+): (.+)$/);
    if (!field || !current) continue;
    let value = field[2];
    if (value === "null") value = null;
    else if (value.startsWith("\"") && value.endsWith("\"")) value = value.slice(1, -1);
    current[field[1]] = value;
  }
  if (current) items.push(current);
  assert.equal(items.length, 1, `${planRelative} must carry exactly one canonical prohibition`);
  assert.equal(items[0].statement, expectedStatement);
  const unresolved = items[0].status === "unresolved" && items[0].verification === null;
  const resolvedTest = items[0].status === "resolved" && items[0].verification === "test";
  assert.ok(unresolved || resolvedTest, "plan prohibition lifecycle must be unresolved/null or resolved/test");
  return { item: items[0], planBytes: Buffer.from(planText), planPath };
}

function projectCurrentPlan(mode = "defect", sourceText = undefined) {
  const { projectProhibitions, descriptorFromProjection } = loadGsdProjectionApi();
  const isFailure = mode === "failure";
  const isReason = mode === "reason";
  const isBaseline = mode === "baseline";
  const planRelative = isFailure ? FAILURE_PLAN_REL : isReason ? REASON_PLAN_REL : isBaseline ? BASELINE_PLAN_REL : PLAN_REL;
  const expectedStatement = isFailure ? FAILURE_STATEMENT : isReason
    ? "MUST NOT normalize malformed defect-reason bytes, reject valid UTF-8 control/NUL reasons, decode textual escapes or truncate admitted reason output."
    : isBaseline ? "MUST NOT regenerate frozen C expectations, exempt a whole path/route family or accept unrelated bytes through an amendment."
      : STATEMENT;
  const { item, planBytes, planPath } = parsePlanProhibition(planRelative, expectedStatement, sourceText);
  const projection = projectProhibitions([item]);
  assert.equal(projection.length, 1, "projectProhibitions must preserve one live item");
  const projected = projection[0];
  assert.equal(projected.status, item.status, "canonical serializer must preserve current lifecycle status");
  if (item.status === "unresolved") {
    assert.equal(projected.verification, undefined, "null verification must stay absent in canonical projection");
  } else {
    assert.equal(projected.verification, "test", "canonical serializer must preserve test-tier resolution");
  }
  const check = descriptorFromProjection(projected);
  assert.deepEqual(check, {
    kind: "node-test",
    target: ADAPTER_REL,
    violationFixture: isFailure ? FAILURE_REL : isReason ? REASON_REL : isBaseline ? BASELINE_REL : DEFECT_REL,
    cleanFixture: CLEAN_REL,
  }, "descriptorFromProjection must reconstruct this plan's wired flat fields");
  return { item, projected, check, planBytes, planPath, planRelative };
}

function gitHead() {
  const result = spawnSync("git", ["rev-parse", "HEAD"], { cwd: ROOT, encoding: "utf8", shell: false });
  assert.equal(result.status, 0, `capture initial projection HEAD: ${result.stderr || result.stdout}`);
  return result.stdout.trim();
}

function gitFileAt(revision, relative) {
  const result = spawnSync("git", ["show", `${revision}:${relative}`], { cwd: ROOT, encoding: null, shell: false });
  assert.equal(result.status, 0, `read historical source ${revision}:${relative}: ${result.stderr || "git show failed"}`);
  return Buffer.from(result.stdout);
}

function ensureInitialProjection() {
  const projectionPath = path.join(ROOT, PROJECTION_REL);
  if (fs.existsSync(projectionPath)) {
    const stored = JSON.parse(fs.readFileSync(projectionPath, "utf8"));
    assert.equal(stored.source_plan.path, PLAN_REL, "initial source plan path must remain the original plan");
    const sourceBytes = gitFileAt(stored.source_head, stored.source_plan.path);
    assert.equal(stored.source_plan.sha256, digest(sourceBytes), "initial projection must bind the original plan bytes at source_head");
    const original = parsePlanProhibition(PLAN_REL, STATEMENT, sourceBytes.toString("utf8"));
    assert.equal(original.item.status, "unresolved", "initial source snapshot must retain unresolved status");
    assert.equal(original.item.verification, null, "initial source snapshot must retain null verification");
    const { projectProhibitions, descriptorFromProjection } = loadGsdProjectionApi();
    const projected = projectProhibitions([original.item]);
    assert.deepEqual(stored.projection, projected, "preserved initial canonical unresolved projection must match historical source bytes");
    assert.deepEqual(stored.source_prohibition, original.item, "preserved initial source status and null verification must match historical source bytes");
    assert.deepEqual(descriptorFromProjection(stored.projection[0]), {
      kind: "node-test", target: ADAPTER_REL, violationFixture: DEFECT_REL, cleanFixture: CLEAN_REL,
    }, "initial historical projection must retain the original canonical descriptor");
    const live = projectCurrentPlan("defect");
    assert.equal(live.item.statement, stored.source_prohibition.statement, "live transition must retain the historical prohibition statement");
    assert.deepEqual(live.check, descriptorFromProjection(stored.projection[0]), "live transition must preserve the original descriptor");
    assert.equal(stored.source_plan.path, PLAN_REL);
    return stored;
  }
  const current = projectCurrentPlan();
  assert.equal(current.item.status, "unresolved", "initial projection can only be captured before lifecycle resolution");
  const stored = {
    version: 1,
    source_head: gitHead(),
    captured_utc: new Date().toISOString(),
    source_plan: { path: PLAN_REL, sha256: digest(current.planBytes) },
    source_prohibition: current.item,
    projection: [current.projected],
  };
  fs.writeFileSync(projectionPath, `${JSON.stringify(stored, null, 2)}\n`, { flag: "wx" });
  return stored;
}

function saveRaw(directory, name, value) {
  const filePath = path.join(directory, name);
  const bytes = Buffer.isBuffer(value) ? value : Buffer.from(value);
  fs.writeFileSync(filePath, bytes, { flag: "wx" });
  return { path: filePath, sha256: digest(bytes), bytes: bytes.length };
}

function tempEvidenceDir() {
  return fs.mkdtempSync(path.join(TMP_ROOT, "schway371-prohibition-proof-"));
}

function runSampleMode(subject) {
  if (subject === "native" || subject === "all") {
    const result = runRootInventory(subject);
    if (result.kind !== "green") throw new Error(`${subject} root inventory failed: ${result.reason || result.kind}`);
    const initial = ensureInitialProjection();
    const current = projectCurrentPlan("baseline");
    const directory = tempEvidenceDir();
    const log = saveRaw(directory, `${subject}-go.log`, result.output);
    const manifest = {
      mode: `sample-${subject}`,
      source_head: initial.source_head,
      initial_projection_sha256: digest(Buffer.from(JSON.stringify(initial.projection))),
      current_plan: current.planRelative,
      current_plan_sha256: digest(current.planBytes),
      root_count: Object.keys(result.roots).length,
      roots: result.roots,
      source_bindings: sourceBindings(current),
      raw_log: log,
    };
    saveRaw(directory, "sample-manifest.json", `${JSON.stringify(manifest, null, 2)}\n`);
    process.stdout.write(`${subject} bounded root inventory passed (${Object.keys(result.roots).length} distinct roots); evidence ${directory}\n`);
    return;
  }
  if (subject !== "defect" && subject !== "failure" && subject !== "reason" && subject !== "baseline") throw new Error("unsupported closed sample mode");
  const initial = ensureInitialProjection();
  const current = projectCurrentPlan(subject);
  const result = runExactRoots(subject, subject);
  if (result.kind !== "violation") throw new Error(`${subject} sample did not reach the expected semantic RED: ${result.reason || result.kind}`);
  const directory = tempEvidenceDir();
  const log = saveRaw(directory, `${subject}-go.log`, result.output);
  const manifest = {
    mode: `sample-${subject}`,
    source_head: initial.source_head,
    projection_sha256: digest(Buffer.from(JSON.stringify(initial.projection))),
    current_plan: current.planRelative,
    current_plan_sha256: digest(current.planBytes),
    source_sha256: result.overlay.originalSourceSha256,
    replacement_sha256: result.overlay.replacementSourceSha256,
    overlay_sha256: result.overlay.overlayJsonSha256,
    source_unchanged: result.overlay.sourceUnchanged,
    ...(result.overlay.baselineHead ? { baseline_head: result.overlay.baselineHead } : {}),
    roots: result.roots,
    failures: result.failureNames,
    raw_log: log,
  };
  saveRaw(directory, "sample-manifest.json", `${JSON.stringify(manifest, null, 2)}\n`);
  process.stdout.write(`${subject} sample passed: exact original-source semantic REDs reached; evidence ${directory}\n`);
}

function parseTapCounts(output) {
  const value = (name) => {
    const match = output.match(new RegExp(`^# ${name} (\\d+)$`, "m"));
    return match ? Number(match[1]) : -1;
  };
  return { tests: value("tests"), pass: value("pass"), fail: value("fail"), cancelled: value("cancelled"), skipped: value("skipped"), todo: value("todo") };
}

function runDirectTap(mode, sample = "defect") {
  const env = { ...process.env, GOCACHE, GOFLAGS: "" };
  delete env.NODE_TEST_CONTEXT;
  delete env.NODE_OPTIONS;
  delete env.GSD_PROHIB_SUBJECT;
  env.GSD_PROHIB_MODE = sample;
  if (mode === "bad") env.GSD_PROHIB_SUBJECT = FIXTURES[sample].relative;
  if (mode === "clean") env.GSD_PROHIB_SUBJECT = CLEAN_REL;
  const result = spawnSync(process.execPath, ["--test", "--test-reporter=tap", "--", ADAPTER_REL], {
    cwd: ROOT,
    env,
    encoding: "utf8",
    timeout: 90_000,
    maxBuffer: MAX_OUTPUT,
    stdio: ["ignore", "pipe", "pipe"],
    shell: false,
    windowsHide: true,
  });
  const output = `${typeof result.stdout === "string" ? result.stdout : ""}${typeof result.stderr === "string" ? result.stderr : ""}`;
  const counts = parseTapCounts(output);
  const bad = mode === "bad";
  const expectedCounts = { tests: 1, pass: bad ? 0 : 1, fail: bad ? 1 : 0, cancelled: 0, skipped: 0, todo: 0 };
  const failedNames = [...output.matchAll(/^not ok \d+ - (.+)$/gm)].map((match) => match[1]);
  assert.equal(result.error, undefined, `${mode}: Node test target must spawn`);
  assert.equal(result.signal, null, `${mode}: Node test target must not be interrupted`);
  assert.equal(result.status, bad ? 1 : 0, `${mode}: direct TAP exit status`);
  assert.deepEqual(counts, expectedCounts, `${mode}: direct TAP counts`);
  assert.deepEqual(failedNames, bad ? [TEST_NAME] : [], `${mode}: only the named prohibition assertion may fail`);
  assert.ok(output.includes(TEST_NAME), `${mode}: TAP output must name its non-vacuous test`);
  return { output, counts, status: result.status, sha256: digest(Buffer.from(output)) };
}

function runGoControls(evidenceDir, sample = "defect") {
  const controls = {};
  for (const mode of ["bad", "clean", "default"]) {
    const subject = mode === "bad" ? sample : "none";
    const result = runExactRoots(subject, sample);
    const expected = mode === "bad" ? "violation" : "green";
    assert.equal(result.kind, expected, `${mode}: exact ${sample} Go roots must be ${expected}: ${result.reason || ""}`);
    const log = saveRaw(evidenceDir, `${mode}-go.log`, result.output);
    controls[mode] = { roots: result.roots, failureNames: result.failureNames, log };
    if (result.overlay) controls[mode].overlay = result.overlay;
  }
  return controls;
}

function runTapControls(evidenceDir, sample = "defect") {
  const controls = {};
  for (const mode of ["bad", "clean", "default"]) {
    const tap = runDirectTap(mode, sample);
    controls[mode] = { ...tap, log: saveRaw(evidenceDir, `${mode}-node.tap`, tap.output) };
  }
  return controls;
}

function createResolvedCandidate(current) {
  const { projectProhibitions, descriptorFromProjection } = loadGsdProjectionApi();
  const candidate = current.item.status === "resolved"
    ? [current.projected]
    : projectProhibitions([{ ...current.item, status: "resolved", verification: "test" }]);
  assert.equal(candidate.length, 1);
  assert.equal(candidate[0].status, "resolved");
  assert.equal(candidate[0].verification, "test");
  const check = descriptorFromProjection(candidate[0]);
  assert.deepEqual(check, current.check, "resolved candidate must retain the same installed descriptor read-back");
  return { prohibition: candidate[0], check, mode: "auto" };
}

function readCanonicalProducerSuiteSelections(evidenceDir, subject) {
  const expectedRoots = exactRootNamesForMode(subject);
  const controls = {};
  for (const lane of ["bad", "clean", "default"]) {
    const recordPath = path.join(evidenceDir, `canonical-suite-${lane}.json`);
    const bytes = fs.readFileSync(recordPath);
    const receipt = JSON.parse(bytes.toString("utf8"));
    const expectedSubject = lane === "bad" ? subject : "none";
    const expectedSource = lane === "bad" ? FIXTURES[subject].relative
      : lane === "clean" ? CLEAN_REL : "default";
    const expectedResult = lane === "bad" ? "violation" : "green";
    assert.equal(receipt.version, 1, `${lane}: canonical suite receipt version`);
    assert.equal(receipt.control, lane, `${lane}: canonical suite receipt lane`);
    assert.equal(receipt.subject, expectedSubject, `${lane}: canonical producer must use the intended fixture subject`);
    assert.equal(receipt.source, expectedSource, `${lane}: canonical producer must use the intended fixture source`);
    assert.equal(receipt.mode, subject, `${lane}: canonical producer must retain the requested proof mode`);
    assert.deepEqual(receipt.exact_root_inventory, expectedRoots, `${lane}: canonical producer exact root inventory`);
    assert.equal(receipt.semantic_result, expectedResult, `${lane}: canonical producer semantic result`);
    if (subject === "reason") {
      assert.equal(receipt.reason_overlay_binding_controls?.baseline, "accepted", `${lane}: matching reason overlay bindings must pass`);
      assert.deepEqual(receipt.reason_overlay_binding_controls?.controls?.map(({ name, result }) => ({ name, result })), [
        { name: "omit-parser-source-binding", result: "rejected" },
        { name: "change-corevalidate-overlay-original-source-hash", result: "rejected" },
      ], `${lane}: missing binding and changed overlay hash controls must be refused`);
    }
    controls[lane] = {
      ...receipt,
      evidence: { path: recordPath, sha256: digest(bytes), bytes: bytes.length },
    };
  }
  const recordedControls = fs.readdirSync(evidenceDir)
    .filter((name) => /^canonical-suite-(bad|clean|default)\.json$/.test(name))
    .sort();
  assert.deepEqual(recordedControls, ["canonical-suite-bad.json", "canonical-suite-clean.json", "canonical-suite-default.json"],
    "canonical producer must record exactly its bad, clean and default suite selections");
  return { mode: subject, root_inventory: expectedRoots, controls };
}

function runCanonicalProducer(request, evidenceDir, subject) {
  const cli = path.join(os.homedir(), ".codex", "gsd-core", "bin", "gsd-tools.cjs");
  assert.ok(fs.existsSync(cli), `installed canonical GSD producer is missing at ${cli}`);
  const requestText = `${JSON.stringify(request, null, 2)}\n`;
  const requestRecord = saveRaw(evidenceDir, "canonical-request.json", requestText);
  const env = { ...process.env, GOCACHE, GOFLAGS: "", GSD_PROHIB_MODE: subject,
    GSD_PROHIB_PRODUCER_RECEIPT_DIR: evidenceDir };
  delete env.GSD_PROHIB_SUBJECT;
  const result = spawnSync(process.execPath, [cli, "check", "prohibition-enforcement", requestRecord.path, "--raw"], {
    cwd: ROOT,
    env,
    encoding: "utf8",
    timeout: 120_000,
    maxBuffer: MAX_OUTPUT,
    stdio: ["ignore", "pipe", "pipe"],
    shell: false,
    windowsHide: true,
  });
  const stdout = typeof result.stdout === "string" ? result.stdout : "";
  const stderr = typeof result.stderr === "string" ? result.stderr : "";
  const stdoutRecord = saveRaw(evidenceDir, "canonical-producer.stdout.json", stdout);
  const stderrRecord = saveRaw(evidenceDir, "canonical-producer.stderr.log", stderr);
  assert.equal(result.error, undefined, "canonical producer must spawn");
  assert.equal(result.signal, null, "canonical producer must not be interrupted");
  assert.equal(result.status, 0, `canonical producer exit: ${stderr || stdout}`);
  const verdict = JSON.parse(stdout);
  assert.equal(verdict.status, "green");
  assert.equal(verdict.flagged, false);
  assert.equal(verdict.tier, "test");
  assert.equal(verdict.located, true);
  assert.equal(verdict.kind, "node-test");
  assert.equal(verdict.evidence.length, 1);
  assert.equal(verdict.evidence[0].target, ADAPTER_REL);
  assert.equal(verdict.evidence[0].failFirst, true);
  assert.equal(verdict.evidence[0].passed, true);
  assert.equal(verdict.evidence[0].failFirstProof, "violation-fixture");
  const suiteSelection = readCanonicalProducerSuiteSelections(evidenceDir, subject);
  return { verdict, requestRecord, stdoutRecord, stderrRecord, suiteSelection };
}

function sourceBindings(current) {
  const selected = new Set([
    PLAN_REL, FAILURE_PLAN_REL, REASON_PLAN_REL, BASELINE_PLAN_REL, ADAPTER_REL,
    SOURCE_REL, TEST_REL, MODEL_REL, REASON_TEST_REL, CGEN_REL, PARSER_REL, COREVALIDATE_REL,
    CGEN_TEST_REL, COMPATIBILITY_TEST_REL,
    "examples/phase24/error.schway", CLEAN_REL, DEFECT_REL, FAILURE_REL, REASON_REL, BASELINE_REL,
    "testdata/phase371/pre-emitter-census.json", "testdata/phase371/compatibility-amendments.json",
  ]);
  return Object.fromEntries([...selected].map((relative) => {
    const bytes = fs.readFileSync(path.join(ROOT, relative));
    return [relative, { sha256: digest(bytes), bytes: bytes.length }];
  }));
}

function assertReasonOverlaySourceBindings(proof) {
  assert.equal(proof.subject, "reason", "reason overlay bindings belong to the reason proof");
  const sources = proof.go_controls?.bad?.overlay?.sources;
  assert.ok(Array.isArray(sources), "reason proof must retain per-source overlay records");
  assert.deepEqual(sources.map((source) => source.path).sort(), [...REASON_OVERLAY_SOURCE_RELATIVES].sort(),
    "reason overlay must bind exactly the emitter, parser and core validator inputs");
  for (const relative of REASON_OVERLAY_SOURCE_RELATIVES) {
    const source = sources.find((item) => item.path === relative);
    const binding = proof.source_bindings?.[relative];
    assert.ok(binding && typeof binding.sha256 === "string" && Number.isInteger(binding.bytes),
      `reason proof source binding missing: ${relative}`);
    assert.equal(source.originalSourceSha256, binding.sha256,
      `reason overlay original source hash mismatch: ${relative}`);
  }
}

function assertLegacyReasonOverlaySourceHashes(proof, sourceRevision) {
  const sources = proof.go_controls?.bad?.overlay?.sources;
  assert.ok(Array.isArray(sources), "historical reason proof must retain its original per-source overlay records");
  for (const relative of [PARSER_REL, COREVALIDATE_REL]) {
    assert.equal(proof.source_bindings[relative], undefined,
      `historical reason proof must preserve its recorded omission for ${relative}`);
    const source = sources.find((item) => item.path === relative);
    assert.ok(source && typeof source.originalSourceSha256 === "string",
      `historical reason overlay must preserve the recorded original hash for ${relative}`);
    const archived = gitFileAt(sourceRevision, relative);
    assert.equal(digest(archived), source.originalSourceSha256,
      `historical reason overlay hash must match the preserved source revision for ${relative}`);
  }
}

function runReasonOverlayBindingNegativeControls() {
  const currentBindings = sourceBindings(projectCurrentPlan("reason"));
  const currentOverlaySources = REASON_OVERLAY_SOURCE_RELATIVES.map((relative) => ({
    path: relative,
    originalSourceSha256: currentBindings[relative].sha256,
    replacementSourceSha256: "0".repeat(64),
    baselineHead: "negative-control-fixture",
  }));
  const baseline = {
    subject: "reason",
    source_bindings: currentBindings,
    go_controls: { bad: { overlay: { sources: currentOverlaySources } } },
  };
  assert.doesNotThrow(() => assertReasonOverlaySourceBindings(baseline),
    "complete current source bindings must accept the matching reason overlay");

  const missingBinding = JSON.parse(JSON.stringify(baseline));
  delete missingBinding.source_bindings[PARSER_REL];
  assert.throws(() => assertReasonOverlaySourceBindings(missingBinding),
    new RegExp(`reason proof source binding missing: ${escapeRegex(PARSER_REL)}`),
    "omitting parser.go must be refused by the reason overlay archive check");

  const changedOverlayHash = JSON.parse(JSON.stringify(baseline));
  changedOverlayHash.go_controls.bad.overlay.sources.find((source) => source.path === COREVALIDATE_REL).originalSourceSha256 = "f".repeat(64);
  assert.throws(() => assertReasonOverlaySourceBindings(changedOverlayHash),
    new RegExp(`reason overlay original source hash mismatch: ${escapeRegex(COREVALIDATE_REL)}`),
    "changing the corevalidate.go overlay source hash must be refused by the archive check");

  return {
    baseline: "accepted",
    controls: [
      { name: "omit-parser-source-binding", result: "rejected", reason: `reason proof source binding missing: ${PARSER_REL}` },
      { name: "change-corevalidate-overlay-original-source-hash", result: "rejected", reason: `reason overlay original source hash mismatch: ${COREVALIDATE_REL}` },
    ],
  };
}

function sourceBindingChanges(from, to) {
  const changed = {};
  for (const relative of [...new Set([...Object.keys(from), ...Object.keys(to)])].sort()) {
    const before = from[relative] || null;
    const after = to[relative] || null;
    if (JSON.stringify(before) !== JSON.stringify(after)) changed[relative] = { from: before, to: after };
  }
  return changed;
}

function writeJsonAtomic(relative, value) {
  const target = path.join(ROOT, relative);
  const temp = `${target}.tmp-${process.pid}-${crypto.randomBytes(5).toString("hex")}`;
  fs.writeFileSync(temp, `${JSON.stringify(value, null, 2)}\n`, { flag: "wx" });
  fs.renameSync(temp, target);
}

function proofSubjects() {
  return ["defect", "failure", "reason", "baseline"];
}

function archivedFileRecords(value, records = []) {
  if (!value || typeof value !== "object") return records;
  if (typeof value.path === "string" && typeof value.sha256 === "string" && Number.isInteger(value.bytes)) records.push(value);
  for (const child of Object.values(value)) archivedFileRecords(child, records);
  return records;
}

function assertArchivedSourceBindings(proof, sourceRevision) {
  assert.ok(proof.source_bindings && typeof proof.source_bindings === "object", `${proof.subject}: source bindings are required`);
  for (const [relative, expected] of Object.entries(proof.source_bindings)) {
    let actual = fs.existsSync(path.join(ROOT, relative)) ? fs.readFileSync(path.join(ROOT, relative)) : null;
    if (!actual || digest(actual) !== expected.sha256 || actual.length !== expected.bytes) {
      assert.ok(sourceRevision, `${proof.subject}: archived source binding revision is required`);
      actual = gitFileAt(sourceRevision, relative);
    }
    assert.equal(digest(actual), expected.sha256, `${proof.subject}: source binding changed for ${relative}`);
    assert.equal(actual.length, expected.bytes, `${proof.subject}: source binding byte count changed for ${relative}`);
  }
}

function assertArchivedProof(subject, proof, sourceRevision) {
  assert.equal(proof.subject, subject);
  assert.equal(proof.state, "machine-evidence-complete-pending-lifecycle-resolution");
  assert.equal(proof.lifecycle_status, "unresolved");
  assert.equal(proof.candidate_status, "resolved");
  assert.equal(proof.candidate_verification, "test");
  assert.equal(proof.current_plan_projection.source_prohibition.status, "unresolved");
  assert.equal(proof.current_plan_projection.source_prohibition.verification, null);
  assert.equal(proof.current_plan_projection.source_plan.sha256, proof.source_bindings[proof.current_plan_projection.source_plan.path].sha256);
  assertArchivedSourceBindings(proof, sourceRevision);
  if (proof.proof_format_version === CURRENT_PROOF_FORMAT_VERSION) {
    for (const relative of [PARSER_REL, COREVALIDATE_REL]) {
      assert.ok(proof.source_bindings[relative], `${subject}: current proof format must bind ${relative}`);
    }
    if (subject === "reason") assertReasonOverlaySourceBindings(proof);
  } else {
    assert.equal(proof.proof_format_version, undefined, `${subject}: archived proof format must be explicitly current or preserved legacy`);
    if (subject === "reason") assertLegacyReasonOverlaySourceHashes(proof, sourceRevision);
  }

  for (const record of archivedFileRecords(proof)) {
    const bytes = fs.readFileSync(record.path);
    assert.equal(bytes.length, record.bytes, `${subject}: archived evidence byte count changed: ${record.path}`);
    assert.equal(digest(bytes), record.sha256, `${subject}: archived evidence hash changed: ${record.path}`);
  }

  for (const mode of ["bad", "clean", "default"]) {
    const control = proof.go_controls[mode];
    assert.ok(control && control.roots && Object.keys(control.roots).length > 0, `${subject}/${mode}: Go roots are required`);
    assert.equal(control.failureNames.length > 0, mode === "bad", `${subject}/${mode}: Go semantic result must match its lane`);
    for (const root of Object.values(control.roots)) assert.equal(root.skip, 0, `${subject}/${mode}: Go root must not skip`);
    const tap = proof.tap_controls[mode];
    assert.ok(tap && tap.log, `${subject}/${mode}: TAP record is required`);
    const tapOutput = fs.readFileSync(tap.log.path, "utf8");
    assert.deepEqual(parseTapCounts(tapOutput), tap.counts, `${subject}/${mode}: stored TAP counts must match raw output`);
    const failedNames = [...tapOutput.matchAll(/^not ok \d+ - (.+)$/gm)].map((match) => match[1]);
    assert.deepEqual(failedNames, mode === "bad" ? [TEST_NAME] : [], `${subject}/${mode}: TAP must reach only the named assertion`);
    assert.equal(tap.status, mode === "bad" ? 1 : 0, `${subject}/${mode}: TAP exit must match its lane`);
    assert.equal(tap.counts.skipped, 0, `${subject}/${mode}: TAP must not skip`);
    assert.equal(tap.counts.todo, 0, `${subject}/${mode}: TAP must not mark todo`);
  }

  const producer = proof.canonical_producer;
  assert.equal(producer.status, "green", `${subject}: prior canonical producer must be green`);
  assert.equal(producer.flagged, false, `${subject}: prior canonical producer must be unflagged`);
  assert.equal(producer.tier, "test");
  assert.equal(producer.located, true);
  assert.equal(producer.evidence.length, 1);
  assert.equal(producer.evidence[0].failFirst, true);
  assert.equal(producer.evidence[0].passed, true);
  const producerStdout = JSON.parse(fs.readFileSync(producer.stdout.path, "utf8"));
  assert.equal(producerStdout.status, producer.status, `${subject}: producer receipt must match its raw canonical output`);
  assert.equal(producerStdout.flagged, producer.flagged, `${subject}: producer flag must match its raw canonical output`);
  if (producer.suite_selection) {
    const selection = producer.suite_selection;
    const expectedRoots = exactRootNamesForMode(subject);
    assert.equal(selection.mode, subject, `${subject}: canonical producer mode must be recorded`);
    assert.deepEqual(selection.root_inventory, expectedRoots, `${subject}: canonical producer root inventory must be exact`);
    assert.deepEqual(Object.keys(selection.controls).sort(), ["bad", "clean", "default"], `${subject}: canonical producer controls must be complete`);
    for (const [lane, expectedResult] of [["bad", "violation"], ["clean", "green"], ["default", "green"]]) {
      const control = selection.controls[lane];
      assert.equal(control.mode, subject, `${subject}/${lane}: canonical producer mode`);
      assert.deepEqual(control.exact_root_inventory, expectedRoots, `${subject}/${lane}: exact canonical roots`);
      assert.equal(control.semantic_result, expectedResult, `${subject}/${lane}: expected canonical result`);
    }
  }
}

function ensurePreResolutionArchive(initial) {
  const resolutionPath = path.join(ROOT, RESOLUTION_REL);
  if (fs.existsSync(resolutionPath)) {
    const saved = JSON.parse(fs.readFileSync(resolutionPath, "utf8"));
    assert.equal(saved.version, 1);
    assert.ok(saved.pre_resolution_snapshot, "pre-resolution receipts must remain archived before any live proof overwrite");
    assert.ok(saved.pre_resolution_snapshot.source_binding_revision, "archived source binding revision must remain inspectable after this commit");
    assert.equal(saved.pre_resolution_snapshot.proofs_sha256, digest(Buffer.from(JSON.stringify(saved.pre_resolution_snapshot.proofs))));
    for (const subject of proofSubjects()) assertArchivedProof(subject, saved.pre_resolution_snapshot.proofs[subject], saved.pre_resolution_snapshot.source_binding_revision);
    return saved;
  }

  const oldProofs = initial.proofs;
  assert.deepEqual(Object.keys(oldProofs || {}).sort(), proofSubjects().sort(), "all four existing proof records must be archived together");
  const sourceBindingRevision = gitHead();
  for (const subject of proofSubjects()) assertArchivedProof(subject, oldProofs[subject], sourceBindingRevision);
  const snapshot = {
    captured_utc: new Date().toISOString(),
    source_binding_revision: sourceBindingRevision,
    source_head: initial.source_head,
    source_plan: initial.source_plan,
    source_prohibition: initial.source_prohibition,
    initial_projection: initial.projection,
    proofs: JSON.parse(JSON.stringify(oldProofs)),
    proofs_sha256: digest(Buffer.from(JSON.stringify(oldProofs))),
    final_root_union: initial.final_root_union,
  };
  const record = {
    version: 1,
    state: "pre-resolution-archived",
    captured_utc: new Date().toISOString(),
    machine_authority: "Normal GSD test-tier lifecycle promotion delegated for this phase after the recorded semantic RED, clean/default controls, and canonical producers passed.",
    transition: { from_status: "unresolved", from_verification: null, to_status: "resolved", to_verification: "test" },
    pre_resolution_snapshot: snapshot,
  };
  fs.writeFileSync(resolutionPath, `${JSON.stringify(record, null, 2)}\n`, { flag: "wx" });
  return record;
}

function beginSourceEvolution(initial, current) {
  if (current.item.status === "unresolved") {
    assert.equal(current.item.verification, null, "unresolved lifecycle source evolution must retain null verification");
    return null;
  }
  assert.equal(current.item.status, "resolved", "source evolution only supports the two declared plan lifecycle states");
  const resolutionPath = path.join(ROOT, RESOLUTION_REL);
  const record = JSON.parse(fs.readFileSync(resolutionPath, "utf8"));
  assert.equal(record.state, "resolved-test", "source evolution requires the existing resolved/test lifecycle");
  const toBindings = sourceBindings(current);
  const existing = record.source_evolution_history || [];
  const sameTarget = existing.find((entry) => JSON.stringify(entry.to_source_bindings) === JSON.stringify(toBindings));
  if (sameTarget) {
    assert.ok(sameTarget.status === "pending-fresh-proofs" || sameTarget.status === "verified", "source evolution record has an unsupported state");
    return sameTarget;
  }

  const resolvedSubjects = record.transition.subjects;
  assert.ok(resolvedSubjects && resolvedSubjects.baseline, "resolved lifecycle source bindings are required");
  const fromBindings = resolvedSubjects.baseline.current_source_bindings;
  assert.ok(fromBindings && typeof fromBindings === "object", "resolved baseline source bindings are required");
  const changed = sourceBindingChanges(fromBindings, toBindings);
  if (Object.keys(changed).length === 0) return null;
  const changedPaths = Object.keys(changed).sort();
  const proofBindingCorrection = changedPaths.some((relative) => relative === PARSER_REL || relative === COREVALIDATE_REL)
    && changedPaths.every((relative) => PROOF_BINDING_EVOLUTION_PATHS.includes(relative));
  const storageEvolution = changedPaths.every((relative) => STORAGE_EVOLUTION_PATHS.includes(relative));
  assert.ok(proofBindingCorrection || storageEvolution,
    "source evolution is limited to the existing storage repair or this phase-local proof-binding correction");

  const priorProofs = initial.proofs;
  assert.deepEqual(Object.keys(priorProofs || {}).sort(), proofSubjects().sort(), "all four current resolved proofs must be snapshotted before replacement");
  for (const subject of proofSubjects()) {
    const proof = priorProofs[subject];
    assert.equal(proof.lifecycle_status, "resolved", `${subject}: previous proof must record the resolved lifecycle`);
    assert.deepEqual(proof.source_bindings, resolvedSubjects[subject].current_source_bindings,
      `${subject}: prior proof bindings must match the previous resolved record`);
  }
  const priorReceiptsIncomplete = proofSubjects().some((subject) => {
    const proof = priorProofs[subject];
    const selection = proof.canonical_producer?.suite_selection;
    return proof.proof_format_version !== CURRENT_PROOF_FORMAT_VERSION
      || !selection || selection.mode !== subject
      || !selection.controls || Object.keys(selection.controls).sort().join(",") !== "bad,clean,default";
  });
  {
    const historicalSourceRevision = record.pre_resolution_snapshot?.source_binding_revision;
    if (priorProofs.reason.proof_format_version === CURRENT_PROOF_FORMAT_VERSION) {
      assertReasonOverlaySourceBindings(priorProofs.reason);
    } else {
      assert.ok(historicalSourceRevision, "historical reason proof source revision must remain available for the correction amendment");
      assertLegacyReasonOverlaySourceHashes(priorProofs.reason, historicalSourceRevision);
    }
  }
  assert.ok(initial.final_root_union && initial.final_root_union.raw_log, "previous resolved root union must be preserved before proof replacement");

  const event = {
    id: proofBindingCorrection
      ? `T-37.1-${priorReceiptsIncomplete ? "CR01-CR02-proof-binding-correction" : "CR02-proof-binding-reverification"}-${existing.length + 1}`
      : `T-37.1-08-counted-storage-repair-${existing.length + 1}`,
    captured_utc: new Date().toISOString(),
    status: "pending-fresh-proofs",
    reason: proofBindingCorrection
      ? priorReceiptsIncomplete
        ? "Dated CR-01/CR-02 correction: prior green receipts remain preserved as historical but incomplete because the canonical clean/default mode was not recorded and reason parser/corevalidate bindings were omitted; this event supersedes them only after fresh current-format proofs."
        : "Dated CR-02 follow-up: prior current-format proofs already carry complete canonical modes and reason source bindings; their immutable payloads are preserved while this adapter guard edit receives fresh source-bound proofs."
      : "Bounded source evolution for block-scoped static readonly counted defect-reason bytes; prior resolved proof facts are preserved below.",
    ...(proofBindingCorrection ? {
      amendment_date: "2026-10-16",
      review_findings: priorReceiptsIncomplete ? ["CR-01", "CR-02"] : ["CR-02"],
      prior_receipt_qualification: priorReceiptsIncomplete
        ? "Historical only; the original proof payloads and hashes are preserved unchanged, but their clean/default canonical producer mode and parser.go/corevalidate.go reason bindings were incomplete."
        : "Prior current-format proof payloads and hashes are preserved unchanged; their canonical producer mode and parser.go/corevalidate.go reason bindings were complete, and this event records a fresh proof set for the subsequent adapter guard edit.",
    } : {}),
    lifecycle: { status: "resolved", verification: "test" },
    from_source_bindings: fromBindings,
    to_source_bindings: toBindings,
    changed_bindings: changed,
    previous_resolved_proofs: JSON.parse(JSON.stringify(priorProofs)),
    previous_resolved_proofs_sha256: digest(Buffer.from(JSON.stringify(priorProofs))),
    previous_final_root_union: JSON.parse(JSON.stringify(initial.final_root_union)),
    previous_final_root_union_sha256: digest(Buffer.from(JSON.stringify(initial.final_root_union))),
  };
  record.source_evolution_history = [...existing, event];
  writeJsonAtomic(RESOLUTION_REL, record);
  return event;
}

function replacePlanField(planText, field, value) {
  const pattern = new RegExp(`^(      ${field}: ).*$`, "m");
  assert.match(planText, pattern, `negative-control copy must contain ${field}`);
  return planText.replace(pattern, `$1${value}`);
}

function runTransitionNegativeControls() {
  const cleanBytes = fs.readFileSync(path.join(ROOT, PLAN_REL));
  const cleanText = cleanBytes.toString("utf8");
  const temporary = tempEvidenceDir();
  const resolution = JSON.parse(fs.readFileSync(path.join(ROOT, RESOLUTION_REL), "utf8"));
  const priorSourceRevision = resolution.pre_resolution_snapshot.source_binding_revision;
  assert.ok(priorSourceRevision, "prior-state round-trip requires the archived source revision");
  const mutations = [
    ["invalid-status", (text) => replacePlanField(text, "status", "archived")],
    ["invalid-verification-tier", (text) => replacePlanField(text, "verification", "judgment")],
    ["invalid-partial-transition", (text) => replacePlanField(replacePlanField(text, "status", "unresolved"), "verification", "test")],
    ["altered-descriptor", (text) => replacePlanField(text, "check_target", "internal/other.test.cjs")],
  ];
  const outcomes = [];
  for (const [name, mutate] of mutations) {
    const mutatedText = mutate(cleanText);
    const copied = saveRaw(temporary, `${name}.plan.md`, mutatedText);
    let rejected = false;
    let reason = "";
    try {
      projectCurrentPlan("defect", fs.readFileSync(copied.path, "utf8"));
    } catch (error) {
      rejected = true;
      reason = error.message;
    }
    assert.equal(rejected, true, `${name}: malformed lifecycle or descriptor mutation must fail closed`);
    outcomes.push({ name, result: "rejected", reason, mutation_copy: copied });
  }
  assert.deepEqual(fs.readFileSync(path.join(ROOT, PLAN_REL)), cleanBytes, "temporary negative-control mutations must leave the live plan bytes untouched");
  const restored = projectCurrentPlan("defect");
  const priorStateRoundTrip = {};
  for (const subject of proofSubjects()) {
    const planRelative = subject === "defect" ? PLAN_REL : ({ failure: FAILURE_PLAN_REL, reason: REASON_PLAN_REL, baseline: BASELINE_PLAN_REL })[subject];
    const historicalBytes = gitFileAt(priorSourceRevision, planRelative);
    const historical = projectCurrentPlan(subject, historicalBytes.toString("utf8"));
    assert.equal(historical.item.status, "unresolved", `${subject}: prior-state plan snapshot must remain unresolved`);
    assert.equal(historical.item.verification, null, `${subject}: prior-state plan snapshot must remain null-tier`);
    const candidate = createResolvedCandidate(historical);
    assert.equal(candidate.prohibition.status, "resolved", `${subject}: prior-state canonical candidate must resolve`);
    assert.equal(candidate.prohibition.verification, "test", `${subject}: prior-state candidate must use the test tier`);
    assert.deepEqual(candidate.check, historical.check, `${subject}: prior-state round-trip must preserve its descriptor`);
    priorStateRoundTrip[subject] = {
      plan: historical.planRelative,
      source_revision: priorSourceRevision,
      source_plan_sha256: digest(historicalBytes),
      source_status: historical.item.status,
      source_verification: historical.item.verification,
      canonical_candidate: candidate.prohibition,
      descriptor_unchanged: true,
    };
  }
  return {
    evidence_directory: temporary,
    clean_plan_sha256: digest(cleanBytes),
    clean_plan_status: restored.item.status,
    clean_plan_verification: restored.item.verification,
    outcomes,
    restored_clean: true,
    prior_state_roundtrip: priorStateRoundTrip,
  };
}

function currentLifecycle() {
  const current = Object.fromEntries(proofSubjects().map((subject) => [subject, projectCurrentPlan(subject)]));
  const states = new Set(Object.values(current).map(({ item }) => `${item.status}/${item.verification}`));
  assert.ok(states.size === 1 && (states.has("unresolved/null") || states.has("resolved/test")), "all four live plans must share one allowed lifecycle state");
  return { status: Object.values(current)[0].item.status, verification: Object.values(current)[0].item.verification, current };
}

function finalizeResolutionRecord(stored, lifecycle, negativeControls) {
  assert.equal(lifecycle.status, "resolved", "only canonical resolved/test live plans can finalize lifecycle resolution");
  assert.equal(lifecycle.verification, "test");
  const resolutionPath = path.join(ROOT, RESOLUTION_REL);
  const record = JSON.parse(fs.readFileSync(resolutionPath, "utf8"));
  assert.ok(record.state === "pre-resolution-archived" || record.state === "resolved-test", "resolution record must retain its immutable pre-resolution archive");
  const items = {};
  for (const subject of proofSubjects()) {
    const live = lifecycle.current[subject];
    const proof = stored.proofs[subject];
    assert.equal(live.item.status, "resolved");
    assert.equal(live.item.verification, "test");
    assert.deepEqual(proof.current_plan_projection.projection, [live.projected], `${subject}: fresh proof must bind canonical live projection`);
    assert.equal(proof.lifecycle_status, "resolved");
    assert.equal(proof.canonical_producer.status, "green");
    assert.equal(proof.canonical_producer.flagged, false);
    items[subject] = {
      plan: live.planRelative,
      plan_sha256: digest(live.planBytes),
      source_prohibition: live.item,
      canonical_projection: live.projected,
      descriptor: live.check,
      proof_state: proof.state,
      proof_manifest: proof.manifest,
      current_source_bindings: proof.source_bindings,
      canonical_producer: proof.canonical_producer,
    };
  }
  const sourceEvolution = (record.source_evolution_history || []).find((entry) =>
    entry.status === "pending-fresh-proofs" && JSON.stringify(entry.to_source_bindings) === JSON.stringify(sourceBindings(projectCurrentPlan("baseline"))));
  if (sourceEvolution) {
    assert.equal(sourceEvolution.previous_resolved_proofs_sha256,
      digest(Buffer.from(JSON.stringify(sourceEvolution.previous_resolved_proofs))), "prior resolved proofs must remain byte-for-byte archived");
    assert.equal(sourceEvolution.previous_final_root_union_sha256,
      digest(Buffer.from(JSON.stringify(sourceEvolution.previous_final_root_union))), "prior resolved root union must remain byte-for-byte archived");
    sourceEvolution.status = "verified";
    sourceEvolution.verified_utc = new Date().toISOString();
    sourceEvolution.fresh_proofs = Object.fromEntries(proofSubjects().map((subject) => {
      const proof = stored.proofs[subject];
      assert.deepEqual(proof.source_bindings, sourceEvolution.to_source_bindings,
        `${subject}: fresh proof must bind every current source-evolution target`);
      return [subject, {
        proof_sha256: digest(Buffer.from(JSON.stringify(proof))),
        proof_manifest: proof.manifest,
        source_bindings: proof.source_bindings,
        canonical_producer: proof.canonical_producer,
      }];
    }));
    sourceEvolution.final_root_union = stored.final_root_union;
  }
  record.state = "resolved-test";
  record.resolved_utc = new Date().toISOString();
  record.transition = {
    ...record.transition,
    descriptor_unchanged: true,
    subjects: items,
    authority_basis: "Fresh canonical producer runs passed for all four live resolved/test descriptors; every bad lane reached its named semantic assertion; clean/default controls passed with no skips; the bounded 19-root union passed.",
  };
  record.transition_negative_controls = negativeControls;
  record.final_root_union = stored.final_root_union;
  writeJsonAtomic(RESOLUTION_REL, record);
  const resolutionBytes = fs.readFileSync(resolutionPath);
  stored.live_prohibitions = Object.fromEntries(Object.entries(lifecycle.current).map(([subject, live]) => [subject, {
    plan: live.planRelative,
    plan_sha256: digest(live.planBytes),
    source_prohibition: live.item,
    canonical_projection: live.projected,
    descriptor: live.check,
  }]));
  stored.lifecycle_status = "resolved";
  stored.resolution_record = { path: RESOLUTION_REL, sha256: digest(resolutionBytes), bytes: resolutionBytes.length };
  if (sourceEvolution) {
    stored.source_evolution = {
      id: sourceEvolution.id,
      status: sourceEvolution.status,
      changed_bindings: sourceEvolution.changed_bindings,
      fresh_proofs: sourceEvolution.fresh_proofs,
    };
  }
  writeJsonAtomic(PROJECTION_REL, stored);
  return record;
}

function runProofMode(subject) {
  if (subject !== "defect" && subject !== "failure" && subject !== "reason" && subject !== "baseline") throw new Error("only closed defect/failure/reason/baseline proofs are supported");
  const initial = ensureInitialProjection();
  ensurePreResolutionArchive(initial);
  const current = projectCurrentPlan(subject);
  beginSourceEvolution(initial, current);
  const lifecycle = current.item.status;
  assert.ok(lifecycle === "unresolved" || lifecycle === "resolved", "current lifecycle must be one of the supported states");
  const lifecycleStatus = lifecycle;
  const evidenceState = lifecycle === "resolved"
    ? "machine-evidence-and-lifecycle-resolved"
    : "machine-evidence-complete-pending-lifecycle-resolution";

  const evidenceDir = tempEvidenceDir();
  const goControls = runGoControls(evidenceDir, subject);
  const tapControls = runTapControls(evidenceDir, subject);
  const canonicalItem = createResolvedCandidate(current);
  const producer = runCanonicalProducer({ prohibition: canonicalItem.prohibition, check: current.check, mode: "auto" }, evidenceDir, subject);
  const bindings = sourceBindings(current);
  const proof = {
    proof_format_version: CURRENT_PROOF_FORMAT_VERSION,
    subject,
    state: evidenceState,
    captured_utc: new Date().toISOString(),
    candidate_status: "resolved",
    candidate_verification: "test",
    lifecycle_status: lifecycleStatus,
    evidence_directory: evidenceDir,
    source_bindings: bindings,
    canonical_projection: canonicalItem.prohibition,
    current_plan_projection: {
      source_plan: { path: current.planRelative, sha256: digest(current.planBytes) },
      source_prohibition: current.item,
      projection: [current.projected],
    },
    go_controls: goControls,
    tap_controls: Object.fromEntries(Object.entries(tapControls).map(([mode, control]) => [mode, {
      counts: control.counts,
      status: control.status,
      sha256: control.sha256,
      log: control.log,
    }])),
    canonical_producer: {
      status: producer.verdict.status,
      flagged: producer.verdict.flagged,
      tier: producer.verdict.tier,
      located: producer.verdict.located,
      kind: producer.verdict.kind,
      evidence: producer.verdict.evidence,
      request: producer.requestRecord,
      stdout: producer.stdoutRecord,
      stderr: producer.stderrRecord,
      suite_selection: producer.suiteSelection,
    },
  };
  for (const relative of [PARSER_REL, COREVALIDATE_REL]) {
    assert.ok(proof.source_bindings[relative], `${subject}: current proof must bind ${relative}`);
  }
  if (subject === "reason") assertReasonOverlaySourceBindings(proof);
  const manifest = {
    version: 1,
    subject,
    initial_projection_sha256: digest(Buffer.from(JSON.stringify(initial.projection))),
    current_plan_projection: proof.current_plan_projection,
    lifecycle_status: lifecycleStatus,
    proof,
  };
  const manifestRecord = saveRaw(evidenceDir, "evidence-manifest.json", `${JSON.stringify(manifest, null, 2)}\n`);
  proof.manifest = manifestRecord;
  const priorProofs = initial.proofs || { defect: initial.proof };
  const projectionRecord = {
    ...initial,
    proofs: { ...priorProofs, [subject]: proof },
    lifecycle_status: lifecycleStatus,
    live_prohibitions: {
      ...(initial.live_prohibitions || {}),
      [subject]: {
        plan: current.planRelative,
        plan_sha256: digest(current.planBytes),
        source_prohibition: current.item,
        canonical_projection: current.projected,
        descriptor: current.check,
      },
    },
    ...(subject === "defect" ? { proof }
      : subject === "failure" ? { failure_plan_projection: proof.current_plan_projection, failure_proof: proof }
        : subject === "reason" ? { reason_plan_projection: proof.current_plan_projection, reason_proof: proof }
          : { baseline_plan_projection: proof.current_plan_projection, baseline_proof: proof }),
  };
  writeJsonAtomic(PROJECTION_REL, projectionRecord);
  process.stdout.write(`canonical ${subject} proof passed for live ${lifecycleStatus} lifecycle; evidence ${evidenceDir}\n`);
}

function runAllProofMode() {
  const initial = ensureInitialProjection();
  ensurePreResolutionArchive(initial);
  const negativeControls = runTransitionNegativeControls();
  for (const subject of ["defect", "failure", "reason", "baseline"]) runProofMode(subject);
  const inventory = runRootInventory("all");
  if (inventory.kind !== "green" || Object.keys(inventory.roots).length !== 19) {
    throw new Error(`final 19-root union failed: ${inventory.reason || inventory.kind}`);
  }
  for (const [name, root] of Object.entries(inventory.roots)) {
    if (root.run !== 1 || root.pass !== 1 || root.fail !== 0 || root.skip !== 0 || root.subtestFail !== 0 || root.subtestSkip !== 0) {
      throw new Error(`final root union contains a failed, repeated, or skipped root/subtest: ${name}`);
    }
  }
  const projectionPath = path.join(ROOT, PROJECTION_REL);
  const stored = JSON.parse(fs.readFileSync(projectionPath, "utf8"));
  const lifecycle = currentLifecycle();
  for (const subject of ["defect", "failure", "reason", "baseline"]) {
    const proof = stored.proofs && stored.proofs[subject];
    assert.ok(proof, `${subject} source-bound proof must remain in the projection history`);
    assert.equal(proof.proof_format_version, CURRENT_PROOF_FORMAT_VERSION,
      `${subject} must use the current complete source-binding proof format`);
    for (const relative of [PARSER_REL, COREVALIDATE_REL]) {
      assert.ok(proof.source_bindings[relative], `${subject} current source bindings must include ${relative}`);
    }
    if (subject === "reason") assertReasonOverlaySourceBindings(proof);
    assert.equal(proof.canonical_producer.status, "green", `${subject} canonical producer proof must be green`);
    assert.equal(proof.canonical_producer.flagged, false, `${subject} canonical producer proof must be unflagged`);
    assert.equal(proof.lifecycle_status, lifecycle.status, `${subject} proof must state the actual current lifecycle`);
    assert.equal(proof.current_plan_projection.source_prohibition.status, lifecycle.status, `${subject} proof must bind live plan status`);
    assert.equal(proof.current_plan_projection.source_prohibition.verification, lifecycle.verification, `${subject} proof must bind live verification tier`);
  }
  const evidenceDir = tempEvidenceDir();
  const log = saveRaw(evidenceDir, "final-19-root-union.go.log", inventory.output);
  const union = {
    state: lifecycle.status === "resolved" ? "machine-evidence-and-lifecycle-resolved" : "machine-evidence-complete-pending-lifecycle-resolution",
    captured_utc: new Date().toISOString(),
    root_count: Object.keys(inventory.roots).length,
    distinct_package_count: 3,
    roots: inventory.roots,
    raw_log: log,
    source_bindings: sourceBindings(projectCurrentPlan("baseline")),
    initial_projection_status: stored.projection[0].status,
    current_plan_status: lifecycle.status,
    current_verification: lifecycle.verification,
  };
  assert.equal(union.initial_projection_status, "unresolved", "initial unresolved history must be retained");
  stored.final_root_union = union;
  if (lifecycle.status === "resolved") finalizeResolutionRecord(stored, lifecycle, negativeControls);
  else {
    stored.lifecycle_status = "unresolved";
    writeJsonAtomic(PROJECTION_REL, stored);
  }
  process.stdout.write(`all prohibition proofs and the final 19-root union passed; live lifecycle is ${lifecycle.status}/${lifecycle.verification}; evidence ${evidenceDir}\n`);
}

function directMode() {
  const sampleIndex = process.argv.indexOf("--sample");
  if (sampleIndex >= 0) {
    const subject = process.argv[sampleIndex + 1];
    if (subject !== "defect" && subject !== "failure" && subject !== "reason" && subject !== "baseline" && subject !== "native" && subject !== "all") throw new Error(`unsupported sample: ${subject}`);
    runSampleMode(subject);
    return true;
  }
  const proveIndex = process.argv.indexOf("--prove");
  if (proveIndex >= 0) {
    const subject = process.argv[proveIndex + 1];
    if (subject === "all") {
      runAllProofMode();
      return true;
    }
    if (subject !== "defect" && subject !== "failure" && subject !== "reason" && subject !== "baseline") throw new Error(`unsupported proof: ${subject}`);
    runProofMode(subject);
    return true;
  }
  return false;
}

try {
  if (!directMode()) {
    const selection = selectSubject();
    let result = null;
    let invalidReason = selection.error || "";
    if (!invalidReason) {
      result = runExactRoots(selection.subject, selection.mode);
      if (result.kind === "invalid") invalidReason = result.reason;
    }
    test(TEST_NAME, { skip: invalidReason || false }, () => {
      const expectedRoots = exactRootNamesForMode(selection.mode);
      assert.deepEqual(Object.keys(result.roots || {}).sort(), expectedRoots,
        `${selection.source}: exact ${selection.mode} root inventory must be selected`);
      recordCanonicalProducerSuiteSelection(selection, result);
      const rootCount = expectedRoots.length;
      assert.equal(result.kind, "green", `${selection.source}: exact ${rootCount}-root Go suite produced ${result.kind}`);
    });
  }
} catch (error) {
  process.stderr.write(`Phase37.1 prohibition adapter failed: ${error.message}\n`);
  process.exitCode = 1;
}
