#define _POSIX_C_SOURCE 200809L

#include <errno.h>
#include <fcntl.h>
#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/stat.h>
#include <sys/types.h>
#include <unistd.h>

/* Phase 31 counting observer for the checksum file-bytes adapter (D-10).
   Fault mode comes from SCHWAY_PHASE31_FAULT; the receipt line is written at
   exit to SCHWAY_PHASE31_RECEIPT. */

#define PHASE31_TABLE 64u

static void *live[PHASE31_TABLE];
static unsigned long allocations;
static unsigned long frees;
static unsigned long doubles;
static unsigned long bytes_read;
static unsigned long read_calls;
static int registered;
static int parsed;
static const char *fault = "none";
static char saved_path[4096];
static int grown;

static void write_receipt(void) {
  const char *path = getenv("SCHWAY_PHASE31_RECEIPT");
  FILE *file;
  if (path == NULL || path[0] == '\0') return;
  file = fopen(path, "wb");
  if (file == NULL) return;
  (void)fprintf(file, "allocations=%lu frees=%lu doubles=%lu bytes_read=%lu\n", allocations, frees, doubles, bytes_read);
  (void)fclose(file);
}

static void ensure(void) {
  const char *value;
  if (!registered) {
    registered = 1;
    (void)atexit(write_receipt);
  }
  if (!parsed) {
    parsed = 1;
    value = getenv("SCHWAY_PHASE31_FAULT");
    if (value != NULL && value[0] != '\0') fault = value;
  }
}

static int is(const char *mode) {
  ensure();
  return strcmp(fault, mode) == 0;
}

static void track(void *pointer) {
  size_t index;
  for (index = 0u; index < PHASE31_TABLE; index++) {
    if (live[index] == NULL) {
      live[index] = pointer;
      return;
    }
  }
}

static int untrack(void *pointer) {
  size_t index;
  for (index = 0u; index < PHASE31_TABLE; index++) {
    if (live[index] == pointer) {
      live[index] = NULL;
      return 1;
    }
  }
  return 0;
}

int phase31_observer_open(const char *path, int flags) {
  ensure();
  if (path != NULL) {
    (void)strncpy(saved_path, path, sizeof saved_path - 1u);
    saved_path[sizeof saved_path - 1u] = '\0';
  }
  return open(path, flags);
}

int phase31_observer_fstat(int descriptor, struct stat *details) {
  if (is("fstat-fail")) {
    errno = EIO;
    return -1;
  }
  return fstat(descriptor, details);
}

ssize_t phase31_observer_read(int descriptor, void *buffer, size_t count) {
  ssize_t amount;
  int first;
  ensure();
  first = read_calls == 0ul;
  read_calls++;
  if (first && is("read-eio")) {
    errno = EIO;
    return -1;
  }
  if (first && is("read-eintr-once")) {
    errno = EINTR;
    return -1;
  }
  if (is("read-one-byte") && count > 1u) count = 1u;
  amount = read(descriptor, buffer, count);
  if (amount > 0) bytes_read += (unsigned long)amount;
  if (is("grow-after-first-read") && first && !grown && saved_path[0] != '\0') {
    int out = open(saved_path, O_WRONLY | O_APPEND);
    grown = 1;
    if (out >= 0) {
      static const unsigned char extra[16] = {0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
                                              0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF};
      (void)write(out, extra, sizeof extra);
      (void)close(out);
    }
  }
  return amount;
}

void *phase31_observer_malloc(size_t size) {
  void *pointer;
  if (is("malloc-null")) return NULL;
  pointer = malloc(size);
  if (pointer != NULL) {
    allocations++;
    track(pointer);
  }
  return pointer;
}

void *phase31_observer_realloc(void *pointer, size_t size) {
  void *moved;
  if (is("realloc-null")) return NULL;
  moved = realloc(pointer, size);
  if (moved != NULL && moved != pointer) {
    (void)untrack(pointer);
    track(moved);
  }
  return moved;
}

void phase31_observer_free(void *pointer) {
  ensure();
  if (pointer == NULL) return;
  if (!untrack(pointer)) {
    doubles++;
    return;
  }
  frees++;
  free(pointer);
}

int phase31_observer_close(int descriptor) {
  int status = close(descriptor);
  if (is("close-fail")) {
    errno = EIO;
    return -1;
  }
  return status;
}
