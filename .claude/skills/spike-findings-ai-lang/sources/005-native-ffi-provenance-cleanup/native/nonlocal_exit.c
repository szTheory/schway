#include <setjmp.h>
#include <stdio.h>

static jmp_buf destination;
static int acquired = 0;
static int released = 0;

static void foreign_callback(void) { longjmp(destination, 1); }

static void call_with_resource(void) {
  acquired++;
  foreign_callback();
  released++;
}

int main(void) {
  if (setjmp(destination) == 0) call_with_resource();
  printf("{\"id\":\"nonlocal-exit\",\"valid\":%s,\"a\":%d,\"b\":%d,\"trace\":\"cleanup-skipped\"}\n",
         acquired == released ? "true" : "false", acquired, released);
  return acquired == released ? 0 : 6;
}
