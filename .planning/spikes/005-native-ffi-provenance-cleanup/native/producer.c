#include <stdint.h>

struct LangPacket {
  uint64_t id;
  uint32_t count;
  uint8_t tag;
};

struct LangPacket make_packet(void) {
  struct LangPacket packet = {0x1122334455667788ULL, 7U, 9U};
  return packet;
}
