"use strict";

const assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");
const { test } = require("node:test");

const ROOT = path.resolve(__dirname, "../../..");
const PHASE_REL = ".planning/phases/36-scalar-helper-calls/";
const TMP_ROOT = "/tmp";
const MAX_OUTPUT = 8 * 1024 * 1024;
const ROOTS = {
  outcome: {
    package: "./internal/compiler/native",
    name: "TestPhase36CaptureOutcomeSeparation",
    source: "internal/compiler/native/native_app.go",
    needle: 'return outcome, report, &ToolError{Code: "native.evidence_capacity_exhausted", Err: errors.New("application evidence exceeds the capture limit")}',
    replacement: "return RunOutcome{}, report, nil",
    expectedAssertions: [
      "capture exhaustion changed successful work",
      "capture exhaustion hid the later defect",
    ],
    expectedFailingChildren: [
      "successful-work-after-capture-exhaustion",
      "late-defect-after-capture-exhaustion",
    ],
  },
  causal: {
    package: "./internal/compiler/executionpeer",
    name: "TestPhase36CallCausalAttributionControls",
    source: "internal/compiler/executionpeer/executionpeer.go",
    anchor: "func Validate(program core.Program, document execution.Execution) error {",
    insertion: '\n\tif document.Outcome.Value == "999" {\n\t\treturn refusal("numeric", -1, "", "", "", "pretend numeric validation")\n\t}',
    expectedAssertions: [
      "causal peer should accept the structurally plausible false-success trace; numeric replay is not its claim",
    ],
    expectedFailingChildren: [],
  },
};

const FIXTURES = {
  none: { file: "prohibition-clean.json", data: { version: 1, subject: "none" } },
  outcome: { file: "prohibition-outcome-violation.json", data: { version: 1, subject: "outcome" } },
  causal: { file: "prohibition-causal-violation.json", data: { version: 1, subject: "causal" } },
};

function exactCount(text, needle) {
  return text.split(needle).length - 1;
}

function selectSubject() {
  const raw = process.env.GSD_PROHIB_SUBJECT;
  if (raw === undefined) return { subject: "none", source: "default" };

  for (const [subject, fixture] of Object.entries(FIXTURES)) {
    const relative = PHASE_REL + fixture.file;
    const absolute = path.resolve(ROOT, relative);
    if (raw !== relative && raw !== absolute) continue;

    try {
      const bytes = fs.readFileSync(absolute);
      if (bytes.length > 4096) return { error: `fixture exceeds 4 KiB: ${fixture.file}` };
      const parsed = JSON.parse(bytes.toString("utf8"));
      if (JSON.stringify(parsed) !== JSON.stringify(fixture.data)) {
        return { error: `fixture does not match its closed subject schema: ${fixture.file}` };
      }
      return { subject, source: fixture.file };
    } catch (error) {
      return { error: `fixture unavailable or malformed (${fixture.file}): ${error.message}` };
    }
  }
  return { error: "GSD_PROHIB_SUBJECT is outside the closed fixture allowlist" };
}

function overlaySource(kind, spec, sourcePath, tempDir) {
  const original = fs.readFileSync(sourcePath, "utf8");
  let changed;
  if (kind === "outcome") {
    if (exactCount(original, spec.needle) !== 1) throw new Error("native capacity branch anchor count is not exactly one");
    changed = original.replace(spec.needle, spec.replacement);
  } else {
    if (exactCount(original, spec.anchor) !== 1) throw new Error("causal Validate anchor count is not exactly one");
    changed = original.replace(spec.anchor, spec.anchor + spec.insertion);
  }

  const replacementPath = path.join(tempDir, "mutated.go");
  const overlayPath = path.join(tempDir, "overlay.json");
  fs.writeFileSync(replacementPath, changed, { flag: "wx" });
  fs.writeFileSync(overlayPath, JSON.stringify({ Replace: { [sourcePath]: replacementPath } }), { flag: "wx" });
  return overlayPath;
}

function rootLineCount(output, expression) {
  return output.split(/\r?\n/).filter((line) => expression.test(line)).length;
}

