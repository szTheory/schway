/* schway_foreign_resource_private.h
 *
 * Private layout for the byte-frozen foreign translation unit
 * (schway_foreign_resource.c). Per D-04-10, the Schway compiler forms
 * its whole contract with this symbol from Lang source alone and never
 * opens this header -- it is reachable only from schway_foreign_resource.c
 * itself and, in a later plan, the generated conformance translation unit
 * that is the single explicit place Lang's declaration and this private
 * layout are permitted to meet (D-04-11). Quarantine here is an
 * information-flow property, not an authorship one: this file is
 * hand-written and lives in-repo, but no generated C ever includes it.
 *
 * This phase's frozen record: a single-byte payload block allocated via
 * libc malloc, its address encoded as a 64-bit handle for a future plan's
 * fuller layout-inspection needs, even though this plan's own generated
 * call site narrows the on-the-wire result to a one-byte value (see
 * schway_foreign_resource.c's public _SCHWAY_schway_res_open_result).
 */
#ifndef SCHWAY_FOREIGN_RESOURCE_PRIVATE_H
#define SCHWAY_FOREIGN_RESOURCE_PRIVATE_H

#include <stdint.h>

typedef struct schway_foreign_resource_block {
  unsigned char payload;
} schway_foreign_resource_block;

/* schway_foreign_resource_open_impl allocates one block via libc malloc,
 * storing request in it, and reports success/failure plus the block's
 * address as a handle. Returns 1 on success, 0 on allocation failure. */
int schway_foreign_resource_open_impl(unsigned char request, uint64_t *out_handle);

/* schway_foreign_resource_release_impl frees a handle previously returned by
 * schway_foreign_resource_open_impl. */
void schway_foreign_resource_release_impl(uint64_t handle);

#endif /* SCHWAY_FOREIGN_RESOURCE_PRIVATE_H */
