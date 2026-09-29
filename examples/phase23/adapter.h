#ifndef SCHWAY_PHASE23_FILE_BYTE_ADAPTER_H
#define SCHWAY_PHASE23_FILE_BYTE_ADAPTER_H

#include <stddef.h>
#include <stdint.h>

typedef struct schway_file_byte_owner {
  unsigned char *data;
  uint64_t length;
} schway_file_byte_owner;

typedef struct schway_file_byte_acquire_result {
  int32_t status;
  schway_file_byte_owner owner;
} schway_file_byte_acquire_result;

typedef struct schway_file_byte_use_result {
  int32_t status;
  uint64_t value;
} schway_file_byte_use_result;

enum {
  SCHWAY_FILE_BYTE_USE_OK = 0,
  SCHWAY_FILE_BYTE_USE_UNSUPPORTED = 1,
  SCHWAY_FILE_BYTE_USE_INVALID_OWNER = 2
};

_Static_assert(sizeof(void *) == 8u, "Phase 23 file-byte ABI requires 64-bit pointers");
_Static_assert(sizeof(schway_file_byte_owner) == 16u, "unexpected owner size");
_Static_assert(_Alignof(schway_file_byte_owner) == 8u, "unexpected owner alignment");
_Static_assert(offsetof(schway_file_byte_owner, data) == 0u, "unexpected owner pointer offset");
_Static_assert(offsetof(schway_file_byte_owner, length) == 8u, "unexpected owner length offset");
_Static_assert(sizeof(schway_file_byte_acquire_result) == 24u, "unexpected acquire result size");
_Static_assert(_Alignof(schway_file_byte_acquire_result) == 8u, "unexpected acquire result alignment");
_Static_assert(offsetof(schway_file_byte_acquire_result, status) == 0u, "unexpected acquire status offset");
_Static_assert(offsetof(schway_file_byte_acquire_result, owner) == 8u, "unexpected acquire owner offset");
_Static_assert(sizeof(schway_file_byte_use_result) == 16u, "unexpected use result size");
_Static_assert(_Alignof(schway_file_byte_use_result) == 8u, "unexpected use result alignment");
_Static_assert(offsetof(schway_file_byte_use_result, status) == 0u, "unexpected use status offset");
_Static_assert(offsetof(schway_file_byte_use_result, value) == 8u, "unexpected use value offset");

typedef schway_file_byte_acquire_result schway_file_byte_acquire_fn(const char *path);
typedef schway_file_byte_use_result schway_file_byte_use_fn(schway_file_byte_owner owner);
typedef void schway_file_byte_release_fn(schway_file_byte_owner owner);

schway_file_byte_acquire_result schway_file_byte_acquire(const char *path);
schway_file_byte_use_result schway_file_byte_use(schway_file_byte_owner owner);
void schway_file_byte_release(schway_file_byte_owner owner);

#endif
