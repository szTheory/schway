/* Test-only bridge for byte-frozen Phase 4/5 native evidence. The historical
 * generated C defines lang_nonlocal_landing; compile the current Schway
 * implementation against that exact landing-pad symbol, then expose the
 * original Lang-era foreign function name through this operation-specific
 * adapter. */
#define schway_nonlocal_landing lang_nonlocal_landing
#include "../../../native/schway_foreign_nonlocal.c"

/* The frozen program terminates with abort() after writing its final JSON
 * record. Make stdout unbuffered before main so pipe capture sees the same
 * complete event stream on libc implementations whose abort does not flush. */
#include <stdio.h>

__attribute__((constructor))
static void schway_legacy_nonlocal_unbuffer_stdout(void) {
  (void)setvbuf(stdout, NULL, _IONBF, 0);
}

typedef struct _LANG_lang_nonlocal_probe_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_nonlocal_probe_result;

_LANG_lang_nonlocal_probe_result _LANG_lang_nonlocal_probe(unsigned char argument) {
  _SCHWAY_schway_nonlocal_probe_result current = _SCHWAY_schway_nonlocal_probe(argument);
  _LANG_lang_nonlocal_probe_result result = {current.ok, current.value};
  return result;
}
