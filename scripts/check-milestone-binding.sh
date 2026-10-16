#!/bin/sh
# M007 binding checks. Metadata is finite JSON data; this script never evaluates it.
set -eu
root=$(CDPATH= cd "$(dirname "$0")/.." && pwd -P)
exec python3 - "$root" "$@" <<'PY'
import datetime
import hashlib
import json
import os
import pathlib
import re
import stat
import subprocess
import sys

root = pathlib.Path(sys.argv[1]).resolve()
args = sys.argv[2:]
mode = args[0] if args else "usage"

class Refusal(Exception):
    def __init__(self, reason):
        self.reason = reason

def refuse(reason):
    raise Refusal(reason)

def fail(reason):
    marker = {"run": "binding", "freeze": "freeze-preflight", "final": "final"}.get(mode, mode)
    print(f"phase38 {marker}: status=fail reason={reason}", file=sys.stderr)
    raise SystemExit(1)

def safe_path(value):
    if not isinstance(value, str) or not value or "\\" in value:
        refuse("unsafe_path")
    path = pathlib.PurePosixPath(value)
    if path.is_absolute() or any(part in ("", ".", "..") for part in path.parts):
        refuse("unsafe_path")
    resolved = (root / pathlib.Path(*path.parts)).resolve()
    if resolved == root or root not in resolved.parents:
        refuse("unsafe_path")
    return resolved

def safe_regular_file(value, limit, missing_reason, type_reason):
    resolved = safe_path(value)
    cursor = root
    for part in pathlib.PurePosixPath(value).parts:
        cursor = cursor / part
        if cursor.is_symlink():
            refuse("unsafe_path")
    try:
        info = cursor.stat()
    except OSError:
        refuse(missing_reason)
    if not stat.S_ISREG(info.st_mode):
        refuse(type_reason)
    if info.st_size > limit:
        refuse("preparation_too_large")
    data = read_bytes(resolved, limit)
    return resolved, data

def read_bytes(path, limit=16 * 1024 * 1024):
    if path.stat().st_size > limit:
        refuse("input_too_large")
    data = path.read_bytes()
    if len(data) > limit:
        refuse("input_too_large")
    return data

def unique_object(pairs):
    out = {}
    for key, value in pairs:
        if key in out:
            refuse("duplicate_json_key")
        out[key] = value
    return out

def load_json(value):
    data = read_bytes(safe_path(value), 1024 * 1024)
    try:
        obj = json.loads(data.decode("utf-8"), object_pairs_hook=unique_object,
                         parse_constant=lambda _: refuse("invalid_json_number"))
    except (UnicodeDecodeError, json.JSONDecodeError):
        refuse("invalid_json")
    if type(obj) is not dict:
        refuse("json_object_required")
    return obj

def keys(obj, expected, reason):
    if type(obj) is not dict or set(obj) != set(expected):
        refuse(reason)

def positive_int(value):
    return type(value) is int and 0 < value <= 9223372036854775807

def hex_string(value, size):
    return isinstance(value, str) and re.fullmatch(r"[0-9a-f]{%d}" % size, value) is not None

