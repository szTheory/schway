/* Test-only resource alias for the byte-preserved Phase 1-5 fixture corpus.
 * Other foreign operations have separate adapters so a fixture links only
 * the implementation symbols it declares. */

typedef struct _SCHWAY_schway_res_open_result {
  unsigned char ok;
  unsigned char value;
} _SCHWAY_schway_res_open_result;

typedef struct _LANG_lang_res_open_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_res_open_result;

extern _SCHWAY_schway_res_open_result _SCHWAY_schway_res_open(unsigned char argument);

_LANG_lang_res_open_result _LANG_lang_res_open(unsigned char argument) {
  _SCHWAY_schway_res_open_result current = _SCHWAY_schway_res_open(argument);
  _LANG_lang_res_open_result result = {current.ok, current.value};
  return result;
}
