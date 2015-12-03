package block

import "regexp"

var reOrderedList = regexp.MustCompile(`^( *([0-9]+)\. +)[^ ]`)