def freeze_data(path):
    f = load_json(path)
    keys(f, {"schema", "freeze_sha", "codedigest", "manifest", "repository",
             "destination_ref", "workflow_path", "inputs", "preparation"}, "freeze_schema")
    if f["schema"] != "schway.m007-freeze/1": refuse("freeze_schema")
    if not hex_string(f["freeze_sha"], 40) or not hex_string(f["codedigest"], 64): refuse("freeze_identity")
    if f["repository"] != "szTheory/schway" or not isinstance(f["destination_ref"], str) or not f["destination_ref"].startswith("refs/heads/"):
        refuse("freeze_destination")
    if f["workflow_path"] != ".github/workflows/ci.yml": refuse("freeze_workflow")
    keys(f["manifest"], {"path", "sha256"}, "manifest_identity")
    if f["manifest"]["path"] != "testdata/receipts/M007.manifest" or not hex_string(f["manifest"]["sha256"], 64):
        refuse("manifest_identity")
    inputs = {"record_validation_corpus": False, "phase26_focused": False, "spike008_strict": False}
    if (type(f["inputs"]) is not dict or set(f["inputs"]) != set(inputs) or
        any(type(value) is not bool for value in f["inputs"].values()) or f["inputs"] != inputs):
        refuse("dispatch_inputs")
    prep_keys = {"review", "security", "validation", "audit"}
    if type(f["preparation"]) is not dict or set(f["preparation"]) != prep_keys:
        refuse("preparation_links")
    prep_paths = {
        "review": ".planning/phases/38-milestone-receipts/38-PRE-FREEZE-REVIEW.md",
        "security": ".planning/phases/38-milestone-receipts/38-PRE-FREEZE-SECURITY.md",
        "validation": ".planning/phases/38-milestone-receipts/38-VALIDATION.md",
        "audit": ".planning/M007-MILESTONE-AUDIT.md",
    }
    dispositions = {"review": "clear", "security": "secured", "validation": "focused", "audit": "EVD-17-pending"}
    marker_pattern = re.compile(rb"(?m)^<!-- phase38-preparation/1 kind=([a-z]+) source_sha=([0-9a-f]{40}) codedigest=([0-9a-f]{64}) disposition=([A-Za-z0-9-]+) -->$")
    for kind, record in f["preparation"].items():
        keys(record, {"path", "sha256", "source_sha", "codedigest", "disposition"}, "preparation_link")
        if (record["path"] != prep_paths[kind] or not hex_string(record["sha256"], 64) or
            not hex_string(record["source_sha"], 40) or not hex_string(record["codedigest"], 64) or
            not isinstance(record["disposition"], str)):
            refuse("preparation_link")
        if record["source_sha"] != f["freeze_sha"] or record["codedigest"] != f["codedigest"]:
            refuse("preparation_stale")
        if record["disposition"] != dispositions[kind]:
            refuse("preparation_open")
        path, data = safe_regular_file(record["path"], 1024 * 1024,
                                       "preparation_missing", "preparation_not_regular")
        if hashlib.sha256(data).hexdigest() != record["sha256"]:
            refuse("preparation_digest_mismatch")
        frontmatter = re.match(rb"\A---\r?\n.*?\r?\n---(?:\r?\n|$)", data, re.S)
        if frontmatter is None:
            refuse("preparation_record")
        body = data[frontmatter.end():]
        markers = marker_pattern.findall(body)
        marker_count = data.count(b"<!-- phase38-preparation/")
        if len(markers) != 1 or marker_count != 1:
            refuse("preparation_record")
        marker_kind, marker_sha, marker_digest, marker_disposition = markers[0]
        if marker_kind.decode("ascii") != kind:
            refuse("preparation_record")
        if marker_sha.decode("ascii") != f["freeze_sha"] or marker_digest.decode("ascii") != f["codedigest"]:
            refuse("preparation_stale")
        if marker_disposition.decode("ascii") != dispositions[kind]:
            refuse("preparation_open")
    manifest = safe_path(f["manifest"]["path"])
    if hashlib.sha256(read_bytes(manifest)).hexdigest() != f["manifest"]["sha256"]:
        refuse("manifest_digest_mismatch")
    return f

