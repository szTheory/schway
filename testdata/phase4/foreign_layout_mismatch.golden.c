/* foreign_layout_mismatch.golden.c
 *
 * This file exists only to be refused (D-04-11/D-11). It declares
 * schway_foreign_layout_probe_block with its two fields transposed relative
 * to what the layout mutation control's own Lang-side contract declares
 * (`first` at offset 0, `second` at offset 1) -- so compiling the generated
 * conformance unit against this header, under the project's existing
 * `-Werror`, must fail its offsetof _Static_assert pair. It is a seeded
 * mutation, not a revert: no correct, non-transposed version of this exact
 * private header is ever committed, because control:foreign.layout_mismatch
 * (session.LayoutMutationRunner) attacks this frozen fixture directly rather
 * than reverting a production hunk (D-10, "the two mutation directions
 * attack different artifacts").
 */
#ifndef SCHWAY_FOREIGN_LAYOUT_PROBE_PRIVATE_H
#define SCHWAY_FOREIGN_LAYOUT_PROBE_PRIVATE_H

typedef struct schway_foreign_layout_probe_block {
  unsigned char second;
  unsigned char first;
} schway_foreign_layout_probe_block;

#endif /* SCHWAY_FOREIGN_LAYOUT_PROBE_PRIVATE_H */
