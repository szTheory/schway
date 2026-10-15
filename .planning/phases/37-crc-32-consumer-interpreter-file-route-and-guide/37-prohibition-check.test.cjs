"use strict";

const assert = require("node:assert/strict");
const { spawnSync } = require("node:child_process");
const crypto = require("node:crypto");
const fs = require("node:fs");
const os = require("node:os");
const path = require("node:path");
const { test } = require("node:test");

const ROOT = path.resolve(__dirname, "../../..");
const PHASE_REL = ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/";
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
      tempDir = fs.mkdtempSync(path.join(TMP_ROOT, "schway37-prohibition-"));
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

const STATEMENTS = {
  outcome: "MUST NOT turn optional capture exhaustion into application success or hide a later checked defect.",
  causal: "MUST NOT present causal structural acceptance as independent CRC numeric verification.",
};
const TEST_NAMES = {
  outcome: "native capture exhaustion preserves the actual application outcome",
  causal: "causal peer does not claim numeric replay of a false-success trace",
};

function digest(bytes) {
  return crypto.createHash("sha256").update(bytes).digest("hex");
}

function projectRequest(kind) {
  const installed = path.join(os.homedir(), ".codex", "gsd-core", "bin", "lib");
  const { projectProhibitions } = require(path.join(installed, "probe-core.cjs"));
  const { descriptorFromProjection } = require(path.join(installed, "prohibition-enforcement.cjs"));
  const projection = projectProhibitions([{
    statement: STATEMENTS[kind],
    status: "resolved",
    verification: "test",
    check_kind: "node-test",
    check_target: `${PHASE_REL}37-prohibition-check.test.cjs`,
    check_violation_fixture: `${PHASE_REL}${FIXTURES[kind].file}`,
    check_clean_fixture: `${PHASE_REL}${FIXTURES.none.file}`,
  }]);
  assert.equal(projection.length, 1, `${kind}: projectProhibitions must return one item`);
  const projected = projection[0];
  const check = descriptorFromProjection(projected);
  assert.deepEqual(check, {
    kind: "node-test",
    target: `${PHASE_REL}37-prohibition-check.test.cjs`,
    violationFixture: `${PHASE_REL}${FIXTURES[kind].file}`,
    cleanFixture: `${PHASE_REL}${FIXTURES.none.file}`,
  });
  return {
    projection: projected,
    request: { prohibition: projected, check, mode: "auto" },
  };
}

function planProhibitionItems() {
  const planPath = path.join(ROOT, PHASE_REL, "37-01-PLAN.md");
  const planText = fs.readFileSync(planPath, "utf8");
  const match = planText.match(/^  prohibitions:\n([\s\S]*?)(?=^  artifacts:)/m);
  assert.ok(match, "plan must retain a must_haves.prohibitions block");
  assert.ok(!/^\s{6}check:\s*$/m.test(match[1]), "plan descriptors must use flat check_* scalars");
  const items = [];
  let current = null;
  for (const line of match[1].split(/\r?\n/)) {
    const start = line.match(/^    - statement: "(.*)"$/);
    if (start) {
      if (current) items.push(current);
      current = { statement: start[1] };
      continue;
    }
    const field = line.match(/^      ([a-z_]+): (.+)$/);
    if (field && current) {
      let value = field[2];
      if (value.startsWith('"') && value.endsWith('"')) value = value.slice(1, -1);
      current[field[1]] = value;
    }
  }
  if (current) items.push(current);
  return items;
}

function assertPlanProjectionMatches(projectedItems) {
  const historyPath = path.join(ROOT, PHASE_REL, "37-PROHIBITION-PROJECTION.json");
  const history = JSON.parse(fs.readFileSync(historyPath, "utf8"));
  assert.deepEqual(history, [
    { statement: STATEMENTS.outcome, status: "unresolved" },
    { statement: STATEMENTS.causal, status: "unresolved" },
  ], "initial descriptorless unresolved projection history must remain intact");

  const planItems = planProhibitionItems();
  assert.equal(planItems.length, projectedItems.length, "plan must wire both recalled prohibitions");
  const fields = [
    "statement", "status", "verification", "check_kind", "check_target",
    "check_violation_fixture", "check_clean_fixture",
  ];
  for (let i = 0; i < projectedItems.length; i++) {
    assert.deepEqual(
      Object.keys(planItems[i]).sort(),
      Object.keys(projectedItems[i]).sort(),
      `plan prohibition ${i} flat fields must exactly match projectProhibitions`,
    );
    for (const field of fields) {
      assert.equal(planItems[i][field], projectedItems[i][field], `plan prohibition ${i} field ${field} must match projectProhibitions`);
    }
  }
}

