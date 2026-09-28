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

typedef lang_file_byte_acquire_result lang_file_byte_acquire_fn(const char *path);
typedef lang_file_byte_use_result lang_file_byte_use_fn(lang_file_byte_owner owner);
typedef void lang_file_byte_release_fn(lang_file_byte_owner owner);

lang_file_byte_acquire_result lang_file_byte_acquire(const char *path);
lang_file_byte_use_result lang_file_byte_use(lang_file_byte_owner owner);
void lang_file_byte_release(lang_file_byte_owner owner);

#endif
