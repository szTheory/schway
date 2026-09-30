/* Test-only bridge for byte-frozen Phase 4/5 native evidence. The historical
 * generated C defines lang_nonlocal_landing; compile the current Schway
 * implementation against that exact landing-pad symbol, then expose the
 * original Lang-era foreign function name through this operation-specific
 * adapter. */
#define schway_nonlocal_landing lang_nonlocal_landing
#include "../../../native/schway_foreign_nonlocal.c"

typedef struct _LANG_lang_nonlocal_probe_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_nonlocal_probe_result;

_LANG_lang_nonlocal_probe_result _LANG_lang_nonlocal_probe(unsigned char argument) {
  _SCHWAY_schway_nonlocal_probe_result current = _SCHWAY_schway_nonlocal_probe(argument);
  _LANG_lang_nonlocal_probe_result result = {current.ok, current.value};
  return result;
}
