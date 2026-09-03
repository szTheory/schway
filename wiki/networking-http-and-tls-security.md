---
id: networking-http-tls-security
title: Networking, HTTP, and TLS security boundary
summary: A first-party networking posture that keeps crypto replaceable, protocol parsing strict, unsafe compatibility explicit, and resource limits observable.
type: design
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-02
tags: [networking, http, tls, security, toolkit]
related: [effects-and-capabilities, context-telemetry-security, resources-locks-caching, concurrency-memory, memory-reclamation-policy, runtime-profiles-dogfooding, application-architecture-data, performance-observability-delivery, convergence-work-program, research-ledger]
---

# Networking, HTTP, and TLS security boundary

## Reader and outcome

This note is for the runtime, networking-kit, and security teams. After reading
it, they should be able to expose a safe HTTP/TLS happy path, integrate an
audited provider, and reject dangerous compatibility without turning crypto or
web policy into core language syntax.

## Boundary decision

Lang should make secure networking turnkey, but it should not implement
cryptographic primitives in the compiler, freeze TLS into the language, or ask
applications to assemble protocol safety from dozens of switches.

```text
core language
  ownership + resources + effects + Secret<T> + validation + budgets
        |
official networking kit
  typed endpoints + strict HTTP + pools + deadlines + TLS policy
        |
narrow provider boundary
  audited TLS protocol/crypto implementation + platform trust integration
        |
independently versioned security updates
```

