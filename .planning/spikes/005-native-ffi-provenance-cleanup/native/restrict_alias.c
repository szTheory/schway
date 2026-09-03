#include <stdio.h>

#ifdef UNSOUND_NOALIAS
static int update(int *restrict left, int *restrict right) {
#else
static int update(int *left, int *right) {
#endif
  *left = 1;
  *right = 2;
  return *left;
}

int main(void) {
  int value = 0;
  int observed = update(&value, &value);
  printf("%d\n", observed);
  return observed == 2 ? 0 : 3;
}
