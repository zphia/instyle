package instyle

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	// keySizeMax is used to determine the maximum size of a named style.
	// This is important for optimization in that it allows for a map to be used to find styles by name.
	keySizeMax = 64

	// styleDepthMax is used to determine how many levels of nesting are allowed.
	styleDepthMax = 5
)

var (
	openingBegin = []rune("[~")
	openingClose = []rune("]")
	closing      = []rune("[/]")

	reset = []rune{'\033', '[', '0', 'm'}
)

type Styler interface {
	// Apply will parse and replace any valid style tags in the original rune array and return the result.
	//
	// Style tags are applied to a maximum depth of 5 nested tags.
	Apply(original []rune) (output []rune)

	// ApplyStr will call Apply while casting the arguments to/from strings.
	// There is a reasonable performance hit to this over Apply.
	ApplyStr(original string) (output string)

	// ApplyStrf will call Apply while casting the arguments to/from strings and using fmt.Sprintf.
	// Only the format string will be parsed for style tags.
	//
	// There is a reasonable performance hit to this over Apply.
	ApplyStrf(format string, args ...any) (output string)

	// Register will add a new named style which can be used in future calls of Apply.
	// The value expected is an ANSI escape code such as `31` for red.
	// To have a style map to multiple ANSI escape codes, separate them with a semicolon.
	//
	// Names have a maximum length of 16 characters.
	//
	//    s := instyle.NewStyler()
	//    s.Register("error", "1;31")
	//    _ = s.Apply([]rune("[~error]Something unexpected happened"))
	Register(name string, value string) (self Styler)

	// RegisterLipGlossStyle will extract the text styling from a [lipgloss.Style] and register it under the name provided.
	// Specifically, the following will be captured if set on the style:
	//
	//  - Foreground color
	//  - Background color
	//  - Text styling of bold / faint / italic / underline / blink / strikethrough
	//
	// [lipgloss.Style]: https://github.com/charmbracelet/lipgloss
	RegisterLipGlossStyle(name string, value lipgloss.Style) (self Styler)
}

type styleSet struct {
	named map[[keySizeMax]rune][]rune
}

func NewStyler() Styler {
	s := new(styleSet)
	s.named = make(map[[keySizeMax]rune][]rune)
	return s
}

func (s *styleSet) Register(name string, value string) Styler {
	parsed := [keySizeMax]rune{}
	for k, v := range name[:int(math.Min(keySizeMax, float64(len(name))))] {
		parsed[k] = v
	}

	s.named[parsed] = []rune(value)
	return s
}

func (s *styleSet) RegisterLipGlossStyle(name string, value lipgloss.Style) Styler {
	p := lipgloss.ColorProfile()

	var sequence []string

	if _, noColor := value.GetForeground().(lipgloss.NoColor); !noColor {
		sequence = append(sequence, p.FromColor(value.GetForeground()).Sequence(false))
	}

	if _, noColor := value.GetBackground().(lipgloss.NoColor); !noColor {
		sequence = append(sequence, p.FromColor(value.GetBackground()).Sequence(true))
	}

	if value.GetBold() {
		sequence = append(sequence, "1")
	}

	if value.GetFaint() {
		sequence = append(sequence, "2")
	}

	if value.GetItalic() {
		sequence = append(sequence, "3")
	}

	if value.GetUnderline() {
		sequence = append(sequence, "4")
	}

	if value.GetBlink() {
		sequence = append(sequence, "5")
	}

	if value.GetStrikethrough() {
		sequence = append(sequence, "6")
	}

	return s.Register(name, strings.Join(sequence, ";"))
}

func (s *styleSet) Apply(runes []rune) []rune {
	var (
		appliedStyleStack = [styleDepthMax][]rune{}
		ok                = false
		output            = make([]rune, 0, len(runes)*4/3+10) // Pre-allocate n * 1.33 + 10 the size of the passed runes.
	)

	output = append(output, reset...)

	for i, nest := 0, 0; i < len(runes); i++ {
		r := runes[i]

		if r == openingBegin[0] && nest < styleDepthMax {
			if appliedStyleStack[nest], i, ok = s.parseOpening(runes, i); ok {
				if nest = nest + 1; nest > 0 {
					output = append(output, appliedStyleStack[nest-1]...)
					continue
				}
			}
		}

		if r == closing[0] && nest > 0 {
			if i, ok = checkSequence(closing, runes, i); ok {
				if nest = nest - 1; nest >= 0 {
					output = append(output, reset...)
					appliedStyleStack[nest] = nil

					if i+1 == len(runes) {
						appliedStyleStack[0] = nil
					}

					for i := 0; i < len(appliedStyleStack) && i < nest; i++ {
						output = append(output, appliedStyleStack[i]...)
					}

					continue
				}
			}
		}

		output = append(output, r)
	}

	if appliedStyleStack[0] != nil {
		output = append(output, reset...)
	}

	return output
}

// ApplyStr will call Apply while casting the arguments to/from strings.
// There is a reasonable performance hit to this over Apply.
func (s *styleSet) ApplyStr(original string) (output string) {
	return string(s.Apply([]rune(original)))
}

// ApplyStrf will call Apply while casting the arguments to/from strings and using fmt.Sprintf.
// There is a reasonable performance hit to this over Apply.
func (s *styleSet) ApplyStrf(format string, args ...any) (output string) {
	return fmt.Sprintf(string(s.Apply([]rune(format))), args...)
}

