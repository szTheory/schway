#include <stddef.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>

#ifdef INJECT_LAYOUT_BUG
struct LangPacket {
  uint8_t tag;
  uint32_t count;
  uint64_t id;
};
#else
struct LangPacket {
  uint64_t id;
  uint32_t count;
  uint8_t tag;
};
#endif

extern struct LangPacket make_packet(void);

static int failures = 0;

static void emit(const char *id, int valid, unsigned long long a,
                 unsigned long long b, const char *trace) {
  printf("{\"id\":\"%s\",\"valid\":%s,\"a\":%llu,\"b\":%llu,\"trace\":\"%s\"}\n",
         id, valid ? "true" : "false", a, b, trace);
  if (!valid) failures++;
}

static void layout_probe(void) {
  struct LangPacket packet = make_packet();
  int valid = packet.id == 0x1122334455667788ULL && packet.count == 7U && packet.tag == 9U;
  emit("layout-roundtrip", valid, (unsigned long long)packet.id,
       (unsigned long long)packet.count, valid ? "tag:9" : "abi-mismatch");
}

struct ResourceState {
  int acquired;
  int released;
  char trace[96];
  size_t used;
};

static void append_trace(struct ResourceState *state, const char *text) {
  int written = snprintf(state->trace + state->used, sizeof(state->trace) - state->used,
                         "%s%s", state->used ? "," : "", text);
  if (written > 0) state->used += (size_t)written;
}

static void acquire(struct ResourceState *state, const char *name) {
  state->acquired++;
  append_trace(state, name);
}

static void release(struct ResourceState *state, const char *name) {
  state->released++;
  append_trace(state, name);
}

static void cleanup_probe(int fail_stage, const char *id) {
  struct ResourceState state = {0};
  int has_a = 0;
  int has_b = 0;
  acquire(&state, "acquire:a");
  has_a = 1;
  if (fail_stage == 1) goto cleanup;
  acquire(&state, "acquire:b");
  has_b = 1;
  if (fail_stage == 2) goto cleanup;

cleanup:
  if (has_b) {
#ifndef INJECT_CLEANUP_LEAK
    release(&state, "release:b");
#endif
  }
  if (has_a) release(&state, "release:a");
  emit(id, state.acquired == state.released,
       (unsigned long long)state.acquired,
       (unsigned long long)state.released, state.trace);
}

struct AllocationHeader {
  uint32_t magic;
  uint32_t allocator;
};

static void *owned_alloc(uint32_t allocator, size_t size) {
  struct AllocationHeader *header = malloc(sizeof(*header) + size);
  if (!header) return NULL;
  header->magic = 0xA110CA7EU;
  header->allocator = allocator;
  return header + 1;
}

static int owned_free(uint32_t expected_allocator, void *pointer) {
  struct AllocationHeader *header = ((struct AllocationHeader *)pointer) - 1;
  if (header->magic != 0xA110CA7EU) return 0;
#ifndef INJECT_ALLOCATOR_PAIRING_BUG
  if (header->allocator != expected_allocator) return 0;
#else
  (void)expected_allocator;
#endif
  header->magic = 0;
  free(header);
  return 1;
}

static void allocator_probe(void) {
  void *pointer = owned_alloc(7U, 32U);
  if (!pointer) {
    emit("allocator-pairing", 0, 0, 0, "allocation-failed");
    return;
  }
  int wrong_rejected = !owned_free(8U, pointer);
#ifdef INJECT_ALLOCATOR_PAIRING_BUG
  emit("allocator-pairing", 0, (unsigned long long)wrong_rejected, 0,
       "wrong-allocator-accepted");
#else
  int right_accepted = owned_free(7U, pointer);
  emit("allocator-pairing", wrong_rejected && right_accepted, 1, 1,
       "mismatch-rejected,release");
#endif
}

struct OwnerToken { uint64_t generation; int alive; };
struct BorrowedView { const int *pointer; const struct OwnerToken *owner; uint64_t generation; };

static int read_view(struct BorrowedView view, int *output) {
#ifndef INJECT_RETENTION_BUG
  if (!view.owner->alive || view.owner->generation != view.generation) return 0;
#endif
  *output = *view.pointer;
  return 1;
}

static void retention_probe(void) {
  int value = 41;
  int output = 0;
  struct OwnerToken owner = {3U, 1};
  struct BorrowedView retained = {&value, &owner, owner.generation};
  int live_read = read_view(retained, &output) && output == 41;
  owner.alive = 0;
  owner.generation++;
  int stale_rejected = !read_view(retained, &output);
  emit("callback-retention", live_read && stale_rejected, 1,
       (unsigned long long)stale_rejected,
       stale_rejected ? "live-read,stale-rejected" : "stale-read-accepted");
}

int main(void) {
  layout_probe();
  cleanup_probe(1, "cleanup-fail-a");
  cleanup_probe(2, "cleanup-fail-b");
  cleanup_probe(0, "cleanup-success");
  allocator_probe();
  retention_probe();
  return failures == 0 ? 0 : 2;
}
