package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"regexp"
)

const version = "2.2"

var foregroundColors = []string{
	//"\033[30m", // Black
	"\033[31m", // Red
	"\033[32m", // Green
	"\033[33m", // Yellow
	"\033[34m", // Blue
	"\033[35m", // Magenta
	"\033[36m", // Cyan
	//"\033[37m", // White
}

var backgroundColors = []string{
	//"\033[40m", // Black
	"\033[41m", // Red
	"\033[44m", // Blue
	"\033[45m", // Magenta
	"\033[42m", // Green
	"\033[43m", // Yellow
	"\033[46m", // Cyan
	//"\033[47m", // White
}

const resetForegroundColor = "\033[0m"
const resetBackgroundColor = "\033[49m"

const maxLineLength int = 1_048_576

func insertString(original string, toInsert string, index int) string {
	if index < 0 || index > len(original) {
		panic("index out of bounds")
	}
	return original[:index] + toInsert + original[index:]
}

type rangeWithID struct {
	startIndex, endIndex, id int
}

// addRange adds a new range to the ordered list of non-overlapping ranges.
// It ensures that the list stays ordered and any existing ranges are subtracted
// from the new range, potentially splitting it into multiple pieces.
func addRange(ranges []rangeWithID, newRange rangeWithID) []rangeWithID {
	var result []rangeWithID
	inserted := false

	for _, existingRange := range ranges {
		if newRange.endIndex <= existingRange.startIndex {
			// The new range is entirely before the existing range.
			if !inserted {
				result = append(result, newRange)
				inserted = true
			}
			result = append(result, existingRange)
		} else if newRange.startIndex >= existingRange.endIndex {
			// The new range is entirely after the existing range.
			result = append(result, existingRange)
		} else {
			// There is an overlap; we may need to split the new range.
			if !inserted && newRange.startIndex < existingRange.startIndex {
				// Add the non-overlapping piece before the existing range.
				result = append(result, rangeWithID{newRange.startIndex, existingRange.startIndex, newRange.id})
			}
			result = append(result, existingRange)
			if newRange.endIndex > existingRange.endIndex {
				// Update the new range to start from the end of the existing range.
				newRange.startIndex = existingRange.endIndex
			} else {
				// The new range is fully covered by the existing range; nothing left to add.
				inserted = true
				newRange.startIndex = newRange.endIndex
			}
		}
	}

	// If the new range was not inserted because it is after all existing ranges,
	// or if it still has a remaining piece after processing overlaps, add it now.
	if !inserted {
		result = append(result, newRange)
	}

	return result
}

func match(line string, regexps []*regexp.Regexp, varyGroupColors, fullMatchHighlight bool) []rangeWithID {
	var ranges []rangeWithID
	colorIdx := 0
	for _, re := range regexps {
		numGroups := re.NumSubexp()
		matchRanges := re.FindAllStringSubmatchIndex(line, -1)
		firstGroupToColorize := min(1, numGroups)
		groupsToColorize := numGroups + 1 - firstGroupToColorize
		if fullMatchHighlight {
			firstGroupToColorize = 0
			groupsToColorize = 1
		}
		for _, matchRange := range matchRanges {
			// if there is no capturing group, the full match will be colorized (group 0)
			// if there are capturing groups, all groups but group 0 (the full match) will be colorized, unless
			// fullMatchHighlight == true
			for i := 0; i < groupsToColorize; i++ {
				curColorIdx := colorIdx
				if varyGroupColors {
					curColorIdx += groupsToColorize - 1 - i
				}
				gIdx := (i + firstGroupToColorize) * 2
				matchRangeStart := matchRange[gIdx]
				matchRangeEnd := matchRange[gIdx+1]
				if matchRangeEnd > matchRangeStart {
					ranges = addRange(ranges, rangeWithID{matchRangeStart, matchRangeEnd, curColorIdx})
				}
			}
		}
		if varyGroupColors {
			colorIdx += groupsToColorize
		} else {
			colorIdx++
		}
	}
	return ranges
}

func colorize(s string, colors [][]string, ranges []rangeWithID, patternColorCount int) string {
	for i, r := range ranges {
		colorIdx := patternColorCount - r.id - 1
		for colorIdx < 0 {
			colorIdx += len(colors)
		}
		color := colors[colorIdx%len(colors)]
		s = insertString(s, color[0], r.startIndex)
		incRanges(ranges, len(color[0]))
		// ranges[i] was modified by incRanges, so we need to use that, not the stale r
		s = insertString(s, color[1], ranges[i].endIndex)
		incRanges(ranges, len(color[1]))
	}
	return s
}

func incRanges(ranges []rangeWithID, inc int) {
	for i, r := range ranges {
		ranges[i] = rangeWithID{r.startIndex + inc, r.endIndex + inc, r.id}
	}
}

func printUsage() {
	fmt.Println("Usage: colorexp [options] patterns...")
	pflag.PrintDefaults()
}

