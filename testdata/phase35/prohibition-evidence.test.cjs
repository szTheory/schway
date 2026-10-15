// Standard-library adapter for GSD's node-test descriptor. The Go checker
// owns the evidence rules and negative controls; this file only bridges TAP.
const { test } = require('node:test');
const assert = require('node:assert/strict');
const { spawnSync } = require('node:child_process');
const path = require('node:path');

test('Phase 35 evidence rejects false causal and lane claims', () => {
  const result = spawnSync('go', [
    'test', './internal/compiler/session',
    '-run', '^TestPhase35EvidenceGrounding$', '-count=1', '-json', '-timeout=60s',
  ], {
    cwd: path.resolve(__dirname, '../..'),
    env: { ...process.env, GOCACHE: process.env.GOCACHE || '/tmp/ai-lang-verification-gocache' },
    encoding: 'utf8', timeout: 65000, maxBuffer: 4 * 1024 * 1024,
  });
  assert.ifError(result.error);
  assert.equal(result.status, 0, result.stdout + result.stderr);
  const events = result.stdout.trim().split('\n').map(line => JSON.parse(line));
  const root = events.filter(event => event.Test === 'TestPhase35EvidenceGrounding');
  assert.ok(root.some(event => event.Action === 'run'), 'Go evidence test did not run');
  assert.ok(root.some(event => event.Action === 'pass'), 'Go evidence test did not pass');
  assert.ok(!events.some(event => event.Action === 'fail' || event.Action === 'skip'),
    'evidence check failed or skipped');
});
