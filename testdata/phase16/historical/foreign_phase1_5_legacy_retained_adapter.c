/* Test-only retained-pointer alias for frozen Phase 5 C evidence. */
typedef struct _SCHWAY_schway_retained_touch_result {
  unsigned char ok;
  unsigned char value;
} _SCHWAY_schway_retained_touch_result;

typedef struct _LANG_lang_retained_touch_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_retained_touch_result;

extern _SCHWAY_schway_retained_touch_result _SCHWAY_schway_retained_touch(unsigned char argument);

_LANG_lang_retained_touch_result _LANG_lang_retained_touch(unsigned char argument) {
  _SCHWAY_schway_retained_touch_result current = _SCHWAY_schway_retained_touch(argument);
  _LANG_lang_retained_touch_result result = {current.ok, current.value};
  return result;
}