def manifest_inventory(path):
    data = read_bytes(safe_path(path))
    try:
        lines = data.decode("utf-8").splitlines()
    except UnicodeDecodeError:
        refuse("manifest_encoding")
    controls = None
    cases, roots, events, reached, structured = [], {}, {}, [], []
    for line in lines:
        if line.startswith("controls="):
            if controls is not None: refuse("control_inventory")
            controls = line[len("controls="):].split(",")
        elif line.startswith("# case "):
            fields = line.split(maxsplit=6)
            if len(fields) != 7: refuse("case_inventory")
            cases.append((fields[2], fields[3], line))
        elif line.startswith("# root "):
            fields = line.split()
            if len(fields) != 5 or fields[2] in roots: refuse("control_inventory")
            try: pattern = re.compile(fields[4])
            except re.error: refuse("control_inventory")
            roots[fields[2]] = (fields[3], pattern)
        elif line.startswith("# event "):
            fields = line.split()
            if len(fields) != 4: refuse("control_event_inventory")
            events.setdefault(fields[2], []).append(fields[3])
        elif line.startswith("# reached "):
            fields = line.split()
            if len(fields) != 6: refuse("reached_inventory")
            reached.append(tuple(fields[2:]))
        elif line.startswith("# structured "):
            fields = line.split()
            if len(fields) < 5: refuse("structured_inventory")
            structured.append(fields[2:])
    if (not controls or any(not x or re.search(r"\s", x) for x in controls) or
        len(set(controls)) != len(controls) or set(controls) != set(roots) or
        set(events) != set(controls)):
        refuse("control_inventory")
    case_keys = {(program, key) for program, key, _ in cases}
    if len(case_keys) != len(cases): refuse("case_inventory")
    old_data = read_bytes(root / "testdata/receipts/M006.manifest")
    old_rows = [line for line in old_data.decode("utf-8").splitlines() if line.startswith("# case ")]
    current_rows = [line for _, _, line in cases]
    if [line for line in current_rows if line in set(old_rows)] != old_rows:
        refuse("inherited_manifest_changed")
    old_keys = set()
    for line in old_rows:
        fields = line.split(maxsplit=6)
        if len(fields) != 7: refuse("case_inventory")
        old_keys.add((fields[2], fields[3]))
    event_sets = {}
    for name in controls:
        package, pattern = roots[name]
        names = events[name]
        if not package.startswith("./") or not names or len(set(names)) != len(names):
            refuse("control_event_inventory")
        for test_name in names:
            top = test_name.split("/", 1)[0]
            if not pattern.fullmatch(top) or re.search(r"[\x00-\x20]", test_name):
                refuse("control_event_inventory")
        event_sets[name] = set(names)
    reached_keys = set()
    reached_cases = {}
    for program, key, package, test_name in reached:
        if (program, key) not in case_keys: refuse("reached_inventory")
        matches = [name for name, tests in event_sets.items() if test_name in tests]
        if len(matches) != 1 or roots[matches[0]][0] != package:
            refuse("reached_inventory")
        item = (program, key, package, test_name)
        if item in reached_keys: refuse("reached_inventory")
        reached_keys.add(item)
        reached_cases.setdefault((program, key), set()).add(test_name)
    for program, key in case_keys - old_keys:
        names = reached_cases.get((program, key), set())
        if program in ("crc32", "crc_low"):
            if len(names) != 2 or not any("/interpreter" in name for name in names) or not any("/native" in name for name in names):
                refuse("reached_inventory")
        elif len(names) != 1:
            refuse("reached_inventory")
    for fields in structured:
        name, package, *tokens = fields
        refs = [token for token in tokens if "=" not in token]
        matches = [control for control, tests in event_sets.items()
                   if roots[control][0] == package and refs and all(ref in tests for ref in refs)]
        if len(matches) != 1: refuse("structured_inventory")
    return controls, roots, event_sets, cases, reached

package_import_cache = {}
def package_import_path(selector):
    if selector in package_import_cache:
        return package_import_cache[selector]
    if not selector.startswith("./") or re.search(r"[\x00-\x20]", selector):
        refuse("control_package")
    try:
        resolved = subprocess.run(["go", "list", "-f", "{{.ImportPath}}", selector],
                                  cwd=root, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                                  text=True, timeout=60)
    except (OSError, subprocess.TimeoutExpired):
        refuse("control_package")
    lines = resolved.stdout.splitlines()
    if resolved.returncode != 0 or len(lines) != 1 or not lines[0] or re.search(r"[\x00-\x20]", lines[0]):
        refuse("control_package")
    package_import_cache[selector] = lines[0]
    return lines[0]