func main() {
	var (
		fixedStrings       bool
		fullMatchHighlight bool
		showHelp           bool
		noHighlight        bool
		onlyHighlight      bool
		ignoreCase         bool
		showVersion        bool
		varyGroupColorsOn  bool
		varyGroupColorsOff bool
		varyGroupColors    bool
		onlyMatchingLines  bool
	)

	pflag.BoolVarP(&fixedStrings, "fixed-strings", "F", false, "Do not interpret regular expression metacharacters.")
	pflag.BoolVarP(&fullMatchHighlight, "full-match-highlight", "f", false, "Highlight the entire match, even if pattern contains capturing groups.")
	pflag.BoolVarP(&showHelp, "help", "", false, "Display this help and exit.")
	pflag.BoolVarP(&noHighlight, "no-highlight", "h", false, "Do not color by changing the background color.")
	pflag.BoolVarP(&onlyHighlight, "only-highlight", "H", false, "Only color by changing the background color.")
	pflag.BoolVarP(&ignoreCase, "ignore-case", "i", false, "Perform case insensitive matching.")
	pflag.BoolVarP(&showVersion, "version", "V", false, "Display version information and exit.")

	pflag.BoolVarP(&varyGroupColorsOn, "vary-group-colors-on", "G", false, "Turn on changing of colors for every capturing group. Defaults to on if exactly one pattern is given.")
	pflag.BoolVarP(&varyGroupColorsOff, "vary-group-colors-off", "g", false, "Turn off changing of colors for every capturing group. Defaults to on if exactly one pattern is given.")

	pflag.BoolVarP(&onlyMatchingLines, "only-matching-lines", "o", false, "Only print lines with matches (suppress lines without matches).")

	pflag.Parse()

	if showVersion {
		fmt.Printf("colorexp %v\n", version)
		os.Exit(0)
	}

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	regexStrings := pflag.Args()

	if len(regexStrings) == 0 {
		_, _ = fmt.Printf("Error: At least one pattern argument is required.\n\n")
		printUsage()
		os.Exit(1)
	}

	varyGroupColorsOnChanged := pflag.Lookup("vary-group-colors-on").Changed
	varyGroupColorsOffChanged := pflag.Lookup("vary-group-colors-off").Changed
	if varyGroupColorsOnChanged {
		if varyGroupColorsOffChanged {
			_, _ = fmt.Printf("Error: -g/-G arguments cannot both be used at the same time.\n\n")
			printUsage()
			os.Exit(1)
		}
		varyGroupColors = true
	} else if varyGroupColorsOffChanged {
		varyGroupColors = false
	} else {
		varyGroupColors = len(regexStrings) == 1
	}

	if fullMatchHighlight && (varyGroupColorsOnChanged || varyGroupColorsOffChanged) {
		_, _ = fmt.Printf("Error: -f and -g/-G arguments cannot both be used at the same time.\n\n")
		printUsage()
		os.Exit(1)
	}

	// Note that the order in regexps is the reverse of the original order, to implement the "last regexp wins" logic
	var regexps []*regexp.Regexp
	for _, regexString := range regexStrings {
		if fixedStrings {
			regexString = regexp.QuoteMeta(regexString)
		}
		if ignoreCase {
			regexString = "(?i)" + regexString
		}
		re, err := regexp.Compile(regexString)
		if err != nil {
			_, _ = fmt.Printf("Invalid regular expression: %v\n", err)
			os.Exit(1)
		}
		// insert at the beginning of the slice, to reverse the order,
		// so that the last regex takes precedence
		regexps = append([]*regexp.Regexp{re}, regexps...)
		//regexps = append(regexps, re)
	}

	var colors [][]string
	if !pflag.Lookup("only-highlight").Changed {
		for _, foreColor := range foregroundColors {
			colors = append(colors, []string{foreColor, resetForegroundColor})
		}
	} else if pflag.Lookup("no-highlight").Changed {
		_, _ = fmt.Printf("Error: -h/-H arguments cannot both be used at the same time.\n\n")
		printUsage()
		os.Exit(1)
	}
	if !pflag.Lookup("no-highlight").Changed {
		for _, backColor := range backgroundColors {
			colors = append(colors, []string{backColor, resetBackgroundColor})
		}
	}

	patternColorCount := len(regexps)
	if varyGroupColors {
		for _, re := range regexps {
			patternColorCount += max(0, re.NumSubexp()-1)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	buf := make([]byte, maxLineLength)
	scanner.Buffer(buf, maxLineLength)
	for scanner.Scan() {
		line := scanner.Text()
		ranges := match(line, regexps, varyGroupColors, fullMatchHighlight)
		if onlyMatchingLines && len(ranges) == 0 {
			continue
		}
		colorizedLine := colorize(line, colors, ranges, patternColorCount)
		fmt.Println(colorizedLine)
	}

	if err := scanner.Err(); err != nil {
		_, _ = fmt.Printf("Error reading standard input: %v\n", err)
		os.Exit(2)
	}
}
