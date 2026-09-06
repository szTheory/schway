/* lang_foreign_retained.c
 *
 * The hand-written, byte-frozen, FOURTH separately-compiled foreign
 * translation unit -- a deliberately hostile variant of
 * lang_foreign_resource.c (D-05-06/D-05-09). Per D-04-10, the Codename
 * Lang compiler forms its whole contract with the symbol below from Lang
 * source alone -- it never parses this file's internals. Quarantine is
 * permanent and non-discharging.
 *
 * This file is frozen from the commit that adds it, in the same class as
 * lang_foreign_resource.c, lang_foreign_nonlocal.c, and
 * lang_foreign_arena.c: it is a golden fixture, not ordinary source under
 * active development.
 *
 * ONE declared symbol, call-count dispatch (a documented, verified
 * deviation from an earlier three-distinct-symbol draft): check.go's
 * resource-lifecycle shape independently resolves each step's OWN
 * declared symbol at admission time, but core.Function carries exactly
 * ONE ForeignContract, and cgen.go's generated C accordingly calls that
 * SAME symbol for every step in the sequence -- confirmed empirically by
 * dumping the generated C for a three-distinct-symbol draft of this
 * fixture, which called only the first declared symbol three times. This
 * is precisely lang_foreign_nonlocal.c's own established precedent
 * (D-04-17/D-04-18): ONE declared symbol, behavior keyed on a `static`
 * call counter, shared between this file and testdata/phase5/
 * retained_pointer.lang's three sequential calls.
 *
 * D-05-06's FFI-003 retained-pointer lifetime shape: the FIRST call
 * (`lang_retained_open`, below, an internal helper -- see the public
 * dispatcher `_LANG_lang_retained_touch`) genuinely allocates a one-byte
 * buffer via libc malloc and stashes its pointer in `static` storage --
 * entirely internal to this translation unit, needing no calls back into
 * Lang (M001 has none by design). The SECOND call reads the stashed
 * pointer while it is still live (an ordinary "touch"). The THIRD call
 * (`lang_retained_stale`, below) frees the SAME stashed pointer and
 * immediately reads through it again in the same call -- a genuine
 * heap-use-after-free, reached through a real Lang program's declared
 * foreign calls (testdata/phase5/retained_pointer.lang), never injected
 * into cgen's own release emission (already attacked independently by
 * Phase 4's release-omitted/release-order-transposed controls) and never
 * a standalone hostile C file unrelated to the boundary.
 *
 * The dangling read writes directly into the public wrapper's own
 * returned struct field. Because this translation unit is compiled
 * separately from the generated C that calls it (no cross-TU inlining in
 * the sanitizer lane's -O1, non-LTO build), THIS file's own compiler can
 * never prove that returned value is unused, so it can never
 * dead-store-eliminate the read -- exactly the "observable the optimizer
 * cannot eliminate" D-05-09 requires, independent of whether the caller
 * happens to use the value.
 *
 * Detection is ASan only (D-05-10): a plain `-O0`/`-O3` run of this
 * fixture is undefined behavior, not a guaranteed crash -- the freed
 * one-byte block is often re-served intact and the program can complete
 * cleanly. This file deliberately avoids any allocator-debug-perturbation
 * environment knob for determinism (D-05-10 names the two rejected knobs
 * explicitly): such knobs perturb freed bytes without trapping the
 * access, and differ across host allocator implementations.
 */
#include <stdlib.h>

/* g_retained_buffer: the static-stashed pointer FFI-003 is shaped around.
 * Module-private; never exposed to Lang, never touched by any other
 * translation unit. */
static unsigned char *g_retained_buffer = NULL;

/* g_retained_calls: the same "second call behaves differently" call-count
 * convention lang_foreign_nonlocal.c already establishes for this exact
 * shared-symbol shape. */
static int g_retained_calls = 0;

/* lang_retained_open: private helper for call 1 -- genuinely allocates a
 * one-byte buffer and stashes it in static storage. */
static unsigned char lang_retained_open(unsigned char argument) {
  unsigned char *block = malloc(1u);
  if (block == NULL) {
    return 0u;
  }
  block[0] = argument;
  g_retained_buffer = block;
  return argument;
}

/* lang_retained_stale: private helper for call 3 -- frees the SAME
 * stashed pointer and immediately reads through it again. The dangling
 * read below is the real heap-use-after-free this fixture exists to
 * exercise -- see the file header for why it cannot be optimized away. */
static unsigned char lang_retained_stale(void) {
  free(g_retained_buffer);
  return g_retained_buffer[0];
}

typedef struct _LANG_lang_retained_touch_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_retained_touch_result;

/* _LANG_lang_retained_touch: the SOLE symbol Lang declares and calls
 * (testdata/phase5/retained_pointer.lang calls it exactly three times).
 * Call 1 opens (allocate + stash); call 2 touches the still-live buffer;
 * call 3 goes stale (free, then dangling read). */
_LANG_lang_retained_touch_result _LANG_lang_retained_touch(unsigned char argument) {
  _LANG_lang_retained_touch_result result;
  g_retained_calls++;
  if (g_retained_calls == 1) {
    result.ok = 1;
    result.value = lang_retained_open(argument);
    return result;
  }
  if (g_retained_calls == 2) {
    result.ok = 1;
    result.value = (g_retained_buffer != NULL) ? g_retained_buffer[0] : 0u;
    return result;
  }
  result.ok = 1;
  result.value = lang_retained_stale();
  return result;
}