def validate_control_events(manifest_path, log_data):
    controls, roots, event_sets, cases, reached = manifest_inventory(manifest_path)
    event_records = {}
    report_records = {}
    try:
        log_lines = log_data.decode("utf-8", "strict").splitlines()
    except UnicodeDecodeError:
        refuse("control_event_inventory")
    for line in log_lines:
        match = re.fullmatch(r"phase38-control-events-hex ([^ ]+) ([0-9a-f]+)", line)
        if match:
            name, encoded = match.groups()
            if name in event_records or len(encoded) > 32 * 1024 * 1024:
                refuse("control_event_inventory")
            try: event_records[name] = bytes.fromhex(encoded)
            except ValueError: refuse("control_event_inventory")
            continue
        match = re.fullmatch(r"phase38 control name=([^ ]+) package=([^ ]+) tests=(.+) status=pass", line)
        if match:
            name, package, test_list = match.groups()
            if name in report_records: refuse("control_event_inventory")
            report_records[name] = (package, test_list.split())
    if set(event_records) != set(controls) or set(report_records) != set(controls):
        refuse("control_event_inventory")
    for name in controls:
        package, _ = roots[name]
        expected_package = package_import_path(package)
        expected = event_sets[name]
        expected_roots = sorted(test for test in expected if "/" not in test)
        if report_records[name] != (package, expected_roots):
            refuse("control_event_inventory")
        raw = event_records[name]
        if not raw or not raw.endswith(b"\n") or len(raw) > 16 * 1024 * 1024:
            refuse("control_event_incomplete")
        counts = {test: {"run": 0, "pass": 0} for test in expected}
        package_lifecycle = {"start": 0, "pass": 0}
        package_pass_seen = False
        for raw_line in raw.splitlines():
            if not raw_line: refuse("control_event_inventory")
            try:
                item = json.loads(raw_line.decode("utf-8"), object_pairs_hook=unique_object,
                                  parse_constant=lambda _: refuse("control_event_inventory"))
            except (UnicodeDecodeError, json.JSONDecodeError):
                refuse("control_event_inventory")
            if (type(item) is not dict or type(item.get("Action")) is not str or
                type(item.get("Package")) is not str or item["Package"] != expected_package):
                refuse("control_event_inventory")
            action, test_name = item["Action"], item.get("Test", "")
            if type(test_name) is not str: refuse("control_event_inventory")
            if action != "start" and not package_lifecycle["start"]:
                refuse("control_event_inventory")
            if action == "start":
                if test_name or package_lifecycle["start"] or package_pass_seen:
                    refuse("control_event_inventory")
                package_lifecycle["start"] += 1
            elif action in ("run", "pass", "skip", "fail"):
                if not test_name:
                    if action == "fail": refuse("control_test_failed")
                    if action == "skip": refuse("control_test_skipped")
                    if action != "pass" or not package_lifecycle["start"] or package_pass_seen:
                        refuse("control_event_inventory")
                    package_lifecycle["pass"] += 1
                    package_pass_seen = True
                    continue
                if not package_lifecycle["start"] or package_pass_seen:
                    refuse("control_event_inventory")
                if test_name not in expected: refuse("control_event_inventory")
                if action == "skip": refuse("control_test_skipped")
                if action == "fail": refuse("control_test_failed")
                counts[test_name][action] += 1
                if counts[test_name][action] > 1: refuse("control_event_inventory")
            elif action == "output":
                if package_pass_seen: refuse("control_event_inventory")
                if test_name and test_name not in expected: refuse("control_event_inventory")
            else:
                refuse("control_event_inventory")
        if (package_lifecycle["start"] != 1 or package_lifecycle["pass"] != 1 or
            any(counts[test]["run"] != 1 or counts[test]["pass"] != 1 for test in expected)):
            refuse("control_event_incomplete")
    return len(controls), len(cases)

