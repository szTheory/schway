#include <stdint.h>

typedef uint8_t T;

static T f(T *restrict p) {
  T copy = *p;
  return copy;
}

int main(void) {
  T value = 37;
  T before = value;
  T copied = f(&value);
  T after = value;
  return before == 37 && copied == 37 && after == 37 ? 0 : 1;
}
