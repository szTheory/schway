#define _POSIX_C_SOURCE 200809L

#include "adapter.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

enum { PHASE24_OWNER_COUNT = 3, PHASE24_ID_CAPACITY = 512 };

typedef struct phase24_owner_record {
  unsigned char *pointer;
  char semantic_id[PHASE24_ID_CAPACITY];
  int live;
} phase24_owner_record;

static FILE *phase24_receipt;
static phase24_owner_record phase24_owners[PHASE24_OWNER_COUNT];
static size_t phase24_allocation_count;
static size_t phase24_release_count;
static int phase24_use_seen;
static int phase24_finalizer_registered;
static int phase24_rejected;

static void phase24_observer_finalize(void);

static void phase24_ensure_finalizer(void) {
  if (phase24_finalizer_registered) return;
  phase24_finalizer_registered = 1;
  if (atexit(phase24_observer_finalize) != 0) phase24_rejected = 1;
}

static FILE *phase24_receipt_file(void) {
  const char *path;
  if (phase24_receipt != NULL) return phase24_receipt;
  path = getenv("SCHWAY_PHASE24_OBSERVER_PATH");
  if (path == NULL || path[0] == '\0') {
    phase24_rejected = 1;
    return NULL;
  }
  phase24_receipt = fopen(path, "wb");
  if (phase24_receipt == NULL) phase24_rejected = 1;
  return phase24_receipt;
}

static void phase24_record(const char *event, const char *detail) {
  FILE *stream;
  phase24_ensure_finalizer();
  stream = phase24_receipt_file();
  if (stream == NULL) return;
  if (fprintf(stream, "%s\t%s\n", event, detail) < 0 || fflush(stream) != 0) {
    phase24_rejected = 1;
  }
}

static int phase24_find_owner(const void *pointer) {
  size_t index;
  for (index = 0u; index < phase24_allocation_count && index < PHASE24_OWNER_COUNT; index++) {
    if (phase24_owners[index].pointer == pointer) return (int)index;
  }
  return -1;
}

static size_t phase24_outstanding(void) {
  size_t index;
  size_t count = 0u;
  for (index = 0u; index < phase24_allocation_count && index < PHASE24_OWNER_COUNT; index++) {
    if (phase24_owners[index].live) count++;
  }
  return count;
}

static int phase24_semantic_identity(size_t index, char *buffer, size_t capacity) {
  const char *operation = getenv("SCHWAY_PHASE24_ACQUIRE_OPERATION_ID");
  const char *collision = getenv("SCHWAY_PHASE24_OBSERVER_COLLIDE");
  size_t activation = index + 1u;
  int written;
  if (operation == NULL || operation[0] == '\0' || index >= PHASE24_OWNER_COUNT) return 0;
  if (collision != NULL && strcmp(collision, "2") == 0 && index == 1u) {
    activation = 1u;
    phase24_rejected = 1;
  }
  written = snprintf(buffer, capacity, "%s@activation%zu", operation, activation);
  if (written < 0 || (size_t)written >= capacity) return 0;
  if (index > 0u) {
    size_t prior;
    for (prior = 0u; prior < index; prior++) {
      if (strcmp(buffer, phase24_owners[prior].semantic_id) == 0) {
        phase24_rejected = 1;
        phase24_record("attempt", "identity-collision");
        break;
      }
    }
  }
  return 1;
}

void *phase24_observer_malloc(size_t size) {
  unsigned char *pointer;
  phase24_owner_record *owner;
  size_t index;
  phase24_ensure_finalizer();
  pointer = (unsigned char *)malloc(size);
  if (pointer == NULL) return NULL;
  index = phase24_allocation_count++;
  if (size != 1u || index >= PHASE24_OWNER_COUNT) {
    phase24_rejected = 1;
    phase24_record("attempt", "unexpected-allocation");
    free(pointer);
    return NULL;
  }
  owner = &phase24_owners[index];
  owner->pointer = pointer;
  owner->live = 1;
  if (!phase24_semantic_identity(index, owner->semantic_id, sizeof(owner->semantic_id))) {
    phase24_rejected = 1;
    phase24_record("attempt", "missing-semantic-identity");
    owner->semantic_id[0] = '\0';
  }
  phase24_record("malloc", owner->semantic_id);
  return pointer;
}