def validate_run(f, run_path, log_args):
    r = load_json(run_path)
    keys(r, {"schema", "repository", "workflow_path", "event", "run_id", "attempt", "url",
             "ref", "head_sha", "status", "conclusion", "jobs", "artifacts", "logs"}, "run_schema")
    if r["schema"] != "schway.m007-hosted-run/1": refuse("run_schema")
    if (r["repository"] != f["repository"] or r["workflow_path"] != f["workflow_path"] or
        r["event"] != "workflow_dispatch" or r["ref"] != f["destination_ref"]): refuse("run_identity")
    if not positive_int(r["run_id"]) or not positive_int(r["attempt"]): refuse("run_identity")
    expected_url = f"https://github.com/{f['repository']}/actions/runs/{r['run_id']}/attempts/{r['attempt']}"
    if r["url"] != expected_url: refuse("run_identity")
    if r["status"] != "completed" or r["conclusion"] != "success": refuse("run_not_successful")
    if r["head_sha"] != f["freeze_sha"]: refuse("wrong_revision")
    if type(r["jobs"]) is not list or len(r["jobs"]) != 4: refuse("required_jobs")
    wanted = {
        "checks (ubuntu-latest)": {"full": "success", "race": "success"},
        "checks (macos-latest)": {"full": "success", "race": "success"},
        "current evidence aggregate (ubuntu-latest)": None,
        "current evidence aggregate (macos-latest)": None,
    }
    found = {}
    for job in r["jobs"]:
        keys(job, {"id", "name", "conclusion", "steps"} if "steps" in job else {"id", "name", "conclusion"}, "job_schema")
        if not positive_int(job["id"]) or job["name"] in found: refuse("required_jobs")
        found[job["name"]] = job
    if set(found) != set(wanted): refuse("required_jobs")
    if len({found[name]["id"] for name in found}) != 4: refuse("required_jobs")
    for name, steps in wanted.items():
        job = found[name]
        if job["conclusion"] != "success": refuse("required_jobs")
        if steps is not None and job.get("steps") != steps: refuse("required_steps")
        if steps is None and "steps" in job: refuse("job_schema")
    if type(r["artifacts"]) is not list or len(r["artifacts"]) != 2: refuse("required_artifacts")
    if type(r["logs"]) is not list or len(r["logs"]) != 2: refuse("required_logs")
    host_os = {"Linux": "ubuntu-latest", "macOS": "macos-latest"}
    artifacts = {}
    for a in r["artifacts"]:
        keys(a, {"name", "id", "archive_digest", "expires_at", "run_id", "attempt"}, "artifact_schema")
        if not positive_int(a["id"]) or a["id"] in {x["id"] for x in artifacts.values()}: refuse("required_artifacts")
        if not positive_int(a["run_id"]) or not positive_int(a["attempt"]): refuse("mixed_attempt")
        if a["run_id"] != r["run_id"] or a["attempt"] != r["attempt"]: refuse("mixed_attempt")
        if not isinstance(a["archive_digest"], str) or re.fullmatch(r"sha256:[0-9a-f]{64}", a["archive_digest"]) is None: refuse("artifact_digest")
        try: expiry = datetime.datetime.fromisoformat(a["expires_at"].replace("Z", "+00:00"))
        except (ValueError, AttributeError): refuse("artifact_expiry")
        if expiry.tzinfo is None or expiry <= datetime.datetime.now(datetime.timezone.utc): refuse("artifact_expired")
        matching = [host for host, os_name in host_os.items() if a["name"] == f"phase38-receipt-{os_name}-{f['freeze_sha']}-{r['run_id']}-{r['attempt']}"]
        if len(matching) != 1 or matching[0] in artifacts: refuse("required_artifacts")
        artifacts[matching[0]] = a
    if set(artifacts) != set(host_os): refuse("required_artifacts")
    by_host = {}
    arg_paths = [safe_path(x) for x in log_args]
    if len(set(arg_paths)) != 2: refuse("required_logs")
    manifest_path = f["manifest"]["path"]
    for item in r["logs"]:
        keys(item, {"artifact_id", "host", "path", "bytes", "sha256"}, "log_schema")
        if type(item["host"]) is not str or item["host"] not in host_os or item["host"] in by_host: refuse("required_logs")
        if not positive_int(item["artifact_id"]): refuse("artifact_log_membership")
        if item["artifact_id"] != artifacts[item["host"]]["id"]: refuse("artifact_log_membership")
        if not positive_int(item["bytes"]) or not hex_string(item["sha256"], 64): refuse("log_identity")
        path = safe_path(item["path"])
        if path not in arg_paths: refuse("log_path_mismatch")
        data = read_bytes(path)
        if len(data) != item["bytes"] or hashlib.sha256(data).hexdigest() != item["sha256"]:
            refuse("log_bytes_mismatch")
        if b"phase38 probe result status=pass" in data: refuse("probe_not_full_receipt")
        identity = re.search(rb"(?m)^phase38 host identity host=([^ ]+) run_id=([0-9]+) attempt=([0-9]+) .* headsha=([^ ]+) codedigest=([^ ]+) tree=([^ ]+)$", data)
        if not identity: refuse("host_identity")
        host, log_run_id, log_attempt, head, digest, tree = [x.decode("ascii", "strict") for x in identity.groups()]
        if log_run_id != str(r["run_id"]): refuse("run_identity")
        if log_attempt != str(r["attempt"]): refuse("mixed_attempt")
        if host != item["host"] or head != f["freeze_sha"]: refuse("wrong_revision")
        if digest != f["codedigest"]: refuse("wrong_codedigest")
        if tree != "clean": refuse("dirty_host")
        validate_control_events(manifest_path, data)
        by_host[item["host"]] = path
    if set(by_host) != set(host_os): refuse("required_logs")
    # The generic gate remains the authority for manifest cases, controls and host-row consistency.
    env = os.environ.copy()
    env["EXPECTED_CODEDIGEST"] = f["codedigest"]
    a, b = by_host["Linux"], by_host["macOS"]
    checked = subprocess.run(["sh", str(root / "scripts/check-receipts.sh"),
                              str(safe_path(f["manifest"]["path"])), str(a), str(b)],
                             cwd=root, env=env, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    if checked.stdout:
        sys.stdout.write(checked.stdout)
    if checked.returncode != 0:
        refuse("receipt_gate_rejected")
    return r

try:
    if mode == "run" and len(args) == 5:
        freeze = freeze_data(args[1])
        validate_run(freeze, args[2], args[3:5])
        print("phase38 binding: status=pass")
    elif mode == "freeze" and len(args) == 2:
        freeze = freeze_data(args[1])
        candidate = subprocess.run(["git", "rev-parse", "HEAD"], cwd=root, check=True, text=True, stdout=subprocess.PIPE).stdout.strip()
        if not hex_string(candidate, 40) or subprocess.run(["git", "cat-file", "-e", freeze["freeze_sha"] + "^{commit}"], cwd=root, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL).returncode:
            refuse("freeze_commit_unknown")
        committed_digest = subprocess.run(["sh", "scripts/codedigest.sh", "--rev", freeze["freeze_sha"]], cwd=root, check=True, text=True, stdout=subprocess.PIPE).stdout.strip()
        if committed_digest != freeze["codedigest"]: refuse("freeze_digest_mismatch")
        if subprocess.run(["git", "merge-base", "--is-ancestor", freeze["freeze_sha"], candidate], cwd=root).returncode:
            refuse("freeze_not_ancestor")
        changed = subprocess.run(["sh", "scripts/check-freeze.sh", freeze["manifest"]["path"], freeze["freeze_sha"], candidate], cwd=root, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
        if changed.returncode:
            sys.stdout.write(changed.stdout)
            refuse("freeze_history_changed")
        paths = subprocess.run(["sh", "scripts/codedigest.sh", "--paths"], cwd=root, check=True, text=True, stdout=subprocess.PIPE).stdout.splitlines()
        status = subprocess.run(["git", "status", "--porcelain", "--untracked-files=all", "--", *paths], cwd=root, check=True, text=True, stdout=subprocess.PIPE).stdout
        if status: refuse("digest_worktree_dirty")
        if (root / ".planning/config.json").exists():
            config = hashlib.sha256((root / ".planning/config.json").read_bytes()).hexdigest()
            if config != "148ccc4c56db8516788c264216ae277d9860d43e00b92de299d8ccff2fd59521": refuse("config_changed")
            staged = subprocess.run(["git", "diff", "--cached", "--name-only"], cwd=root, check=True, text=True, stdout=subprocess.PIPE).stdout.splitlines()
            if ".planning/config.json" in staged: refuse("config_staged")
        print("phase38 freeze-preflight: status=pass")
    elif mode == "final" and len(args) == 5:
        # Plan 03 extends this final check with ingest-base and planning-only delta checks.
        freeze = freeze_data(args[1])
        run = validate_run(freeze, args[2], args[3:5])
        print(f"phase38 final: status=pass candidate={run['head_sha']}")
    else:
        refuse("usage")
except Refusal as exc:
    fail(exc.reason)
PY
