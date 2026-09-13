/* payload_layout_mismatch.golden.c
 *
 * This file exists only to be refused (D-12-37). It declares
 * PayloadProbe_payload with its two payload alternative fields
 * (field_First, field_Second) TRANSPOSED relative to what
 * check.PayloadRecordLayout derives for the purpose-built PayloadProbe data
 * type (session.PayloadLayoutMutationRunner's own fixture): tag at offset 0,
 * field_First at offset 1, field_Second at offset 2. So compiling the
 * generated payload conformance unit (cgen.EmitPayloadConformance) against
 * this header, under the project's existing -Werror flag set, must fail its
 * offsetof _Static_assert pair.
 *
 * It is a seeded mutation, not a revert: no correct, non-transposed version
 * of this exact private header is ever committed, because
 * control:payload.layout_mismatch (session.PayloadLayoutMutationRunner)
 * attacks this frozen fixture directly rather than reverting a production
 * hunk (D-10, "the two mutation directions attack different artifacts").
 *
 * A PASSING compile of this fixture is a CONTROL FAILURE.
 *
 * IMPORTANT (D-12-37/D-12-38): this control polices only the struct
 * DECLARATION's sizeof/_Alignof/offsetof. It is structurally incapable of
 * proving which field a given match arm actually reads for a correct tag --
 * that is D-12-38's decisive control (TestPayloadSlotSwapMutationKilled), a
 * different artifact entirely.
 */
#ifndef LANG_PAYLOAD_LAYOUT_PROBE_PRIVATE_H
#define LANG_PAYLOAD_LAYOUT_PROBE_PRIVATE_H

typedef struct PayloadProbe_payload {
  unsigned char tag;
  unsigned char field_Second;
  unsigned char field_First;
} PayloadProbe_payload;

#endif /* LANG_PAYLOAD_LAYOUT_PROBE_PRIVATE_H */