unsigned char phase24_observer_load_byte(const unsigned char *pointer) {
  int index;
  unsigned char value;
  phase24_ensure_finalizer();
  index = phase24_find_owner(pointer);
  if (index != 2 || !phase24_owners[index].live || pointer == NULL || phase24_use_seen) {
    phase24_rejected = 1;
    phase24_record("attempt", "invalid-use");
    return 0u;
  }
  value = pointer[0];
  phase24_use_seen = 1;
  phase24_record("use", phase24_owners[index].semantic_id);
  return value;
}

void phase24_observer_transfer_return(void) {
  phase24_ensure_finalizer();
  if (phase24_allocation_count != PHASE24_OWNER_COUNT ||
      !phase24_owners[PHASE24_OWNER_COUNT - 1u].live || phase24_use_seen) {
    phase24_rejected = 1;
    phase24_record("attempt", "invalid-helper-return-boundary");
    return;
  }
  phase24_record("transfer_return", phase24_owners[PHASE24_OWNER_COUNT - 1u].semantic_id);
}

void phase24_observer_free(void *pointer) {
  int index;
  size_t expected_index;
  phase24_ensure_finalizer();
  index = phase24_find_owner(pointer);
  if (index < 0) {
    phase24_rejected = 1;
    phase24_record("attempt", "wrong-resource");
    return;
  }
  if (!phase24_owners[index].live) {
    phase24_rejected = 1;
    phase24_record("attempt", "duplicate-release");
    return;
  }
  if ((size_t)index == PHASE24_OWNER_COUNT - 1u && !phase24_use_seen) {
    phase24_rejected = 1;
    phase24_record("attempt", "premature-release");
    return;
  }
  if (phase24_release_count >= PHASE24_OWNER_COUNT) {
    phase24_rejected = 1;
    phase24_record("attempt", "free-order");
  } else {
    expected_index = PHASE24_OWNER_COUNT - 1u - phase24_release_count;
    if ((size_t)index != expected_index) {
      phase24_rejected = 1;
      phase24_record("attempt", "free-order");
    }
  }
  free(phase24_owners[index].pointer);
  phase24_owners[index].live = 0;
  phase24_release_count++;
  phase24_record("free", phase24_owners[index].semantic_id);
}

void phase24_observer_error_report_boundary(void) {
  size_t index;
  phase24_ensure_finalizer();
  for (index = 0u; index < phase24_allocation_count && index < PHASE24_OWNER_COUNT; index++) {
    if (phase24_owners[index].live) {
      phase24_rejected = 1;
      if (index == 2u) phase24_record("attempt", "omitted-release");
      phase24_record("attempt", "report-before-discharge");
    }
  }
  if (!phase24_use_seen || phase24_allocation_count != PHASE24_OWNER_COUNT || phase24_release_count != PHASE24_OWNER_COUNT) {
    phase24_rejected = 1;
    phase24_record("attempt", "incomplete-lifecycle");
  }
  phase24_record("report_boundary", phase24_outstanding() == 0u ? "post-release" : "outstanding-owner");
}

static void phase24_observer_finalize(void) {
  size_t index;
  int reject;
  for (index = 0u; index < phase24_allocation_count && index < PHASE24_OWNER_COUNT; index++) {
    if (phase24_owners[index].live && index == 2u && phase24_release_count < PHASE24_OWNER_COUNT) {
      phase24_rejected = 1;
      phase24_record("attempt", "omitted-release");
      break;
    }
  }
  reject = phase24_rejected || phase24_outstanding() != 0u ||
      phase24_allocation_count != PHASE24_OWNER_COUNT || phase24_release_count != PHASE24_OWNER_COUNT || !phase24_use_seen;
  phase24_record("final", phase24_outstanding() == 0u ? "outstanding=0" : "outstanding=1");
  phase24_record("verdict", reject ? "reject" : "pass");
  if (phase24_receipt != NULL) {
    if (fclose(phase24_receipt) != 0) reject = 1;
    phase24_receipt = NULL;
  }
  if (reject) _Exit(86);
}