The distribution should choose one maintained default provider after an
implementation and security review. Alternative providers exist for platform,
licensing, compliance, or embedded constraints, but an application chooses a
provider at its composition root—not a cipher for every socket. Rustls is a
useful model because it makes the cryptographic-provider boundary explicit and
documents that secure operation still depends on that provider's randomness,
cryptography, and zeroization. ([rustls security
boundary](https://github.com/rustls/rustls/security))

This is one place where the strangler-fig instinct stops: replacing a mature
cryptographic implementation in-house is not a maturity milestone. The useful
in-house asset is the small, typed, testable boundary and secure policy layer.

## Secure TLS happy path

```text
provide Tls = Tls.system(
  policy: TlsPolicy.modern(),
  trust: Trust.system_roots(),
  identity: SecretStore.keypair(name: "checkout-api"),
) lifetime service

provide Http = Http.server(
  listen: Endpoint.tcp(host: "0.0.0.0", port: 8443),
  tls: Tls,
  limits: HttpLimits.service_default(),
) lifetime service
```

`modern()` is a versioned policy record, not a promise that today's algorithms
remain safe forever. Its build evidence records the networking-kit version,
provider/version, policy revision, trust source, and enabled compatibility.
Security advisories can force a release-gate update even when source APIs have
not changed.

Current IETF Best Current Practice requires implementations not to negotiate
SSLv2, SSLv3, TLS 1.0, or TLS 1.1; it requires TLS 1.2 support and recommends
TLS 1.3 with preference when supported. It also notes that new secure transport
protocols should use TLS 1.3. The official policy must track the current BCP
rather than copy these choices permanently into grammar. ([RFC
9325](https://www.rfc-editor.org/rfc/rfc9325.html))

The default contract is:

- authenticated TLS, never opportunistic plaintext fallback;
- certificate-path and service-identity verification enabled together;
- the expected service identity is derived independently from the certificate
  the peer presents;
- ALPN/protocol negotiation is typed and cannot silently select a different
  application protocol;
- private keys and session secrets are non-printable, non-serializable owned
  handles with explicit rotation and zeroization obligations;
- TLS-level compression is off; application compression has independent
  disclosure and decompression limits;
- session tickets, resumption, key rotation, revocation posture, and trust-store
  reload are visible runtime policy and evidence;
- `insecure_skip_verify` does not exist in releasable code.

RFC 9525 requires clients to construct acceptable reference identifiers
independently and match the presented service identity; applications also need
full certificate-path validation. ([RFC
9525](https://www.rfc-editor.org/rfc/rfc9525.html))

Tests sometimes need self-signed peers or deterministic keys. That uses a
test-only `InsecureTestTls` capability that cannot satisfy a production
provider requirement and makes `release` fail if reachable.

## 0-RTT is an authority question

TLS 1.3 early data is disabled by default. RFC 8446 states that TLS does not
provide inherent replay protection for 0-RTT and that some duplicate-message
threats must be handled by the application. “GET” or an `Idempotency-Key`
header alone is not proof of replay safety. ([RFC
8446](https://www.rfc-editor.org/rfc/rfc8446.html), [RFC
8470](https://www.rfc-editor.org/rfc/rfc8470.html))

Enabling early data therefore requires a narrow route-level witness:

```text
route lookup_order
  accepts Get("/orders/{id}")
  early_data: replay_safe(
    authority: read_only,
    response: no_secret_timing_or_state_change,
  )
```

Commands, payments, mutations, secret-bearing responses, rate-limit-sensitive
operations, and routes without explicit anti-replay analysis use ordinary
1-RTT. The compiler can check declared effects; only tests and operational
policy can validate the complete distributed anti-replay design.

## HTTP parsing and resource safety

HTTP belongs in an independently versioned official kit because protocol and
security practice evolve. The core contributes strict parsing tools, byte/text
types, ownership, streaming, limits, and effects.

The HTTP happy path is intentionally less tolerant than historical parsers:

- reject ambiguous message framing instead of guessing;
- reject conflicting `Content-Length` values and requests containing both
  `Transfer-Encoding` and `Content-Length`;
- normalize once according to the selected protocol role and retain raw bytes
  only behind a classified diagnostic capability;
- give every header section, field, URI, body, trailer, chunk, decompressed
  output, connection, queue, and request a size/count/time budget;
- stream bodies under demand/backpressure rather than buffer by default;
- apply authentication/authorization only after the complete header section is
  parsed and validated;
- treat EOF, timeout, cancellation, and incomplete framing as distinct typed
  outcomes;
- make proxy/intermediary and origin roles explicit because their parsing and
  forwarding obligations differ.

RFC 9112 identifies parsing differences between recipients as the basis of
request smuggling and treats conflicting transfer/content lengths as an error.
RFC 9110 requires servers to reject oversized fields rather than silently ignore
them when doing so would alter semantics or framing. ([RFC
9112](https://www.rfc-editor.org/rfc/rfc9112.html), [RFC
9110](https://www.rfc-editor.org/rfc/rfc9110.html))

Strictness must be end-to-end: a perfectly strict backend can still be exposed
when a proxy, CDN, gateway, and server normalize the same bytes differently.
The contract-test kit therefore sends the malformed corpus through the deployed
chain, not only the local parser.

## Network authority and SSRF

An HTTP client is an ambient network capability unless constrained. The safe
provider can restrict schemes, ports, DNS zones, IP ranges, redirects, proxies,
credential forwarding, and response sizes:

```text
provide ShippingApi = Http.client(
  origin: HttpsOrigin("https://shipping.example"),
  egress: Egress.only_public_dns(
    deny: [loopback, link_local, private_network, metadata_service],
  ),
  redirects: Redirects.same_origin(max: 2),
  limits: HttpLimits.external_api(),
) lifetime service
```

The provider resolves and rechecks addresses according to policy at connection
time, so DNS changes cannot turn an approved public name into an internal
target unnoticed. Redirects re-evaluate authority and strip credentials unless
the policy explicitly preserves them. Proxy configuration is deploy input with
provenance, not an ambient environment surprise.

For servers behind a proxy, forwarded identity, scheme, host, and client-address
headers are untrusted unless the immediate peer matches a typed `TrustedProxy`
policy. Otherwise they are ignored rather than half-trusted.

## Connections, retries, and overload

Connections are lexical resources owned by bounded pools. Pool policy names
capacity, queue size, acquisition deadline, idle/lifetime limits, per-origin
partitioning, health validation, and shutdown. DNS resolution, connect,
handshake, headers, body idle, whole request, and pool wait have distinct
deadlines inside one total budget.

Retries are not a hidden HTTP-client feature. They require an error class,
remaining deadline, one owning layer, attempt budget, backoff/jitter policy,
and evidence that repeating the application operation is safe. A connection
failure before the client observes a response does not prove the server did not
commit a command.

The runtime emits bounded semantic events for negotiation, verification,
framing rejection, pool pressure, queue wait, deadline exhaustion, retry,
redirect, byte counts, and protocol downgrade. It never emits credentials,
keys, complete URLs with secret queries, or bodies by default.

## Security update and escape policy

- Networking kit, TLS policy database, trust integration, and provider are
  independently versioned from the language grammar and compiler.
- Builds are reproducible because exact versions and policy revisions are in
  the lock/evidence manifest; releases are safe because the toolchain checks
  those pins against signed advisories and supported-policy windows.
- Compatibility exceptions name a reason, owner, expiry, affected peer, and
  telemetry. There is no permanent “legacy mode.”
- Raw sockets and custom protocol parsers remain possible behind explicit
  effects and review obligations. Custom cryptographic primitives do not enter
  the safe official path merely because they compile.
- HTTP/2, HTTP/3, QUIC, WebSocket, and platform TLS adapters must pass the same
  authority, limit, interop, and malformed-input obligations before promotion.

## Verification gauntlet

The official kit does not graduate on a happy-path request. It needs:

1. protocol conformance and known-answer vectors;
2. grammar-aware fuzzing of pre-authentication and post-authentication state;
3. differential parsing through supported proxies and upstream servers;
4. malformed framing, header, chunk, URI, and compression corpora;
5. certificate expiry, wrong identity, chain failure, rotation, clock skew,
   trust reload, and mutual-TLS cases;
6. 0-RTT replay and retry/duplicate-command attacks;
7. slow clients, floods, queue saturation, cancellation, half-close, reset,
   and decompression/resource-exhaustion tests;
8. SSRF, redirect, DNS rebinding, proxy-header spoofing, credential-forwarding,
   and cross-tenant pool/cache tests;
9. provider upgrade, rollback, mixed-version, and security-advisory drills;
10. memory, CPU, handshake, throughput, and p99.9 latency evidence under load.

Rustls treats network input as fully attacker-controlled and fuzzes protocol
paths using a mock cryptographic provider; that separation is a useful testing
precedent while leaving real-provider validation mandatory. ([rustls security
boundary](https://github.com/rustls/rustls/security))

## Anti-patterns

- writing cryptographic primitives as a language bootstrap milestone;
- exporting raw cipher/version knobs as the ordinary application API;
- disabling verification because a development certificate is inconvenient;
- trusting certificate chain validation without service-identity matching;
- enabling 0-RTT for all nominally idempotent methods;
- permissive parser recovery across security boundaries;
- unlimited headers, bodies, decompression, redirects, retries, pools, queues,
  or handshake work;
- trusting forwarded headers from arbitrary peers;
- retrying a command because no response arrived;
- pinning a vulnerable provider forever in the name of reproducibility;
- silently changing the provider or security policy without build evidence.
