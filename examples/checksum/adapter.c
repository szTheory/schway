#ifndef _POSIX_C_SOURCE
#define _POSIX_C_SOURCE 200809L
#endif

#include "adapter.h"

#include <errno.h>
#include <fcntl.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>

#ifndef SCHWAY_FILE_BYTES_OPEN
#define SCHWAY_FILE_BYTES_OPEN open
#endif
#ifndef SCHWAY_FILE_BYTES_FSTAT
#define SCHWAY_FILE_BYTES_FSTAT fstat
#endif
#ifndef SCHWAY_FILE_BYTES_READ
#define SCHWAY_FILE_BYTES_READ read
#endif
#ifndef SCHWAY_FILE_BYTES_MALLOC
#define SCHWAY_FILE_BYTES_MALLOC malloc
#endif
#ifndef SCHWAY_FILE_BYTES_REALLOC
#define SCHWAY_FILE_BYTES_REALLOC realloc
#endif
#ifndef SCHWAY_FILE_BYTES_FREE
#define SCHWAY_FILE_BYTES_FREE free
#endif
#ifndef SCHWAY_FILE_BYTES_CLOSE
#define SCHWAY_FILE_BYTES_CLOSE close
#endif

#define SCHWAY_FILE_BYTES_LIMIT 65536u

/* The adapter is silent: it never writes to stdout or stderr. The generated
   program maps the status to FileError.<Alternative> (D-17). */

static schway_file_bytes_acquire_result acquire_failure(int32_t status) {
  schway_file_bytes_acquire_result result = {status, {NULL, UINT64_C(0)}};
  return result;
}

static schway_file_bytes_acquire_result acquire_cleanup_failure(
    int descriptor, unsigned char *data, int32_t status) {
  if (data != NULL) SCHWAY_FILE_BYTES_FREE(data);
  (void)SCHWAY_FILE_BYTES_CLOSE(descriptor);
  return acquire_failure(status);
}

schway_file_bytes_acquire_result schway_file_bytes_acquire(const char *path) {
  int descriptor = -1;
  struct stat details;
  unsigned char *data = NULL;
  unsigned char *shrunk = NULL;
  size_t received = 0u;
  schway_file_bytes_acquire_result result;

  if (path == NULL || path[0] == '\0') return acquire_failure(SCHWAY_FILE_BYTES_UNAVAILABLE);
  descriptor = SCHWAY_FILE_BYTES_OPEN(path, O_RDONLY | O_NONBLOCK);
  if (descriptor < 0) return acquire_failure(SCHWAY_FILE_BYTES_UNAVAILABLE);
  if (SCHWAY_FILE_BYTES_FSTAT(descriptor, &details) != 0 || !S_ISREG(details.st_mode)) {
    return acquire_cleanup_failure(descriptor, NULL, SCHWAY_FILE_BYTES_UNAVAILABLE);
  }
  data = (unsigned char *)SCHWAY_FILE_BYTES_MALLOC((size_t)SCHWAY_FILE_BYTES_LIMIT + 1u);
  if (data == NULL) {
    return acquire_cleanup_failure(descriptor, NULL, SCHWAY_FILE_BYTES_ALLOC_FAILED);
  }
  while (received < (size_t)SCHWAY_FILE_BYTES_LIMIT + 1u) {
    ssize_t amount = SCHWAY_FILE_BYTES_READ(
        descriptor, data + received, (size_t)SCHWAY_FILE_BYTES_LIMIT + 1u - received);
    if (amount > 0) {
      received += (size_t)amount;
      continue;
    }
    if (amount == 0) break;
    if (errno == EINTR) continue;
    return acquire_cleanup_failure(descriptor, data, SCHWAY_FILE_BYTES_UNAVAILABLE);
  }
  if (received > (size_t)SCHWAY_FILE_BYTES_LIMIT) {
    return acquire_cleanup_failure(descriptor, data, SCHWAY_FILE_BYTES_TOO_LARGE);
  }
  shrunk = (unsigned char *)SCHWAY_FILE_BYTES_REALLOC(data, received != 0u ? received : 1u);
  if (shrunk == NULL) {
    return acquire_cleanup_failure(descriptor, data, SCHWAY_FILE_BYTES_ALLOC_FAILED);
  }
  data = shrunk;
  if (SCHWAY_FILE_BYTES_CLOSE(descriptor) != 0) {
    SCHWAY_FILE_BYTES_FREE(data);
    return acquire_failure(SCHWAY_FILE_BYTES_UNAVAILABLE);
  }
  result.status = SCHWAY_FILE_BYTES_OK;
  result.owner.data = data;
  result.owner.length = (uint64_t)received;
  return result;
}

schway_file_bytes_view_result schway_file_bytes_view(schway_file_bytes_owner owner) {
  schway_file_bytes_view_result result;
  result.data = owner.data;
  result.length = owner.length;
  return result;
}

void schway_file_bytes_release(schway_file_bytes_owner owner) {
  SCHWAY_FILE_BYTES_FREE(owner.data);
}