function parseTapCounts(output) {
  const value = (name) => {
    const match = output.match(new RegExp(`^# ${name} (\\d+)$`, "m"));
    return match ? Number(match[1]) : -1;
  };
  return {
    tests: value("tests"),
    pass: value("pass"),
    fail: value("fail"),
    cancelled: value("cancelled"),
    skipped: value("skipped"),
    todo: value("todo"),
  };
}

function runDirectTap(kind, mode) {
  const env = { ...process.env, GOCACHE: "/tmp/ai-lang-verification-gocache", GOFLAGS: "" };
  delete env.NODE_TEST_CONTEXT;
  delete env.NODE_OPTIONS;
  delete env.GSD_PROHIB_SUBJECT;
  if (mode === "bad") env.GSD_PROHIB_SUBJECT = `${PHASE_REL}${FIXTURES[kind].file}`;
  if (mode === "clean") env.GSD_PROHIB_SUBJECT = `${PHASE_REL}${FIXTURES.none.file}`;
  const target = `${PHASE_REL}37-prohibition-check.test.cjs`;
  const result = spawnSync(process.execPath, ["--test", "--test-reporter=tap", "--", target], {
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
  const expected = { tests: 2, pass: bad ? 1 : 2, fail: bad ? 1 : 0, cancelled: 0, skipped: 0, todo: 0 };
  const failedNames = [...output.matchAll(/^not ok \d+ - (.+)$/gm)].map((match) => match[1]);
  const expectedFailure = bad ? [TEST_NAMES[kind]] : [];
  assert.equal(result.error, undefined, `${kind}/${mode}: Node test runner must spawn`);
  assert.equal(result.signal, null, `${kind}/${mode}: Node test runner must not be interrupted`);
  assert.equal(result.status, bad ? 1 : 0, `${kind}/${mode}: direct TAP exit status`);
  assert.deepEqual(counts, expected, `${kind}/${mode}: TAP test/pass/fail/skip counts`);
  assert.deepEqual(failedNames, expectedFailure, `${kind}/${mode}: only the named violation assertion may fail`);
  for (const testName of Object.values(TEST_NAMES)) {
    assert.ok(output.includes(testName), `${kind}/${mode}: missing non-vacuous test name ${testName}`);
  }
  return { output, counts, status: result.status, sha256: digest(output) };
}

function runGoRootControls(kind, mode) {
  const subject = mode === "bad" ? kind : "none";
  const roots = {};
  for (const root of ["outcome", "causal"]) {
    const result = runExactRoot(root, subject);
    const expectedKind = mode === "bad" && root === kind ? "violation" : "green";
    assert.equal(result.kind, expectedKind, `${kind}/${mode}: ${root} exact Go root classification`);
    const output = result.output || "";
    const spec = ROOTS[root];
    const name = spec.name.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
    const run = rootLineCount(output, new RegExp(`^=== RUN   ${name}$`));
    const pass = rootLineCount(output, new RegExp(`^--- PASS: ${name} \\([0-9]+(?:\\.[0-9]+)?s\\)$`));
    const fail = rootLineCount(output, new RegExp(`^--- FAIL: ${name} \\([0-9]+(?:\\.[0-9]+)?s\\)$`));
    const skip = rootLineCount(output, new RegExp(`^--- SKIP: ${name}(?:\\s|$)`));
    assert.equal(run, 1, `${kind}/${mode}: ${root} root must run exactly once`);
    assert.equal(pass + fail + skip, 1, `${kind}/${mode}: ${root} must end exactly once`);
    roots[root] = {
      run,
      pass,
      fail,
      skip,
      sha256: digest(output),
    };
  }
  return roots;
}

function runCanonicalProducer(kind, request) {
  const cli = path.join(os.homedir(), ".codex", "gsd-core", "bin", "gsd-tools.cjs");
  assert.ok(fs.existsSync(cli), `canonical GSD producer is missing at ${cli}`);
  const tempDir = fs.mkdtempSync(path.join(TMP_ROOT, "schway37-prohibition-request-"));
  const requestPath = path.join(tempDir, `${kind}-request.json`);
  const requestText = `${JSON.stringify(request, null, 2)}\n`;
  fs.writeFileSync(requestPath, requestText, { flag: "wx" });
  let result;
  try {
    result = spawnSync(process.execPath, [cli, "check", "prohibition-enforcement", requestPath, "--raw"], {
      cwd: ROOT,
      env: { ...process.env, GOCACHE: "/tmp/ai-lang-verification-gocache", GOFLAGS: "" },
      encoding: "utf8",
      timeout: 120_000,
      maxBuffer: MAX_OUTPUT,
      stdio: ["ignore", "pipe", "pipe"],
      shell: false,
      windowsHide: true,
    });
  } finally {
    fs.rmSync(tempDir, { recursive: true, force: true });
  }
  const stdout = typeof result.stdout === "string" ? result.stdout : "";
  const stderr = typeof result.stderr === "string" ? result.stderr : "";
  assert.equal(result.error, undefined, `${kind}: canonical producer must spawn`);
  assert.equal(result.signal, null, `${kind}: canonical producer must not be interrupted`);
  assert.equal(result.status, 0, `${kind}: canonical producer exit: ${stderr || stdout}`);
  const verdict = JSON.parse(stdout);
  assert.equal(verdict.status, "green", `${kind}: canonical producer status`);
  assert.equal(verdict.flagged, false, `${kind}: canonical producer must not flag`);
  assert.equal(verdict.tier, "test", `${kind}: canonical producer tier`);
  assert.equal(verdict.located, true, `${kind}: canonical producer must locate the descriptor`);
  assert.equal(verdict.kind, "node-test", `${kind}: canonical producer check kind`);
  assert.equal(verdict.evidence.length, 1, `${kind}: canonical producer must return one evidence row`);
  assert.equal(verdict.evidence[0].target, `${PHASE_REL}37-prohibition-check.test.cjs`);
  assert.equal(verdict.evidence[0].failFirst, true);
  assert.equal(verdict.evidence[0].passed, true);
  assert.equal(verdict.evidence[0].failFirstProof, "violation-fixture");
  return { verdict, stdout, stderr, requestText };
}

function writeProofEvidence(rows, projections, proofs) {
  const phaseDir = path.join(ROOT, PHASE_REL);
  const planPath = path.join(phaseDir, "37-01-PLAN.md");
  const adapterPath = path.join(phaseDir, "37-prohibition-check.test.cjs");
  const head = spawnSync("git", ["rev-parse", "HEAD"], { cwd: ROOT, encoding: "utf8", shell: false });
  assert.equal(head.status, 0, "capture current HEAD for evidence");
  const date = new Date().toISOString().slice(0, 10);
  const lines = [
    "# Phase 37 prohibition evidence",
    "",
    `Evidence date: ${date}`,
    `Full checkout HEAD: \`${head.stdout.trim()}\``,
    `Host/toolchain: ${os.platform()} ${os.arch()}; Node \`${process.version}\`; Go cache \`/tmp/ai-lang-verification-gocache\`.`,
    `Adapter SHA-256: \`${digest(fs.readFileSync(adapterPath))}\`.`,
    `Plan SHA-256 after descriptor wiring: \`${digest(fs.readFileSync(planPath))}\`.`,
    "",
    "This local proof runs the supported installed \`projectProhibitions\` projection, reconstructs each descriptor with \`descriptorFromProjection\`, and calls the canonical \`check prohibition-enforcement\` producer. The initial descriptorless/unresolved projection remains preserved in \`37-PROHIBITION-PROJECTION.json\`; this evidence proves only the two named existing Go boundaries. Structural causal acceptance is not CRC numeric replay.",
    "",
  ];
  for (let index = 0; index < rows.length; index++) {
    const kind = rows[index];
    const roots = ROOTS;
    const specDigests = {
      source: digest(fs.readFileSync(path.join(ROOT, roots[kind].source))),
      fixture: digest(fs.readFileSync(path.join(phaseDir, FIXTURES[kind].file))),
      cleanFixture: digest(fs.readFileSync(path.join(phaseDir, FIXTURES.none.file))),
    };
    const result = proofs[kind];
    lines.push(
      `## ${kind === "outcome" ? "Native capture outcome" : "Causal peer numeric-limit"} prohibition`,
      "",
      `Statement: \`${STATEMENTS[kind]}\``,
      `Source: \`${roots[kind].source}\` SHA-256 \`${specDigests.source}\`; adapter/test SHA-256 \`${digest(fs.readFileSync(adapterPath))}\`.`,
      `Violation fixture: \`${PHASE_REL}${FIXTURES[kind].file}\` SHA-256 \`${specDigests.fixture}\`; clean fixture: \`${PHASE_REL}${FIXTURES.none.file}\` SHA-256 \`${specDigests.cleanFixture}\`.`,
      `Canonical request SHA-256: \`${digest(result.requestText)}\`; producer stdout SHA-256: \`${digest(result.stdout)}\`; producer verdict: \`green\`, \`flagged=false\`, \`kind=node-test\`, \`failFirstProof=violation-fixture\`.`,
      "",
      "Canonical request:",
      "",
      "```json",
      result.requestText.trimEnd(),
      "```",
      "",
      "| Subject | Go root counts and output digests | Direct node-test TAP counts | Exit | TAP SHA-256 |",
      "| --- | --- | --- | ---: | --- |",
    );
    for (const mode of ["bad", "clean", "default"]) {
      const control = result.controls[mode];
      const rootText = Object.entries(control.roots)
        .map(([name, r]) => `${name}: run=${r.run}, pass=${r.pass}, fail=${r.fail}, skip=${r.skip}`)
        .join("; ");
      const rootDigests = Object.entries(control.roots).map(([name, r]) => `${name}=${r.sha256}`).join("; ");
      lines.push(`| ${mode} | ${rootText}; output SHA-256: ${rootDigests} | tests=${control.tap.counts.tests}, pass=${control.tap.counts.pass}, fail=${control.tap.counts.fail}, skipped=${control.tap.counts.skipped} | ${control.tap.status} | \`${control.tap.sha256}\` |`);
    }
    lines.push(
      "",
      `Exact named root: \`${roots[kind].package} -run '^${roots[kind].name}$' -count=1 -v -timeout=20s\`. Expected violation assertion(s): ${roots[kind].expectedAssertions.map((item) => `\`${item}\``).join("; ")}. The bad subject changes only its hardcoded temporary Go overlay and reaches only these assertions; tracked source bytes are compared before/after and temp overlays are removed. Clean/default controls both pass without skips.`,
      "",
    );
  }
  fs.writeFileSync(path.join(phaseDir, "37-PROHIBITION-EVIDENCE.md"), `${lines.join("\n").trimEnd()}\n`, { flag: "w" });
}

function runProofMode() {
  try {
    const kinds = ["outcome", "causal"];
    const projections = {};
    const proofs = {};
    for (const kind of kinds) {
      const { projection, request } = projectRequest(kind);
      projections[kind] = projection;
      assertPlanProjectionMatches(kinds.map((item) => item === kind ? projection : projectRequest(item).projection));
      const controls = {};
      for (const mode of ["bad", "clean", "default"]) {
        const tap = runDirectTap(kind, mode);
        const roots = runGoRootControls(kind, mode);
        controls[mode] = { tap, roots };
      }
      const producer = runCanonicalProducer(kind, request);
      proofs[kind] = { ...producer, controls };
    }
    assertPlanProjectionMatches(kinds.map((kind) => projections[kind]));
    writeProofEvidence(kinds, projections, proofs);
    process.stdout.write("Phase37 canonical prohibition proof passed for both projected descriptors.\n");
  } catch (error) {
    process.stderr.write(`Phase37 canonical prohibition proof failed: ${error.message}\n`);
    process.exitCode = 1;
  }
}

if (process.argv.includes("--prove")) {
  runProofMode();
} else {
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

  test(TEST_NAMES.outcome, { skip: invalidReason || false }, () => {
    assert.equal(results.outcome.kind, "green", proofMessage("outcome"));
  });

  test(TEST_NAMES.causal, { skip: invalidReason || false }, () => {
    assert.equal(results.causal.kind, "green", proofMessage("causal"));
  });
}
