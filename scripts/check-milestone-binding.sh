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
    if f["inputs"] != inputs or type(f["inputs"]) is not dict:
        refuse("dispatch_inputs")
    prep_keys = {"review", "security", "validation", "audit"}
    if type(f["preparation"]) is not dict or set(f["preparation"]) != prep_keys:
        refuse("preparation_links")
    for record in f["preparation"].values():
        keys(record, {"path", "sha256", "source_sha", "codedigest"}, "preparation_link")
        if not isinstance(record["path"], str) or not hex_string(record["sha256"], 64): refuse("preparation_link")
        if record["source_sha"] != f["freeze_sha"] or record["codedigest"] != f["codedigest"]:
            refuse("preparation_stale")
    manifest = safe_path(f["manifest"]["path"])
    if hashlib.sha256(read_bytes(manifest)).hexdigest() != f["manifest"]["sha256"]:
        refuse("manifest_digest_mismatch")
    return f

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
    for item in r["logs"]:
        keys(item, {"artifact_id", "host", "path", "bytes", "sha256"}, "log_schema")
        if item["host"] not in host_os or item["host"] in by_host: refuse("required_logs")
        if item["artifact_id"] != artifacts[item["host"]]["id"]: refuse("artifact_log_membership")
        if not positive_int(item["bytes"]) or not hex_string(item["sha256"], 64): refuse("log_identity")
        path = safe_path(item["path"])
        if path not in arg_paths: refuse("log_path_mismatch")
        data = read_bytes(path)
        if len(data) != item["bytes"] or hashlib.sha256(data).hexdigest() != item["sha256"]:
            refuse("log_bytes_mismatch")
        if b"phase38 probe result status=pass" in data: refuse("probe_not_full_receipt")
        identity = re.search(rb"(?m)^phase38 host identity host=([^ ]+) .* headsha=([^ ]+) codedigest=([^ ]+) tree=([^ ]+)$", data)
        if not identity: refuse("host_identity")
        host, head, digest, tree = [x.decode("ascii", "strict") for x in identity.groups()]
        if host != item["host"] or head != f["freeze_sha"]: refuse("wrong_revision")
        if digest != f["codedigest"]: refuse("wrong_codedigest")
        if tree != "clean": refuse("dirty_host")
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
        changed = subprocess.run(["sh", "scripts/check-freeze.sh", freeze["manifest"]["path"], freeze["freeze_sha"], candidate, freeze["freeze_sha"]], cwd=root, text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
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
