/* lang_foreign_arena.c
 *
 * The hand-written, byte-frozen, THIRD separately-compiled foreign
 * translation unit for Phase 5's dynamic allocator-mismatch subject
 * (D-05-08). It peers lang_foreign_resource.c's structure exactly: the
 * Codename Lang compiler forms its whole contract with the symbol below
 * from Lang source alone -- it never parses this file's internals.
 * Quarantine is permanent and non-discharging, exactly as
 * lang_foreign_resource.c's own header records.
 *
 * This file is frozen from the commit that adds it, in the same class as
 * lang_foreign_resource.c and lang_foreign_nonlocal.c: it is a golden
 * fixture, not ordinary source under active development.
 *
 * D-05-08's genuine second allocator identity is lang_arena_open_impl /
 * lang_arena_close_impl below: posix_memalign paired with its own matching
 * free(), a real, different physical allocation path from
 * lang_foreign_resource.c's plain malloc. This correctly-paired impl pair
 * is deliberately kept unused by the exposed wrapper below -- it documents
 * the honest pairing this allocator identity is supposed to have, exactly
 * as lang_foreign_resource_private.h's own commented layout is present for
 * documentation even where a single plan narrows the on-the-wire result.
 *
 * The exposed `_LANG_lang_arena_open` wrapper is intentionally the
 * DYNAMIC-half adversarial fixture (D-05-08/D-05-09): it allocates via the
 * genuine posix_memalign identity above, then releases through the
 * mangled Itanium `operator delete(void*)` entry point instead of the
 * matching free() -- a real allocator-identity mismatch ASan's
 * alloc-dealloc-mismatch check exists to catch.
 *
 * WHY operator delete, not plain free() (a documented, verified deviation
 * from the plan's literal "routes its release through the plain-malloc
 * wrapper's free" wording): verified empirically on this host (Apple
 * clang 21, arm64) before authoring this file, `posix_memalign` allocated
 * memory freed via plain libc `free()` produces ZERO AddressSanitizer
 * report -- POSIX guarantees `free()` is the correct way to release
 * `posix_memalign` memory, and ASan's own allocator-family tracking
 * recognizes exactly three buckets (FROM_MALLOC, FROM_NEW, FROM_NEW_BR);
 * every libc allocation function (malloc, calloc, realloc,
 * posix_memalign, aligned_alloc, memalign, valloc) collapses into the SAME
 * FROM_MALLOC bucket and is therefore compatible with free() by design.
 * A genuinely ASan-recognized allocator-identity mismatch requires the
 * C++ operator-new family (FROM_NEW), reached here from plain C source by
 * declaring the Itanium-mangled `_Znwm`/`_ZdlPv` entry points `extern` and
 * linking against the C++ runtime (`-lc++`) -- this translation unit
 * itself stays pure C17 source; only these two runtime symbols are
 * borrowed, never a C++ compile. This is D-05-08's own flagged assumption
 * ("[CITED], not [VERIFIED]") falsified for posix_memalign+free and
 * replaced with a mechanism verified, on this host, to actually produce
 * `ERROR: AddressSanitizer: alloc-dealloc-mismatch`.
 */
#include <stdint.h>
#include <stdlib.h>

/* _Znwm / _ZdlPv are the Itanium C++ ABI's mangled names for
 * `operator new(size_t)` and `operator delete(void*)`, provided by the
 * C++ runtime (libc++/libc++abi) and callable directly from C via extern
 * declaration. Declared here, never defined -- the real symbols are
 * resolved at link time against the C++ runtime library. */
extern void *_Znwm(size_t size);
extern void _ZdlPv(void *pointer);

typedef struct lang_foreign_arena_block {
  unsigned char payload;
} lang_foreign_arena_block;

/* lang_arena_open_impl / lang_arena_close_impl: the CORRECTLY paired
 * private implementation. posix_memalign's own allocator identity, freed
 * via its own matching libc free() -- not itself defective, and NOT
 * called by the exposed wrapper below (see the file header's rationale).
 */
int lang_arena_open_impl(unsigned char request, uint64_t *out_handle) {
  void *block = NULL;
  if (posix_memalign(&block, sizeof(void *), sizeof(lang_foreign_arena_block)) != 0) {
    return 0;
  }
  ((lang_foreign_arena_block *)block)->payload = request;
  *out_handle = (uint64_t)(uintptr_t)block;
  return 1;
}

void lang_arena_close_impl(uint64_t handle) {
  free((void *)(uintptr_t)handle);
}

/* _LANG_lang_arena_open_result is the public, by-value ABI Codename Lang's
 * generated C declares an `extern` matching copy of (cgen.go's
 * emitLinearForeign), identical in shape to lang_foreign_resource.c's own
 * result record. */
typedef struct _LANG_lang_arena_open_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_arena_open_result;

_LANG_lang_arena_open_result _LANG_lang_arena_open(unsigned char argument) {
  _LANG_lang_arena_open_result result;
  void *block = NULL;
  if (posix_memalign(&block, sizeof(void *), sizeof(lang_foreign_arena_block)) != 0) {
    result.ok = 0;
    result.value = 0;
    return result;
  }
  ((lang_foreign_arena_block *)block)->payload = argument;
  result.ok = 1;
  result.value = argument;
  /* D-05-08's dynamic defect: this acquisition's real allocation went
   * through posix_memalign above (the FROM_MALLOC family), but the
   * release below deliberately routes through operator delete's mangled
   * entry point (the FROM_NEW family) -- a genuine allocator-identity
   * mismatch, reached through a real Lang program's declared foreign call
   * (testdata/phase5/allocator_mismatch.lang), never a hand-crafted
   * hostile C file unrelated to the boundary. */
  _ZdlPv(block);
  return result;
}
