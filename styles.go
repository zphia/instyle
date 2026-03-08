package instyle

import "math"

var styles = make(map[[keySizeMax]rune][]rune)

func init() {
	intermediate := map[string][]rune{
		"plain": []rune("22"),

		"reset":     []rune("0"),
		"bold":      []rune("1"),
		"faint":     []rune("2"),
		"italic":    []rune("3"),
		"underline": []rune("4"),
		"blink":     []rune("5"),
		"strike":    []rune("6"),

		"black":   []rune("30"),
		"red":     []rune("31"),
		"green":   []rune("32"),
		"yellow":  []rune("33"),
		"blue":    []rune("34"),
		"magenta": []rune("35"),
		"cyan":    []rune("36"),
		"white":   []rune("37"),
		"default": []rune("39"),

		"bg-black":   []rune("40"),
		"bg-red":     []rune("41"),
		"bg-green":   []rune("42"),
		"bg-yellow":  []rune("43"),
		"bg-blue":    []rune("44"),
		"bg-magenta": []rune("45"),
		"bg-cyan":    []rune("46"),
		"bg-white":   []rune("47"),
		"bg-default": []rune("49"),

		"light-black":   []rune("90"),
		"light-red":     []rune("91"),
		"light-green":   []rune("92"),
		"light-yellow":  []rune("93"),
		"light-blue":    []rune("94"),
		"light-magenta": []rune("95"),
		"light-cyan":    []rune("96"),
		"light-white":   []rune("97"),

		"bg-light-black":   []rune("100"),
		"bg-light-red":     []rune("101"),
		"bg-light-green":   []rune("102"),
		"bg-light-yellow":  []rune("103"),
		"bg-light-blue":    []rune("104"),
		"bg-light-magenta": []rune("105"),
		"bg-light-cyan":    []rune("106"),
		"bg-light-white":   []rune("107"),
	}
	for name, value := range intermediate {
		parsed := [keySizeMax]rune{}
		for k, v := range name[:int(math.Min(keySizeMax, float64(len(name))))] {
			parsed[k] = v
		}

		styles[parsed] = value
	}
}
