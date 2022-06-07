#include "gui.h"

#include <memory>
#include <iostream>
#include <string>
#include <vector>

#include <ftxui/component/component.hpp>       // for Renderer, Button, Vertical
#include <ftxui/component/component_base.hpp>  // for ComponentBase
#include <ftxui/dom/elements.hpp>  // for operator|, Element, text, bold, border, center, color
#include <ftxui/screen/color.hpp>  // for Color, Color::Red

using namespace ftxui;
using namespace std;

Element createLine(int indx, const char *file, const char *entry, int line, int column, int mem, float percent, const char *qualifiers)
{
  return hbox({
    text(to_string(indx)) | size(WIDTH, EQUAL, 6),
    separator(),
    text(file) | flex,
    text(entry),
    separator(),
    text(to_string(line) + ":" + to_string(column)) | size(WIDTH, EQUAL, 10),
    separator(),
    text(to_string(mem) + " B") | size(WIDTH, EQUAL, 10),
    separator(),
    gauge(percent) | size(WIDTH, EQUAL, 16),
    separator(),
    text(qualifiers) | size(WIDTH, EQUAL, 10)
  });
}

static string reset_position;

void draw(stack_call_t *calls, int calls_count, int total_mem)
{
  vector<Element> stack_list(calls_count * 2);
  for (int i = 0; i < calls_count; ++i)
  {
  	stack_list[i*2] = (createLine(i, calls[i].file_name, calls[i].entry_name,
  			calls[i].line, calls[i].column, calls[i].mem_usage, calls[i].mem_usage_percent,
        calls[i].qualifiers));
  	stack_list[i*2 + 1] = (separator());
  }

  auto document = window(text("Stack"), {
    vbox({
      vbox({
        vbox({
          stack_list
        }) | border,
      }) | flex | vscroll_indicator | yframe,
      separator(),
      text("Total usage: " + to_string(total_mem) + " B") | size(HEIGHT, EQUAL, 1),
	  })
  });

  auto screen = Screen::Create(Dimension::Full());
  Render(screen, document);
  cout << reset_position;
  screen.Print();
  reset_position = screen.ResetPosition();
}
