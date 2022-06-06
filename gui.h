#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

typedef struct {
  int line;
  int column;
  int mem_usage;
  float mem_usage_percent;
  const char *file_name;
  const char *entry_name;
  const char *qualifiers;
} stack_call_t;

void draw(stack_call_t *calls, int calls_count, int total_mem);

#ifdef __cplusplus
}
#endif
