#ifndef LANG_PHASE23_FILE_BYTE_ADAPTER_H
#define LANG_PHASE23_FILE_BYTE_ADAPTER_H

#include <stddef.h>
#include <stdint.h>

typedef struct lang_file_byte_owner {
  unsigned char *data;
  uint64_t length;
} lang_file_byte_owner;

typedef struct lang_file_byte_acquire_result {
  int32_t status;
  lang_file_byte_owner owner;
} lang_file_byte_acquire_result;

typedef struct lang_file_byte_use_result {
  int32_t status;
  uint64_t value;
} lang_file_byte_use_result;

enum {
  LANG_FILE_BYTE_USE_OK = 0,
  LANG_FILE_BYTE_USE_UNSUPPORTED = 1,
  LANG_FILE_BYTE_USE_INVALID_OWNER = 2
};

_Static_assert(sizeof(void *) == 8u, "Phase 23 file-byte ABI requires 64-bit pointers");
_Static_assert(sizeof(lang_file_byte_owner) == 16u, "unexpected owner size");
_Static_assert(_Alignof(lang_file_byte_owner) == 8u, "unexpected owner alignment");
_Static_assert(offsetof(lang_file_byte_owner, data) == 0u, "unexpected owner pointer offset");
_Static_assert(offsetof(lang_file_byte_owner, length) == 8u, "unexpected owner length offset");
_Static_assert(sizeof(lang_file_byte_acquire_result) == 24u, "unexpected acquire result size");
_Static_assert(_Alignof(lang_file_byte_acquire_result) == 8u, "unexpected acquire result alignment");
_Static_assert(offsetof(lang_file_byte_acquire_result, status) == 0u, "unexpected acquire status offset");
_Static_assert(offsetof(lang_file_byte_acquire_result, owner) == 8u, "unexpected acquire owner offset");
_Static_assert(sizeof(lang_file_byte_use_result) == 16u, "unexpected use result size");
_Static_assert(_Alignof(lang_file_byte_use_result) == 8u, "unexpected use result alignment");
_Static_assert(offsetof(lang_file_byte_use_result, status) == 0u, "unexpected use status offset");
_Static_assert(offsetof(lang_file_byte_use_result, value) == 8u, "unexpected use value offset");

typedef lang_file_byte_acquire_result lang_file_byte_acquire_fn(const char *path);
typedef lang_file_byte_use_result lang_file_byte_use_fn(lang_file_byte_owner owner);
typedef void lang_file_byte_release_fn(lang_file_byte_owner owner);

lang_file_byte_acquire_result lang_file_byte_acquire(const char *path);
lang_file_byte_use_result lang_file_byte_use(lang_file_byte_owner owner);
void lang_file_byte_release(lang_file_byte_owner owner);

#endif
