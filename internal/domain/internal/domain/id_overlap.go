package domain

import "strings"

const (
	digits    = "0123456789"
	nameChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" + digits + "_.-"
)

type idToken struct {
	chars string
	star  bool
}

func (checks domainChecks) checkTemplateOverlaps() error {
	sets := checks.compiled.Spec.Instances
	return checks.each("instances", len(sets), func(at domainChecks, j int) error {
		for i := range j {
			if sets[i].IDTemplate != "" && sets[j].IDTemplate != "" && templatesOverlap(sets[i].IDTemplate, sets[j].IDTemplate) {
				return at.bad("/id_template", "id_template %q can render an id that id_template %q renders too", sets[j].IDTemplate, sets[i].IDTemplate)
			}
		}
		return nil
	})
}

func templatesOverlap(a, b string) bool {
	product := idProduct{a: templateTokens(a), b: templateTokens(b), alphabet: nameChars + a + b, seen: map[[2]int]bool{}}
	for queue := pairs(closure(product.a, 0), closure(product.b, 0)); len(queue) > 0; queue = queue[1:] {
		if product.accepts(queue[0]) {
			return true
		}
		queue = append(queue, product.next(queue[0])...)
	}
	return false
}

type idProduct struct {
	a, b     []idToken
	alphabet string
	seen     map[[2]int]bool
}

func (p idProduct) accepts(at [2]int) bool {
	return at[0] == len(p.a) && at[1] == len(p.b)
}

func (p idProduct) next(at [2]int) [][2]int {
	if p.seen[at] {
		return nil
	}
	p.seen[at] = true
	var out [][2]int
	for _, c := range p.alphabet {
		out = append(out, pairs(advance(p.a, at[0], c), advance(p.b, at[1], c))...)
	}
	return out
}

func templateTokens(template string) []idToken {
	literals := templatePlaceholder.Split(template, -1)
	placeholders := templatePlaceholder.FindAllString(template, -1)
	var tokens []idToken
	for i, literal := range literals {
		for _, c := range literal {
			tokens = append(tokens, idToken{chars: string(c)})
		}
		if i < len(placeholders) {
			class := placeholderChars(placeholders[i])
			tokens = append(tokens, idToken{chars: class}, idToken{chars: class, star: true})
		}
	}
	return tokens
}

func placeholderChars(placeholder string) string {
	if placeholder == "{n}" {
		return digits
	}
	return nameChars
}

func closure(tokens []idToken, p int) []int {
	states := []int{p}
	for p < len(tokens) && tokens[p].star {
		p++
		states = append(states, p)
	}
	return states
}

func advance(tokens []idToken, p int, c rune) []int {
	if p == len(tokens) || !strings.ContainsRune(tokens[p].chars, c) {
		return nil
	}
	if tokens[p].star {
		return closure(tokens, p)
	}
	return closure(tokens, p+1)
}

func pairs(as, bs []int) [][2]int {
	out := make([][2]int, 0, len(as)*len(bs))
	for _, a := range as {
		for _, b := range bs {
			out = append(out, [2]int{a, b})
		}
	}
	return out
}
