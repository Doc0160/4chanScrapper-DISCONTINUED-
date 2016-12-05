/* ========================================================================
   $File: $
   $Date: $
   $Revision: $
   $Creator: Tristan Magniez $
   ======================================================================== */

package main

import (
    "time"
    "strconv"
)


func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func pause(durationS int64) {
	time.Sleep(time.Duration(durationS) * time.Second)
}

func format_o(i int)string {
	switch{
		case i>1024*1024*1024*1024:
		return strconv.Itoa(i/(1024*1024*1024*1024))+"To"
		case i>1024*1024*1024:
		return strconv.Itoa(i/(1024*1024*1024))+"Go"
		case i>1024*1024:
		return strconv.Itoa(i/(1024*1024))+"Mo"
		case i>1024:
		return strconv.Itoa(i/1024)+"ko"
	}
	return strconv.Itoa(i)+"o"
}
