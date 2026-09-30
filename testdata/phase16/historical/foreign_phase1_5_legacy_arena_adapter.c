/* Test-only allocator alias for frozen Phase 5 C evidence. */
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
