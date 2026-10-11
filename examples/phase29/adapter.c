#include "adapter.h"

#include <stdlib.h>
#include <string.h>

/* Phase 29 fixture bytes generator: a test-only owner keyed by case id.
 * Each case is built here by literals or loops; no other generator is shared. */

static void fill_repeat(unsigned char *data, uint64_t count, unsigned char value) {
  uint64_t index;
  for (index = UINT64_C(0); index < count; index++) data[index] = value;
}

static int case_length(uint64_t case_id, uint64_t *length) {
  switch (case_id) {
    case 0: *length = UINT64_C(0); return 1;
    case 1: *length = UINT64_C(1); return 1;
    case 2: *length = UINT64_C(1); return 1;
    case 3: *length = UINT64_C(9); return 1;
    case 4: *length = UINT64_C(4096); return 1;
    case 5: *length = UINT64_C(5552); return 1;
    case 6: *length = UINT64_C(2); return 1;
    case 7: *length = UINT64_C(4096); return 1;
    case 8: *length = UINT64_C(16384); return 1;
    case 9: *length = UINT64_C(65536); return 1;
    default: return 0;
  }
}

static void case_fill(uint64_t case_id, unsigned char *data, uint64_t length) {
  switch (case_id) {
    case 0: break;
    case 1: data[0] = 0x00u; break;
    case 2: data[0] = (unsigned char)'a'; break;
    case 3: memcpy(data, "Wikipedia", 9u); break;
    case 4: fill_repeat(data, length, 0x00u); break;
    case 5: fill_repeat(data, length, 0xFFu); break;
    case 6: data[0] = 0x80u; data[1] = 0xFFu; break;
    case 7: fill_repeat(data, length, 0xFFu); break;
    case 8: fill_repeat(data, length, 0xFFu); break;
    case 9: fill_repeat(data, length, 0xFFu); break;
    default: break;
  }
}

schway_fixture_bytes_acquire_result schway_fixture_bytes_acquire(uint64_t case_id) {
  schway_fixture_bytes_acquire_result result = {SCHWAY_FIXTURE_BYTES_UNKNOWN_CASE, {NULL, UINT64_C(0)}};
  uint64_t length = UINT64_C(0);
  size_t allocation;
  unsigned char *data;
  if (!case_length(case_id, &length)) return result;
  allocation = length == UINT64_C(0) ? (size_t)1u : (size_t)length;
  data = (unsigned char *)malloc(allocation);
  if (data == NULL) {
    result.status = SCHWAY_FIXTURE_BYTES_ALLOCATION_FAILED;
    return result;
  }
  case_fill(case_id, data, length);
  result.status = SCHWAY_FIXTURE_BYTES_OK;
  result.owner.data = data;
  result.owner.length = length;
  return result;
}

schway_fixture_bytes_view_result schway_fixture_bytes_view(schway_fixture_bytes_owner owner) {
  schway_fixture_bytes_view_result result;
  result.data = owner.data;
  result.length = owner.length;
  return result;
}

void schway_fixture_bytes_release(schway_fixture_bytes_owner owner) {
  free(owner.data);
}
