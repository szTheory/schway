/* schway_foreign_resource.c
 *
 * The hand-written, byte-frozen, separately-compiled foreign translation
 * unit for Phase 4's single audited C boundary (D-04-10). It wraps real
 * libc malloc/free and returns a small record by value. The Schway
 * compiler forms its whole contract with the symbol below from Lang source
 * alone -- it never parses schway_foreign_resource_private.h. Quarantine is
 * permanent and non-discharging: a green Phase 4 gate never proves this
 * file obeys the contract Lang declares for it, only that the declared
 * obligations are inspectable (04-RESEARCH.md, Pattern 4).
 *
 * This file is frozen from the commit that adds it, in the same class as
 * testdata/phase2/owned_transfer.golden.c: it is a golden fixture, not
 * ordinary source under active development.
 */
#include <stdlib.h>

#include "schway_foreign_resource_private.h"

int schway_foreign_resource_open_impl(unsigned char request, uint64_t *out_handle) {
  schway_foreign_resource_block *block = malloc(sizeof(schway_foreign_resource_block));
  if (block == NULL) {
    return 0;
  }
  block->payload = request;
  *out_handle = (uint64_t)(uintptr_t)block;
  return 1;
}

void schway_foreign_resource_release_impl(uint64_t handle) {
  free((void *)(uintptr_t)handle);
}

/* _SCHWAY_schway_res_open_result is the public, by-value ABI Schway's
 * generated C declares an `extern` matching copy of (cgen.go's
 * emitLinearForeign): a success flag plus one byte of value. This plan
 * narrows the acquired resource to its single request byte on the wire --
 * a fuller opaque handle is a later plan's concern (04-CONTEXT.md, FFI-01
 * PARTIAL) -- but the acquisition itself still goes through real libc
 * malloc/free via the private impl above, so a genuine allocation failure
 * (out of memory) is a real, observable `ok == 0` this boundary can
 * produce, not a value manufactured for the demo.
 */
typedef struct _SCHWAY_schway_res_open_result {
  unsigned char ok;
  unsigned char value;
} _SCHWAY_schway_res_open_result;

_SCHWAY_schway_res_open_result _SCHWAY_schway_res_open(unsigned char argument) {
  _SCHWAY_schway_res_open_result result;
  uint64_t handle = 0;
  if (!schway_foreign_resource_open_impl(argument, &handle)) {
    result.ok = 0;
    result.value = 0;
    return result;
  }
  result.ok = 1;
  result.value = argument;
  schway_foreign_resource_release_impl(handle);
  return result;
}
