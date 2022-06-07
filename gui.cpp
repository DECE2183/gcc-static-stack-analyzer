#include "gui.h"

#include <ftxui/dom/elements.hpp>  // for color, Fit, LIGHT, align_right, bold, DOUBLE
#include <ftxui/dom/table.hpp>      // for Table, TableSelection
#include <ftxui/screen/screen.hpp>  // for Screen
#include <iostream>                 // for endl, cout, ostream
#include <string>                   // for basic_string, allocator, string
#include <vector>                   // for vector

#include "ftxui/dom/node.hpp"  // for Render
#include "ftxui/screen/color.hpp"  // for Color, Color::Blue, Color::Cyan, Color::White, ftxui

using namespace ftxui;
using namespace std;

Element createLine(int indx, const char *file, const char *entry, int line, int column, int mem, float percent)
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
    gauge(percent) | size(WIDTH, EQUAL, 16)
  });
}

static string reset_position;

void draw(stack_call_t *calls, int calls_count, int total_mem)
{
  //stack_call_t *calls = (stack_call_t *)calls_addr;

  vector<Element> stack_list(calls_count * 2);
  for (int i = 0; i < calls_count; ++i)
  {
  	stack_list[i*2] = (createLine(i, calls[i].file_name, calls[i].entry_name,
  			calls[i].line, calls[i].column, calls[i].mem_usage, calls[i].mem_usage_percent));
  	stack_list[i*2 + 1] = (separator());
  }
  stack_list.empty();

  auto document = window(text("Stack"), {
    vbox({
      hflow({
        vbox({
          stack_list
        }) | flex,
      }) | border | flex,
      hbox({
        text("Total usage: " + to_string(total_mem) + " B") | flex,
      }) | border,
	})
  });

  auto screen = Screen::Create(Dimension::Full());
  Render(screen, document);
  cout << reset_position;
  screen.Print();
  reset_position = screen.ResetPosition();

  //std::cout << std::endl;
}
