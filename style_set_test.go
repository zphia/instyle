package instyle_test

import (
	"testing"

	"github.com/zphia/instyle"
)

const (
	noStyle   = "[tag] sending request 3.2 second ago... log id: 10298402358"
	withStyle = "[~bold][tag][/] sending request [~faint]3.2 seconds ago...[/] [~bold][~cyan]log id:[/] [~magenta]10298402358[/][/]"
)

func BenchmarkBaseline(b *testing.B) {
	b.Run("BestCase", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			runes := make([]rune, 0, len(noStyle))
			for _, r := range []rune(noStyle) {
				runes = append(runes, r)
			}
		}
	})

	b.Run("PerformanceGoal", func(b *testing.B) { // Ideally it's as fast as an unoptimized copy from one rune away to another.
		for i := 0; i < b.N; i++ {
			var runes []rune
			for _, r := range []rune(noStyle) {
				runes = append(runes, r)
			}
		}
	})
}

func BenchmarkApply(b *testing.B) {
	b.Run("NoStyle", func(b *testing.B) {
		m := instyle.NewStyler()
		for i := 0; i < b.N; i++ {
			_ = m.Apply([]rune(noStyle))
		}
	})

	b.Run("NoStyleToFromString", func(b *testing.B) {
		m := instyle.NewStyler()
		for i := 0; i < b.N; i++ {
			_ = string(m.Apply([]rune(noStyle)))
		}
	})

	b.Run("WithStyle", func(b *testing.B) {
		m := instyle.NewStyler()
		withStyleRunes := []rune(withStyle)
		for i := 0; i < b.N; i++ {
			_ = m.Apply(withStyleRunes)
		}
	})

	b.Run("WithStyleToFromString", func(b *testing.B) {
		m := instyle.NewStyler()
		for i := 0; i < b.N; i++ {
			_ = string(m.Apply([]rune(withStyle)))
		}
	})
}

func BenchmarkApply_ColourStrings(b *testing.B) {
	b.Run("SingleHex", func(b *testing.B) {
		m := instyle.NewStyler()
		for i := 0; i < b.N; i++ {
			_ = m.ApplyStr("[~#028466]colour[/]")
		}
	})

	b.Run("MultipleHex", func(b *testing.B) {
		m := instyle.NewStyler()
		for i := 0; i < b.N; i++ {
			_ = m.ApplyStr("[~#028466]colour1[/] [~#d19645]colour2[/] [~#796ceb]colour3[/]")
		}
	})

	b.Run("SingleRGB", func(b *testing.B) {
		m := instyle.NewStyler()
		for i := 0; i < b.N; i++ {
			_ = string(m.Apply([]rune("[~rgb(53,170,3)]colour[/]")))
		}
	})
}

func TestStyleSet_Apply(t *testing.T) {
	tests := map[string]struct {
		In       string
		Expected string
	}{
		"Simple": {
			In:       "[~bold]bolded text[/]",
			Expected: "\033[0m\033[1mbolded text\033[0m",
		},
		"SequentialTags": {
			In:       "[~red]one[/] [~blue]two[/] [~black]three[/]",
			Expected: "\033[0m\033[31mone\033[0m \033[34mtwo\033[0m \033[30mthree\033[0m",
		},
		"NestedTags": {
			In:       "[~italic]this text is [~bold]bold [~red]red[/]-ish[/] and italic[/]",
			Expected: "\033[0m\033[3mthis text is \033[1mbold \033[31mred\033[0m\033[3m\033[1m-ish\033[0m\033[3m and italic\033[0m",
		},
		"UnclosedTags": {
			In:       "[~bold]bold and [~red]red also",
			Expected: "\033[0m\033[1mbold and \033[31mred also\033[0m",
		},
		"DeDuplicateEndResetTags": {
			In:       "[~bold]nested and [~red]red[/]",
			Expected: "\033[0m\033[1mnested and \033[31mred\033[0m",
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) {
			if result := instyle.NewStyler().Apply([]rune(tc.In)); string(result) != tc.Expected {
				t.Logf("Want: %+v", tc.Expected)
				t.Logf("Got:  %+v", string(result))
				t.FailNow()
			}
		})
	}
}
