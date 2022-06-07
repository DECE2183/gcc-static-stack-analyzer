package main

//#cgo CXXFLAGS: -I${SRCDIR}/FTXUI/include -std=c++17 -fstack-usage
//#cgo CFLAGS: -fstack-usage
//#cgo LDFLAGS: -L${SRCDIR}/FTXUI/build -lftxui-screen -lftxui-dom -lftxui-component
/*
#include <stdlib.h>
#include "gui.h"
*/
import "C"
import "unsafe"

func drawGUI(calls []StackCall, totalMem int) {
	var ccalls []C.stack_call_t
	for _, call := range calls {
		var newCall C.stack_call_t

		newCall.line = C.int(call.line)
		newCall.column = C.int(call.column)
		newCall.mem_usage = C.int(call.memUsage)
		newCall.mem_usage_percent = C.float(call.memUsagePercent)
		newCall.file_name = C.CString(call.fileName)
		newCall.entry_name = C.CString(call.entryName)
		newCall.qualifiers = C.CString(call.qualifiers)

		ccalls = append(ccalls, newCall);
	}
	defer func() {
		for i := range ccalls {
			C.free(unsafe.Pointer(ccalls[i].file_name))
			C.free(unsafe.Pointer(ccalls[i].entry_name))
			C.free(unsafe.Pointer(ccalls[i].qualifiers))
		}
	}()

	ptr := (*C.stack_call_t)(&ccalls[0])

	C.draw(ptr, C.int(len(calls)), C.int(totalMem))
}
