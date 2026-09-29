/* Test-only aliases for the byte-preserved Phase 1-5 fixture corpus.
 * The fixtures keep their historical foreign names; current implementation
 * and ABI remain in the native/schway_foreign_*.c translation units. */

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

typedef struct _SCHWAY_schway_nonlocal_probe_result {
  unsigned char ok;
  unsigned char value;
} _SCHWAY_schway_nonlocal_probe_result;

typedef struct _LANG_lang_nonlocal_probe_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_nonlocal_probe_result;

extern _SCHWAY_schway_nonlocal_probe_result _SCHWAY_schway_nonlocal_probe(unsigned char argument);

_LANG_lang_nonlocal_probe_result _LANG_lang_nonlocal_probe(unsigned char argument) {
  _SCHWAY_schway_nonlocal_probe_result current = _SCHWAY_schway_nonlocal_probe(argument);
  _LANG_lang_nonlocal_probe_result result = {current.ok, current.value};
  return result;
}

typedef struct _SCHWAY_schway_arena_open_result {
  unsigned char ok;
  unsigned char value;
} _SCHWAY_schway_arena_open_result;

typedef struct _LANG_lang_arena_open_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_arena_open_result;

extern _SCHWAY_schway_arena_open_result _SCHWAY_schway_arena_open(unsigned char argument);

_LANG_lang_arena_open_result _LANG_lang_arena_open(unsigned char argument) {
  _SCHWAY_schway_arena_open_result current = _SCHWAY_schway_arena_open(argument);
  _LANG_lang_arena_open_result result = {current.ok, current.value};
  return result;
}

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
