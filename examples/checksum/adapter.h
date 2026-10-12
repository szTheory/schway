#ifndef SCHWAY_CHECKSUM_FILE_BYTES_ADAPTER_H
#define SCHWAY_CHECKSUM_FILE_BYTES_ADAPTER_H

#include <stddef.h>
#include <stdint.h>

typedef struct schway_file_bytes_owner {
  unsigned char *data;
  uint64_t length;
} schway_file_bytes_owner;

typedef struct schway_file_bytes_acquire_result {
  int32_t status;
  schway_file_bytes_owner owner;
} schway_file_bytes_acquire_result;

typedef struct schway_file_bytes_view_result {
  const unsigned char *data;
  uint64_t length;
} schway_file_bytes_view_result;

enum {
  SCHWAY_FILE_BYTES_OK = 0,
  SCHWAY_FILE_BYTES_TOO_LARGE = 1,
  SCHWAY_FILE_BYTES_UNAVAILABLE = 2,
  SCHWAY_FILE_BYTES_ALLOC_FAILED = 3
};

_Static_assert(sizeof(void *) == 8u, "file bytes ABI requires 64-bit pointers");
_Static_assert(sizeof(schway_file_bytes_owner) == 16u, "unexpected owner size");
_Static_assert(_Alignof(schway_file_bytes_owner) == 8u, "unexpected owner alignment");
_Static_assert(offsetof(schway_file_bytes_owner, data) == 0u, "unexpected owner pointer offset");
_Static_assert(offsetof(schway_file_bytes_owner, length) == 8u, "unexpected owner length offset");
_Static_assert(sizeof(schway_file_bytes_acquire_result) == 24u, "unexpected acquire result size");
_Static_assert(_Alignof(schway_file_bytes_acquire_result) == 8u, "unexpected acquire result alignment");
_Static_assert(offsetof(schway_file_bytes_acquire_result, status) == 0u, "unexpected acquire status offset");
_Static_assert(offsetof(schway_file_bytes_acquire_result, owner) == 8u, "unexpected acquire owner offset");
_Static_assert(sizeof(schway_file_bytes_view_result) == 16u, "unexpected view result size");
_Static_assert(_Alignof(schway_file_bytes_view_result) == 8u, "unexpected view result alignment");
_Static_assert(offsetof(schway_file_bytes_view_result, data) == 0u, "unexpected view data offset");
_Static_assert(offsetof(schway_file_bytes_view_result, length) == 8u, "unexpected view length offset");

typedef schway_file_bytes_acquire_result schway_file_bytes_acquire_fn(const char *path);
typedef schway_file_bytes_view_result schway_file_bytes_view_fn(schway_file_bytes_owner owner);
typedef void schway_file_bytes_release_fn(schway_file_bytes_owner owner);

schway_file_bytes_acquire_result schway_file_bytes_acquire(const char *path);
schway_file_bytes_view_result schway_file_bytes_view(schway_file_bytes_owner owner);
void schway_file_bytes_release(schway_file_bytes_owner owner);

#endif
