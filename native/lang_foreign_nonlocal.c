/* lang_foreign_nonlocal.c
 *
 * The hand-written, byte-frozen, second foreign translation unit for
 * Phase 4's process-root setjmp/longjmp landing pad (D-04-17/D-04-18). Its
 * single declared symbol behaves differently by call count: the FIRST call
 * acquires a real one-byte block via libc malloc -- exactly like
 * lang_foreign_resource.c's own acquisition -- and returns normally, so the
 * resource genuinely becomes live before anything else happens. The SECOND
 * call performs a real C nonlocal exit: it longjmps back into Lang's
 * generated process-root landing pad instead of ever returning to its
 * caller. This is what gives the pad a reachable witness through the real
 * generated pipeline (checked, lowered, compiled, linked, executed) rather
 * than a hand-written C harness.
 *
 * The call-count convention below ("second call jumps") is a documented,
 * shared convention with the interpreter's own model of this exact fixture
 * (see interp.go's runLinearBlocks) -- the interpreter cannot actually call
 * C, so both engines are kept in agreement on WHEN the exit happens by
 * sharing this convention rather than by any real control transfer.
 *
 * Per D-04-10, the Codename Lang compiler forms its whole contract with this
 * symbol from Lang source alone -- it never parses any private header this
 * file might declare (it declares none). Quarantine is permanent and
 * non-discharging: a green Phase 4 gate never proves this file obeys the
 * contract Lang declares for it, only that the declared obligations are
 * inspectable (04-RESEARCH.md, Pattern 4).
 *
 * This file is frozen from the commit that adds it, in the same class as
 * lang_foreign_resource.c and testdata/phase2/owned_transfer.golden.c: it is
 * a golden fixture, not ordinary source under active development.
 */
#include <setjmp.h>
#include <stdlib.h>

/* lang_nonlocal_landing is declared, never defined, here: Lang's own
 * generated C defines it and installs the landing point via setjmp() at the
 * top of main(), exactly once per process (D-04-17). This translation unit
 * only ever longjmps into it -- it never calls setjmp itself, and it is the
 * ONLY function in the whole program permitted to call longjmp on it. */
extern jmp_buf lang_nonlocal_landing;

/* _LANG_lang_nonlocal_probe_result is the public, by-value ABI Codename
 * Lang's generated C declares an `extern` matching copy of
 * (cgen.go's emitLinearForeign), identical in shape to
 * lang_foreign_resource.c's own result record. */
typedef struct _LANG_lang_nonlocal_probe_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_nonlocal_probe_result;

static int lang_nonlocal_probe_calls = 0;

_LANG_lang_nonlocal_probe_result _LANG_lang_nonlocal_probe(unsigned char argument) {
  _LANG_lang_nonlocal_probe_result result;
  lang_nonlocal_probe_calls++;
  if (lang_nonlocal_probe_calls == 1) {
    unsigned char *block = malloc(1u);
    if (block == NULL) {
      result.ok = 0;
      result.value = 0;
      return result;
    }
    block[0] = argument;
    free(block);
    result.ok = 1;
    result.value = argument;
    return result;
  }
  /* D-04-18: this call never returns to its caller -- it transfers control
   * back to the process-root landing pad Lang's generated C installed via
   * setjmp(), bypassing every ordinary release call the acquisition above
   * would otherwise reach. No cleanup runs here or anywhere below this
   * call; the pad on the other end of this jump must not run one either. */
  longjmp(lang_nonlocal_landing, 1);
  result.ok = 0; /* unreachable: longjmp never returns */
  result.value = 0;
  return result;
}
