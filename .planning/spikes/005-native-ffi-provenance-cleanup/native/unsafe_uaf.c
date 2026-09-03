#include <stdlib.h>

int main(void) {
  volatile int *value = malloc(sizeof(int));
  if (!value) return 4;
  *value = 17;
  free((void *)value);
  return *value == 17 ? 0 : 5;
}
