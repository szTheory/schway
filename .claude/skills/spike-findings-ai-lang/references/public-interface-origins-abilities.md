# Public Interface: Origins and Abilities

What crosses a separate-compilation boundary so a **body-blind** consumer can
check borrowed and generic results.

## Requirements

From the `ownership-kernel` idea (MANIFEST.md):

- Export borrowed-result origins as verified value/field paths and tagged
  alternatives, with explicit shared/exclusive access in interface metadata.
- Keep `copy`, `drop`, `share`, `send`, and `escape` independent and derive their
  generic propagation structurally.
- Use higher-ranked fresh origins only at genuinely scoped callback/lending
  boundaries; do not force all borrowed results through callback-only APIs.
- Do not infer final source syntax or production performance from kernel
  notation.

## How to Build It

**Name public origins with parameter and field paths, not compiler-chosen region
binders.** The selected default:

```text
fn head(input: borrow Buffer) -> borrow(input) View[Byte]

fn choose(
  left: borrow Buffer,
  right: borrow Buffer,
) -> borrow(left | right) View[Byte]

fn choose_tagged(
  left: borrow Buffer,
  right: borrow Buffer,
) -> Left(borrow(left) View[Byte]) | Right(borrow(right) View[Byte])

fn left_field(pair: borrow PairBuffer) -> borrow(pair.left) View[Byte]
```

**Carry access separately from origin.** Exclusivity is its own fact:

```text
fn head_mut(input: borrow mut Buffer) -> borrow mut(input) View[Byte]
```

**Introduce a quantifier only at the pressure point** — an API that genuinely
lends fresh scoped access to caller code:

```text
fn with_view[R](
  input: borrow Buffer,
  use: for access fn(view: borrow(access) View[Byte]) -> R
    where R: escape,
) -> R
```

The spelling above is experimental. **The stable decision is the normalized
contract**: origin paths and alternatives, shared/exclusive access, a fresh
callback origin, and independently derived abilities.

**Prefer tagged alternatives over an unrefined union where precision matters.**
An origin union correctly keeps both possible owners borrowed — it is a safe
conservative join. A tagged return instead attaches `left` and `right` to
separate alternatives, letting the inactive owner end after pattern refinement.
A union is not a substitute for a path-sensitive result type.

**Derive abilities structurally and keep the five independent.** `copy`, `drop`,
`share`, `send`, `escape` each propagate on their own through generic `Option`,
`Result`, `List`, `View`, iterator, and unique-box wrappers. Spike 003 checked
all **32 atomic combinations** through two independently written recursive
evaluators.

**Preserve origins and abilities through type erasure.** An erased view keeps its
origin and its lack of `escape`. Erasing a data representation is **not**
permission to erase lifetime or authority facts.

**Verify summaries against bodies before publication.** Producer verification
rejects, before an interface can be published: under-declared origins,
borrowed-as-owned results, exclusive-as-shared results, and retained fresh
callback origins.

**Deep-copy at the interface boundary.** See What to Avoid — this one is a
concrete production instruction, not a workbench detail.

## What to Avoid

- **Shallow-copying the exported interface.** The exporter in spike 003 stripped
  bodies but shallow-copied return slices, so mutating the interface mutated the
  oracle's implementation facts and quietly changed later measurements. A
  regression test now proves fault injection cannot reach the oracle. **The
  production compiler needs immutable, content-owned summary representations —
  not merely different top-level structs.**
- **Inferred body-only relationships.** Rejected: separate consumers cannot check
  without bodies, and implementation changes silently alter API legality.
- **Wildcard / erased origins as an ordinary default.** Mojo documents the cost —
  wildcard origins can disable early destruction and obscure exclusivity. Unsafe
  boundary only.
- **Explicit region binders as the leading source surface.** Retain as a possible
  core/formal representation. They introduce compiler-chosen names and
  constraints into routine interfaces.
- **Converting every lending API into a `with_*` callback.** Callback-only
  encoding changed 8 of 13 direct APIs in the corpus. Keep it as a targeted tool
  for scoped resources and callbacks.
- **Treating token count as a usability verdict.** Over a deliberately small
  corpus: value-origin syntax 244 tokens / 1 explicit binder; explicit regions
  272 tokens / 13 binders; callback-only 361 tokens. These show **representation
  cost**, not human or AI superiority. AI generation/repair and human audit
  comparisons belong to the later surface experiment.
- **Letting library authors assert positive abilities freely.** Ability rules
  were *trusted inputs* to this spike. The real compiler must derive positive
  abilities from fields or require an audited unsafe witness.

## Constraints

- No subtyping, variance, recursive types, associated types, or generic
  constraint solver in the workbench.
- Field origins are represented as paths, but **disjoint-field conflict precision
  is not modeled**.
- Higher-ranked evidence covers **one** callback pattern. Borrows returned from
  nested callbacks, lending iterators, and async suspension remain open.
- Type erasure is semantic metadata in the fixture — not a vtable, object layout,
  or native ABI.
- A 1,000-source union produced a ~79 KB JSON interface, consumed in low
  single-digit milliseconds. This demonstrates **linear work**, not an acceptable
  public design. Large repeated origin sets may need interned DAGs or named
  summary nodes, and enormous flat unions need a diagnostic/summary compression
  policy before they are realistic APIs.

## Measured result (spike 003, VALIDATED)

24 deterministic call fixtures pass · 4 dishonest producer summaries rejected ·
66 generated consumer/oracle cases agree · 500 seeded trials over function order,
origin order, and caller alpha-renaming preserve results · injected fault
(removing `right` from `choose`'s origin set) surfaces after 19 generated cases
as `ownership.borrow_conflict`.

## Origin

Synthesized from spikes: 003
Source files available in: `sources/003-public-origins-generic-abilities/`