// parseOpening operates similarly to checkSequence but specifically for the opening of a style tag.
// When a valid style tag is found, the computed sequence of ANSI style runes is returned.
func (s *styleSet) parseOpening(runes []rune, idx int) ([]rune, int, bool) {
	after, ok := checkSequence(openingBegin, runes, idx)
	if !ok {
		return nil, idx, false
	}

	sequence := make([]rune, 0, 10)
	sequence = append(sequence, '\033', '[')

	first := true

	var (
		isNumeric  = true
		isMaybeHex = true
		isMaybeRGB = true
	)

	key := [keySizeMax]rune{}

	for i, count := after+1, 0; i < len(runes); i++ {
		r := runes[i]

		if isClose := r == openingClose[0]; isClose || r == '+' {
			if count == 0 || count >= keySizeMax {
				return nil, idx, false
			}

			if !first {
				sequence = append(sequence, ';')
			}

			first = false

			if found, ok := s.named[key]; ok {
				sequence = append(sequence, found...)
			} else if found, ok := styles[key]; ok {
				sequence = append(sequence, found...)
			} else {
				var match = false

				if isMaybeHex {
					var buffer [16]rune
					buffer[0] = '3'
					buffer[1] = '8'
					buffer[2] = ';'
					buffer[3] = '2'
					buffer[4] = ';'

					n := 5

					if count == 4 {
						n += copy(buffer[n:], byteRunes[hexLookup[key[1]]*16])

						buffer[n] = ';'
						n++

						n += copy(buffer[n:], byteRunes[hexLookup[key[2]]*16])

						buffer[n] = ';'
						n++

						n += copy(buffer[n:], byteRunes[hexLookup[key[3]]*16])

						sequence = append(sequence, buffer[:n]...)
						match = true
					}

					if count == 7 {
						n += copy(buffer[n:], byteRunes[hexLookup[key[1]]*16+hexLookup[key[2]]])

						buffer[n] = ';'
						n++

						n += copy(buffer[n:], byteRunes[hexLookup[key[3]]*16+hexLookup[key[4]]])

						buffer[n] = ';'
						n++

						n += copy(buffer[n:], byteRunes[hexLookup[key[5]]*16+hexLookup[key[6]]])

						sequence = append(sequence, buffer[:n]...)
						match = true
					}
				}

				if !match {
					if isMaybeRGB && (count >= 10 && count <= 16) {
						rgb := [3]int{}

						digit := 0 // which digit in a number we're looking at
						stage := 0 // the current index of rgb being built 0=r, 1=g, 2=b

						success := true

						for _, v := range key[4 : count-1] {
							if v == ',' && stage < 3 && digit > 0 {
								digit = 0
								stage = stage + 1
							} else if v >= '0' && v <= '9' && digit < 3 {
								rgb[stage] = rgb[stage]*10 + int(v-'0')
								digit++
							} else {
								success = false
								break
							}
						}

						if success && stage == 2 && rgb[0] <= 255 && rgb[1] <= 255 && rgb[2] <= 255 {
							var buffer [16]rune
							buffer[0] = '3'
							buffer[1] = '8'
							buffer[2] = ';'
							buffer[3] = '2'
							buffer[4] = ';'

							n := 5
							n += copy(buffer[n:], byteRunes[rgb[0]])

							buffer[n] = ';'
							n++

							n += copy(buffer[n:], byteRunes[rgb[1]])

							buffer[n] = ';'
							n++

							n += copy(buffer[n:], byteRunes[rgb[2]])

							sequence = append(sequence, buffer[:n]...)
							match = true
						}
					}
				}

				if !match {
					if isNumeric {
						sequence = append(sequence, key[:count]...)
						match = true
					}
				}

				if !match {
					return nil, idx, false
				}
			}

			if isClose {
				var ok bool
				if after, ok = checkSequence(openingClose, runes, i); ok {
					break
				} else {
					return nil, idx, false
				}
			}

			count = 0

			isNumeric = true
			isMaybeHex = true
			isMaybeRGB = true

			key = [keySizeMax]rune{}
			continue
		}

		if isNumeric {
			if r < '0' || r > '9' {
				isNumeric = false
			}
		}

		if isMaybeHex {
			if count == 0 && r != '#' || (count > 0 && !((r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))) {
				isMaybeHex = false
			}
		}

		if isMaybeRGB {
			if !(r >= '0' && r <= '9' || r == ',' || r == ')' || r == '(' || r == 'r' || r == 'b' || r == 'g') {
				isMaybeRGB = false
			}
		}

		key[count] = r
		count++
	}

	return append(sequence, 'm'), after, true
}

// checkSequence will attempt to find a sequence of runes at a given index.
// If the sequence is found, the runes index at the end of the sequence is returned.
func checkSequence(sequence []rune, runes []rune, idx int) (int, bool) {
	lenRunes, lenSequence := len(runes), len(sequence)

	// Determine if the sequence would be impossible given current lengths:
	if lenRunes < lenSequence || lenRunes-idx < lenSequence {
		return idx, false
	}

	// Attempt to find the sequence:
	for i := 0; i < lenSequence; i++ {
		if sequence[i] != runes[idx+i] {
			return idx, false
		}
	}

	// Return the index after the sequence:
	return idx + lenSequence - 1, true
}
