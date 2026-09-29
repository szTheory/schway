#define _POSIX_C_SOURCE 200809L

#include "adapter.h"

#include <stdio.h>
#include <stdlib.h>

static FILE *receipt;
static void *allocation;
static int finalizer_registered;
static int allocation_live;
static int use_seen;
static int release_attempt_seen;
static int observer_rejected;

static void phase23_observer_finalize(void);

static void ensure_finalizer(void) {
  if (finalizer_registered) return;
  finalizer_registered = 1;
  if (atexit(phase23_observer_finalize) != 0) observer_rejected = 1;
}

static FILE *receipt_file(void) {
  const char *path;
  if (receipt != NULL) return receipt;
  path = getenv("SCHWAY_PHASE23_OBSERVER_PATH");
  if (path == NULL || path[0] == '\0') {
    observer_rejected = 1;
    return NULL;
  }
  receipt = fopen(path, "wb");
  if (receipt == NULL) observer_rejected = 1;
  return receipt;
}

static void record_event(const char *event, const char *detail) {
  FILE *stream;
  ensure_finalizer();
  stream = receipt_file();
  if (stream == NULL) return;
  if (fprintf(stream, "%s\t%s\n", event, detail) < 0 || fflush(stream) != 0) {
    observer_rejected = 1;
  }
}

static const char *pointer_identity(const void *pointer) {
  return pointer != NULL && pointer == allocation ? "p1" : "unknown";
}

void *phase23_observer_malloc(size_t size) {
  void *pointer;
  ensure_finalizer();
  pointer = malloc(size);
  if (pointer == NULL) return NULL;
  if (allocation != NULL) {
    observer_rejected = 1;
    record_event("attempt", "multiple-allocation");
    return pointer;
  }
  allocation = pointer;
  allocation_live = 1;
  record_event("malloc", pointer_identity(pointer));
  return pointer;
}

unsigned char phase23_observer_load_byte(const unsigned char *pointer) {
  unsigned char value;
  ensure_finalizer();
  if (!allocation_live || pointer == NULL || pointer != allocation || use_seen) {
    observer_rejected = 1;
    record_event("attempt", "invalid-use");
    return 0u;
  }
  value = pointer[0];
  use_seen = 1;
  record_event("use", pointer_identity(pointer));
  return value;
}

void phase23_observer_free(void *pointer) {
  ensure_finalizer();
  release_attempt_seen = 1;
  if (pointer == NULL || pointer != allocation) {
    observer_rejected = 1;
    record_event("attempt", "wrong-resource");
    return;
  }
  if (!allocation_live) {
    observer_rejected = 1;
    record_event("attempt", "duplicate-release");
    return;
  }
  if (!use_seen) {
    observer_rejected = 1;
    record_event("attempt", "premature-release");
    return;
  }
  record_event("free", pointer_identity(pointer));
  free(pointer);
  allocation_live = 0;
}

void phase23_observer_report_boundary(void) {
  ensure_finalizer();
  if (allocation_live || !use_seen) {
    observer_rejected = 1;
    record_event("attempt", "report-before-discharge");
  }
  record_event("report_boundary", "post-release");
}

static void phase23_observer_finalize(void) {
	if (allocation_live && use_seen && !release_attempt_seen) {
		record_event("attempt", "omitted-release");
	}
  int reject = observer_rejected || allocation_live || !use_seen;
  record_event("final", allocation_live ? "outstanding=1" : "outstanding=0");
  record_event("verdict", reject ? "reject" : "pass");
  if (receipt != NULL) {
    if (fclose(receipt) != 0) reject = 1;
    receipt = NULL;
  }
  if (reject) _Exit(86);
}
