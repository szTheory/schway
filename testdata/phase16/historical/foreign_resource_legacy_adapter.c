/* Historical link adapter for the frozen Phase 16 generated-C controls.
 * Current Schway foreign symbols are exported by native/schway_foreign_resource.c.
 * Keep the old symbol namespace inside the historical fixture corpus so those
 * byte-frozen controls remain independently linkable after the source rename. */
typedef struct _LANG_lang_res_open_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_res_open_result;

_LANG_lang_res_open_result _LANG_lang_res_open(unsigned char argument) {
  _LANG_lang_res_open_result result;
  result.ok = 1;
  result.value = argument;
  return result;
}
