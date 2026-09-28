#define _POSIX_C_SOURCE 200809L

#include "adapter.h"

#include <errno.h>
#include <fcntl.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <unistd.h>

enum {
  LANG_FILE_BYTE_OK = 0,
  LANG_FILE_BYTE_EMPTY = 1,
  LANG_FILE_BYTE_TOO_LONG = 2,
  LANG_FILE_BYTE_OPEN_FAILED = 3,
  LANG_FILE_BYTE_NOT_REGULAR = 4,
  LANG_FILE_BYTE_READ_FAILED = 5,
  LANG_FILE_BYTE_ALLOC_FAILED = 6,
  LANG_FILE_BYTE_CLOSE_FAILED = 7,
  LANG_FILE_BYTE_INVALID_PATH = 8,
  LANG_FILE_BYTE_UNSUPPORTED = 1,
  LANG_FILE_BYTE_INVALID_OWNER = 2
};

static lang_file_byte_acquire_result acquire_failure(int32_t status) {
  lang_file_byte_acquire_result result = {status, {NULL, UINT64_C(0)}};
  return result;
}

lang_file_byte_acquire_result lang_file_byte_acquire(const char *path) {
  int descriptor = -1;
  int close_error = 0;
  struct stat details;
  unsigned char *data = NULL;
  size_t received = 0u;
  unsigned char extra;
  lang_file_byte_acquire_result result;

  if (path == NULL || path[0] == '\0') return acquire_failure(LANG_FILE_BYTE_INVALID_PATH);
  descriptor = open(path, O_RDONLY | O_NONBLOCK);
  if (descriptor < 0) return acquire_failure(LANG_FILE_BYTE_OPEN_FAILED);
  if (fstat(descriptor, &details) != 0 || !S_ISREG(details.st_mode)) {
    (void)close(descriptor);
    return acquire_failure(LANG_FILE_BYTE_NOT_REGULAR);
  }
  data = (unsigned char *)malloc(1u);
  if (data == NULL) {
    (void)close(descriptor);
    return acquire_failure(LANG_FILE_BYTE_ALLOC_FAILED);
  }
  while (received < 1u) {
    ssize_t amount = read(descriptor, data + received, 1u - received);
    if (amount > 0) {
      received += (size_t)amount;
      continue;
    }
    if (amount == 0) {
      free(data);
      (void)close(descriptor);
      return acquire_failure(LANG_FILE_BYTE_EMPTY);
    }
    if (errno == EINTR) continue;
    free(data);
    (void)close(descriptor);
    return acquire_failure(LANG_FILE_BYTE_READ_FAILED);
  }
  for (;;) {
    ssize_t amount = read(descriptor, &extra, 1u);
    if (amount == 0) break;
    if (amount > 0) {
      free(data);
      (void)close(descriptor);
      return acquire_failure(LANG_FILE_BYTE_TOO_LONG);
    }
    if (errno == EINTR) continue;
    free(data);
    (void)close(descriptor);
    return acquire_failure(LANG_FILE_BYTE_READ_FAILED);
  }
  if (close(descriptor) != 0) close_error = 1;
  if (close_error) {
    free(data);
    return acquire_failure(LANG_FILE_BYTE_CLOSE_FAILED);
  }
  result.status = LANG_FILE_BYTE_OK;
  result.owner.data = data;
  result.owner.length = UINT64_C(1);
  return result;
}

lang_file_byte_use_result lang_file_byte_use(lang_file_byte_owner owner) {
  lang_file_byte_use_result result = {LANG_FILE_BYTE_INVALID_OWNER, UINT64_C(0)};
  if (owner.data == NULL || owner.length != UINT64_C(1)) return result;
  if (owner.data[0] != 0x41u && owner.data[0] != 0x42u) {
    result.status = LANG_FILE_BYTE_UNSUPPORTED;
    return result;
  }
  result.status = LANG_FILE_BYTE_OK;
  result.value = (uint64_t)owner.data[0];
  return result;
}

void lang_file_byte_release(lang_file_byte_owner owner) {
  free(owner.data);
}
