/* Test-only bridge for byte-frozen Phase 4/5 native evidence. The historical
 * generated C defines lang_nonlocal_landing; compile the current Schway
 * implementation against that exact landing-pad symbol, then let the shared
 * legacy adapter expose the original Lang-era foreign function name. */
#define schway_nonlocal_landing lang_nonlocal_landing
#include "../../../native/schway_foreign_nonlocal.c"
