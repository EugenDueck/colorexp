package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"regexp"
)

const version = "1.0.5"

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

func match(line string, regexps []*regexp.Regexp, varyGroupColors bool) []rangeWithID {
	var ranges []rangeWithID
	colorIdx := 0
	for _, re := range regexps {
		numGroups := re.NumSubexp()
		matchRanges := re.FindAllStringSubmatchIndex(line, -1)
		firstGroupToColorize := min(1, numGroups)
		groupsToColorize := numGroups + 1 - firstGroupToColorize
		for _, matchRange := range matchRanges {
			// if there is no capturing group, the full match will be colorized (group 0)
			// if there are capturing groups, all groups but group 0 (the full match) will be colorized
			for i := 0; i < groupsToColorize; i++ {
				curColorIdx := colorIdx
				if varyGroupColors {
					curColorIdx += i
				}
				gIdx := (i + firstGroupToColorize) * 2
				matchRangeStart := matchRange[gIdx]
				matchRangeEnd := matchRange[gIdx+1]
				if matchRangeEnd > matchRangeStart {
					ranges = addRange(ranges, rangeWithID{matchRangeStart, matchRangeEnd, curColorIdx})
				}
			}
		}
		colorIdx += groupsToColorize
	}
	return ranges
}

func colorize(s string, colors []string, resetColor string, ranges []rangeWithID) string {
	for i, r := range ranges {
		color := colors[r.id%len(colors)]
		s = insertString(s, color, r.startIndex)
		incRanges(ranges, len(color))
		// ranges[i] was modified by incRanges, so we need to use that, not the stale r
		s = insertString(s, resetColor, ranges[i].endIndex)
		incRanges(ranges, len(resetColor))
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
		showHelp           bool
		highlight          bool
		ignoreCase         bool
		showVersion        bool
		varyGroupColorsOn  bool
		varyGroupColorsOff bool
		varyGroupColors    bool
	)

	pflag.BoolVarP(&fixedStrings, "fixed-strings", "F", false, "Do not interpret regular expression metacharacters.")
	pflag.BoolVarP(&showHelp, "help", "h", false, "Display this help and exit.")
	pflag.BoolVarP(&highlight, "highlight", "H", false, "Color by changing the background color. The default is to change the foreground color.")
	pflag.BoolVarP(&ignoreCase, "ignore-case", "i", false, "Perform case insensitive matching.")
	pflag.BoolVarP(&showVersion, "version", "V", false, "Display version information and exit.")

	pflag.BoolVarP(&varyGroupColorsOn, "vary-group-colors-on", "G", false, "Turn on changing of colors for every capturing group. Defaults to on if exactly one pattern is given.")
	pflag.BoolVarP(&varyGroupColorsOff, "vary-group-colors-off", "g", false, "Turn off changing of colors for every capturing group. Defaults to on if exactly one pattern is given.")

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
		_, _ = fmt.Println("Error: At least one pattern argument is required.\n")
		printUsage()
		os.Exit(1)
	}

	if pflag.Lookup("vary-group-colors-on").Changed {
		if pflag.Lookup("vary-group-colors-off").Changed {
			_, _ = fmt.Println("Error: -G/-g arguments cannot both be used at the same time.\n")
			printUsage()
			os.Exit(1)
		}
		varyGroupColors = true
	} else if pflag.Lookup("vary-group-colors-off").Changed {
		varyGroupColors = false
	} else {
		varyGroupColors = len(regexStrings) == 1
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
		// insert at the beginning of the slice, to invert the order,
		// so that the last regex takes precedence
		regexps = append([]*regexp.Regexp{re}, regexps...)
	}

	var colors []string
	var resetColor string
	if highlight {
		colors = backgroundColors
		resetColor = resetBackgroundColor
	} else {
		colors = foregroundColors
		resetColor = resetForegroundColor
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		ranges := match(line, regexps, varyGroupColors)
		colorizedLine := colorize(line, colors, resetColor, ranges)
		fmt.Println(colorizedLine)
	}

	if err := scanner.Err(); err != nil {
		_, _ = fmt.Printf("Error reading standard input: %v\n", err)
		os.Exit(2)
	}
}