function validateGoResult(spec, result, mutated) {
  const stdout = typeof result.stdout === "string" ? result.stdout : "";
  const stderr = typeof result.stderr === "string" ? result.stderr : "";
  const output = `${stdout}\n${stderr}`;
  if (result.error) return { kind: "invalid", reason: `go spawn failed: ${result.error.code || result.error.message}` };
  if (result.signal) return { kind: "invalid", reason: `go was interrupted by ${result.signal}` };
  if (Buffer.byteLength(output) > MAX_OUTPUT) return { kind: "invalid", reason: "Go output exceeded 8 MiB" };

  const name = spec.name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
  const run = new RegExp(`^=== RUN   ${name}$`);
  const pass = new RegExp(`^--- PASS: ${name} \\([0-9]+(?:\\.[0-9]+)?s\\)$`);
  const fail = new RegExp(`^--- FAIL: ${name} \\([0-9]+(?:\\.[0-9]+)?s\\)$`);
  const skip = new RegExp(`^--- SKIP: ${name}(?:\\s|$)`);
  if (rootLineCount(output, run) !== 1 || rootLineCount(output, skip) !== 0 || /^\s*--- SKIP:/m.test(output)) {
    return { kind: "invalid", reason: "exact Go root was missing, duplicated, or skipped", output };
  }

  if (!mutated) {
    if (result.status !== 0 || rootLineCount(output, pass) !== 1 || /^\s*--- FAIL:|^FAIL(?:\s|$)/m.test(output)) {
      return { kind: "invalid", reason: "unmutated exact Go root did not produce one clean root PASS", output };
    }
    return { kind: "green", output };
  }

  const failureLines = output.split(/\r?\n/)
    .filter((line) => /^\s*--- FAIL:/.test(line))
    .map((line) => line.trimStart());
  const expectedRootFailure = rootLineCount(output, fail) === 1;
  const expectedChildFailures = spec.expectedFailingChildren.map((child) => {
    const childName = `${spec.name}/${child}`.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    return new RegExp(`^--- FAIL: ${childName} \\([0-9]+(?:\\.[0-9]+)?s\\)$`);
  });
  const childLines = failureLines.filter((line) => !fail.test(line));
  const childrenMatch = childLines.length === expectedChildFailures.length
    && childLines.every((line) => expectedChildFailures.some((pattern) => pattern.test(line)));
  const assertionsMatch = spec.expectedAssertions.every((assertion) => exactCount(output, assertion) === 1);
  if (result.status !== 1 || !expectedRootFailure || rootLineCount(output, pass) !== 0
      || failureLines.length !== 1 + expectedChildFailures.length || !childrenMatch || !assertionsMatch) {
    return { kind: "invalid", reason: "mutated Go root did not fail only at its named existing assertion(s)", output };
  }
  return { kind: "violation", output };
}

function runExactRoot(kind, selectedSubject) {
  const spec = ROOTS[kind];
  const sourcePath = path.resolve(ROOT, spec.source);
  let originalBytes;
  let tempDir = null;
  let captured;
  let cleanupError = "";
  try {
    originalBytes = fs.readFileSync(sourcePath);
    const mutated = selectedSubject === kind;
    const args = ["test"];
    if (mutated) {
      tempDir = fs.mkdtempSync(path.join(TMP_ROOT, "schway36-prohibition-"));
      const overlayPath = overlaySource(kind, spec, sourcePath, tempDir);
      args.push("-overlay", overlayPath);
    }
    args.push(spec.package, "-run", `^${spec.name}$`, "-count=1", "-v", "-timeout=20s");
    const child = spawnSync("go", args, {
      cwd: ROOT,
      env: { ...process.env, GOCACHE: "/tmp/ai-lang-verification-gocache", GOFLAGS: "" },
      encoding: "utf8",
      timeout: 20_000,
      maxBuffer: MAX_OUTPUT,
      stdio: ["ignore", "pipe", "pipe"],
      shell: false,
      windowsHide: true,
    });
    captured = validateGoResult(spec, child, mutated);
    const afterBytes = fs.readFileSync(sourcePath);
    if (!afterBytes.equals(originalBytes)) captured = { kind: "invalid", reason: `tracked source changed during ${kind} owner run` };
  } catch (error) {
    captured = { kind: "invalid", reason: `overlay preparation or owner capture failed (${kind}): ${error.message}` };
  } finally {
    if (tempDir) {
      try {
        fs.rmSync(tempDir, { recursive: true, force: true });
      } catch (error) {
        cleanupError = error.message;
      }
    }
  }
  if (cleanupError) return { kind: "invalid", reason: `temporary overlay cleanup failed (${kind}): ${cleanupError}` };
  return captured || { kind: "invalid", reason: `owner capture produced no result (${kind})` };
}

const selection = selectSubject();
let invalidReason = selection.error || "";
let results = null;
if (!invalidReason) {
  results = {
    outcome: runExactRoot("outcome", selection.subject),
    causal: runExactRoot("causal", selection.subject),
  };
  for (const [kind, result] of Object.entries(results)) {
    if (result.kind === "invalid") invalidReason = `${kind} owner result is not a valid proof: ${result.reason}`;
  }
}

function proofMessage(kind) {
  const result = results && results[kind];
  return result && result.kind === "violation"
    ? `known-bad ${selection.subject} overlay reached its named existing Go assertion`
    : `${kind} root classification was ${result ? result.kind : "unavailable"}`;
}

test("native capture exhaustion preserves the actual application outcome", { skip: invalidReason || false }, () => {
  assert.equal(results.outcome.kind, "green", proofMessage("outcome"));
});

test("causal peer does not claim numeric replay of a false-success trace", { skip: invalidReason || false }, () => {
  assert.equal(results.causal.kind, "green", proofMessage("causal"));
});
