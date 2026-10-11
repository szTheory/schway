#ifndef SCHWAY_PHASE29_FIXTURE_BYTES_ADAPTER_H
#define SCHWAY_PHASE29_FIXTURE_BYTES_ADAPTER_H

#include <stddef.h>
#include <stdint.h>

typedef struct schway_fixture_bytes_owner {
  unsigned char *data;
  uint64_t length;
} schway_fixture_bytes_owner;

typedef struct schway_fixture_bytes_acquire_result {
  int32_t status;
  schway_fixture_bytes_owner owner;
} schway_fixture_bytes_acquire_result;

typedef struct schway_fixture_bytes_view_result {
  const unsigned char *data;
  uint64_t length;
} schway_fixture_bytes_view_result;

enum {
  SCHWAY_FIXTURE_BYTES_OK = 0,
  SCHWAY_FIXTURE_BYTES_UNKNOWN_CASE = 1,
  SCHWAY_FIXTURE_BYTES_ALLOCATION_FAILED = 2
};

_Static_assert(sizeof(void *) == 8u, "Phase 29 fixture bytes ABI requires 64-bit pointers");
_Static_assert(sizeof(schway_fixture_bytes_owner) == 16u, "unexpected owner size");
_Static_assert(_Alignof(schway_fixture_bytes_owner) == 8u, "unexpected owner alignment");
_Static_assert(offsetof(schway_fixture_bytes_owner, data) == 0u, "unexpected owner pointer offset");
_Static_assert(offsetof(schway_fixture_bytes_owner, length) == 8u, "unexpected owner length offset");
_Static_assert(sizeof(schway_fixture_bytes_acquire_result) == 24u, "unexpected acquire result size");
_Static_assert(_Alignof(schway_fixture_bytes_acquire_result) == 8u, "unexpected acquire result alignment");
_Static_assert(offsetof(schway_fixture_bytes_acquire_result, status) == 0u, "unexpected acquire status offset");
_Static_assert(offsetof(schway_fixture_bytes_acquire_result, owner) == 8u, "unexpected acquire owner offset");
_Static_assert(sizeof(schway_fixture_bytes_view_result) == 16u, "unexpected view result size");
_Static_assert(_Alignof(schway_fixture_bytes_view_result) == 8u, "unexpected view result alignment");
_Static_assert(offsetof(schway_fixture_bytes_view_result, data) == 0u, "unexpected view data offset");
_Static_assert(offsetof(schway_fixture_bytes_view_result, length) == 8u, "unexpected view length offset");

typedef schway_fixture_bytes_acquire_result schway_fixture_bytes_acquire_fn(uint64_t case_id);
typedef schway_fixture_bytes_view_result schway_fixture_bytes_view_fn(schway_fixture_bytes_owner owner);
typedef void schway_fixture_bytes_release_fn(schway_fixture_bytes_owner owner);

schway_fixture_bytes_acquire_result schway_fixture_bytes_acquire(uint64_t case_id);
schway_fixture_bytes_view_result schway_fixture_bytes_view(schway_fixture_bytes_owner owner);
void schway_fixture_bytes_release(schway_fixture_bytes_owner owner);

#endif
