#ifndef _POSIX_C_SOURCE
#define _POSIX_C_SOURCE 200809L
#endif

#include "adapter.h"

#include <errno.h>
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <unistd.h>

#ifndef SCHWAY_FILE_BYTE_OPEN
#define SCHWAY_FILE_BYTE_OPEN open
#endif
#ifndef SCHWAY_FILE_BYTE_FSTAT
#define SCHWAY_FILE_BYTE_FSTAT fstat
#endif
#ifndef SCHWAY_FILE_BYTE_MALLOC
#define SCHWAY_FILE_BYTE_MALLOC malloc
#endif
#ifndef SCHWAY_FILE_BYTE_FREE
#define SCHWAY_FILE_BYTE_FREE free
#endif
#ifndef SCHWAY_FILE_BYTE_LOAD_BYTE
#define SCHWAY_FILE_BYTE_LOAD_BYTE(pointer) (*(pointer))
#endif
#ifndef SCHWAY_FILE_BYTE_READ
#define SCHWAY_FILE_BYTE_READ read
#endif
#ifndef SCHWAY_FILE_BYTE_CLOSE
#define SCHWAY_FILE_BYTE_CLOSE close
#endif

enum {
  SCHWAY_FILE_BYTE_OK = 0,
  SCHWAY_FILE_BYTE_EMPTY = 1,
  SCHWAY_FILE_BYTE_TOO_LONG = 2,
  SCHWAY_FILE_BYTE_OPEN_FAILED = 3,
  SCHWAY_FILE_BYTE_NOT_REGULAR = 4,
  SCHWAY_FILE_BYTE_READ_FAILED = 5,
  SCHWAY_FILE_BYTE_ALLOC_FAILED = 6,
  SCHWAY_FILE_BYTE_CLOSE_FAILED = 7,
  SCHWAY_FILE_BYTE_INVALID_PATH = 8
};

static const char *acquire_status_name(int32_t status) {
  switch (status) {
    case SCHWAY_FILE_BYTE_EMPTY: return "EmptyFile";
    case SCHWAY_FILE_BYTE_TOO_LONG: return "FileTooLong";
    case SCHWAY_FILE_BYTE_OPEN_FAILED: return "OpenFailed";
    case SCHWAY_FILE_BYTE_NOT_REGULAR: return "NotRegular";
    case SCHWAY_FILE_BYTE_READ_FAILED: return "ReadFailed";
    case SCHWAY_FILE_BYTE_ALLOC_FAILED: return "AllocFailed";
    case SCHWAY_FILE_BYTE_CLOSE_FAILED: return "CloseFailed";
    case SCHWAY_FILE_BYTE_INVALID_PATH: return "InvalidPath";
    default: return "AcquireFailed";
  }
}

static schway_file_byte_acquire_result acquire_failure(int32_t status) {
  schway_file_byte_acquire_result result = {status, {NULL, UINT64_C(0)}};
  fputs("schway_file_byte_acquire: ", stderr);
  fputs(acquire_status_name(status), stderr);
  fputc('\n', stderr);
  return result;
}

static schway_file_byte_acquire_result acquire_cleanup_failure(
    int descriptor, unsigned char *data, int32_t primary_status) {
  int32_t status = primary_status;
  if (data != NULL) SCHWAY_FILE_BYTE_FREE(data);
  if (SCHWAY_FILE_BYTE_CLOSE(descriptor) != 0) {
    status = SCHWAY_FILE_BYTE_CLOSE_FAILED;
  }
  return acquire_failure(status);
}

schway_file_byte_acquire_result schway_file_byte_acquire(const char *path) {
  int descriptor = -1;
  int close_error = 0;
  struct stat details;
  unsigned char *data = NULL;
  size_t received = 0u;
  unsigned char extra;
  schway_file_byte_acquire_result result;

  if (path == NULL || path[0] == '\0') return acquire_failure(SCHWAY_FILE_BYTE_INVALID_PATH);
  descriptor = SCHWAY_FILE_BYTE_OPEN(path, O_RDONLY | O_NONBLOCK);
  if (descriptor < 0) return acquire_failure(SCHWAY_FILE_BYTE_OPEN_FAILED);
  if (SCHWAY_FILE_BYTE_FSTAT(descriptor, &details) != 0 || !S_ISREG(details.st_mode)) {
    return acquire_cleanup_failure(descriptor, NULL, SCHWAY_FILE_BYTE_NOT_REGULAR);
  }
  data = (unsigned char *)SCHWAY_FILE_BYTE_MALLOC(1u);
  if (data == NULL) {
    return acquire_cleanup_failure(descriptor, NULL, SCHWAY_FILE_BYTE_ALLOC_FAILED);
  }
  while (received < 1u) {
    ssize_t amount = SCHWAY_FILE_BYTE_READ(descriptor, data + received, 1u - received);
    if (amount > 0) {
      received += (size_t)amount;
      continue;
    }
    if (amount == 0) {
      return acquire_cleanup_failure(descriptor, data, SCHWAY_FILE_BYTE_EMPTY);
    }
    if (errno == EINTR) continue;
    return acquire_cleanup_failure(descriptor, data, SCHWAY_FILE_BYTE_READ_FAILED);
  }
  for (;;) {
    ssize_t amount = SCHWAY_FILE_BYTE_READ(descriptor, &extra, 1u);
    if (amount == 0) break;
    if (amount > 0) {
      return acquire_cleanup_failure(descriptor, data, SCHWAY_FILE_BYTE_TOO_LONG);
    }
    if (errno == EINTR) continue;
    return acquire_cleanup_failure(descriptor, data, SCHWAY_FILE_BYTE_READ_FAILED);
  }
  if (SCHWAY_FILE_BYTE_CLOSE(descriptor) != 0) close_error = 1;
  if (close_error) {
    SCHWAY_FILE_BYTE_FREE(data);
    return acquire_failure(SCHWAY_FILE_BYTE_CLOSE_FAILED);
  }
  result.status = SCHWAY_FILE_BYTE_OK;
  result.owner.data = data;
  result.owner.length = UINT64_C(1);
  return result;
}

schway_file_byte_use_result schway_file_byte_use(schway_file_byte_owner owner) {
  unsigned char value;
  schway_file_byte_use_result result = {SCHWAY_FILE_BYTE_USE_INVALID_OWNER, UINT64_C(0)};
  if (owner.data == NULL || owner.length != UINT64_C(1)) return result;
  value = SCHWAY_FILE_BYTE_LOAD_BYTE(owner.data);
  if (value != 0x41u && value != 0x42u) {
    result.status = SCHWAY_FILE_BYTE_USE_UNSUPPORTED;
    return result;
  }
  result.status = SCHWAY_FILE_BYTE_USE_OK;
  result.value = (uint64_t)value;
  return result;
}

void schway_file_byte_release(schway_file_byte_owner owner) {
  SCHWAY_FILE_BYTE_FREE(owner.data);
}
