package hangulize

import (
	"fmt"
	"regexp"
	"strings"
)

// Pattern is a domain-specific regular expression dialect for Hangulize.
type Pattern struct {
	expr string

	re  *regexp.Regexp // positive regexp
	neg *regexp.Regexp // negative regexp
}

func (p *Pattern) String() string {
	return fmt.Sprintf("/%s/", p.expr)
}

// Match reports whether the pattern matches the given text.
func (p *Pattern) Match(text string) []int {
	offset := 0

	for {
		loc := p.re.FindStringSubmatchIndex(text[offset:])

		// Not matched.
		if len(loc) == 0 {
			return make([]int, 0)
		}

		// p.re looks like (edge)(look)abc(look)(edge).
		// Hold only non-zero-width matches.
		start := offset + loc[5]
		stop := offset + loc[len(loc)-4]

		// Pick matched text.  Call it "highlight".
		highlight := text[loc[0]:loc[1]]

		// Test highlight with p.neg to determine whether skip or not.
		negLoc := p.neg.FindStringIndex(highlight)

		// If no negative match, this match has been succeeded.
		if len(negLoc) == 0 {
			return []int{start, stop}
		}

		// Shift the cursor.
		offset = loc[0] + negLoc[1]
	}
}

// CompilePattern compiles an Pattern pattern for the given language spec.
func CompilePattern(expr string, spec *Spec) *Pattern {
	reExpr := expr

	// spec dependent
	reExpr = expandMacros(reExpr, spec)
	reExpr = expandVars(reExpr, spec)

	// lookaround
	// TODO(sublee): edge specialization
	reExpr, negExpr := expandLookbehind(reExpr)
	reExpr, negExpr = expandLookahead(reExpr, negExpr)

	mustNoZeroWidth(reExpr)

	if negExpr == "" {
		negExpr = "$^" // It's a paradox.  Never matches with anything.
	}

	// Compile regexp.
	re := regexp.MustCompile(reExpr)
	neg := regexp.MustCompile(negExpr)
	return &Pattern{expr, re, neg}
}

var (
	reVar        *regexp.Regexp
	reLookbehind *regexp.Regexp
	reLookahead  *regexp.Regexp
	reZeroWidth  *regexp.Regexp
)

func init() {
	reVar = regexp.MustCompile(`<.+?>`)

	var (
		zeroWidth = `\{(.*?)\}` // {...}

		leftEdge  = `^(\^*)` // "^", "^^", or empty start
		rightEdge = `(\$*)$` // "$", "$$", or empty end

		lookbehind = fmt.Sprintf("%s(?:%s)?", leftEdge, zeroWidth)
		lookahead  = fmt.Sprintf("(?:%s)?%s", zeroWidth, rightEdge)
	)

	reLookbehind = regexp.MustCompile(lookbehind)
	reLookahead = regexp.MustCompile(lookahead)
	reZeroWidth = regexp.MustCompile(zeroWidth)
}

// expandMacros replaces macro sources in the spec to corresponding
// destinations.  It must be evaluated at the first in CompilePattern.
func expandMacros(reExpr string, spec *Spec) string {
	if spec == nil {
		return reExpr
	}

	args := make([]string, len(spec.Macros)*2)

	i := 0
	for src, dst := range spec.Macros {
		args[i] = src
		i++
		args[i] = dst
		i++
	}

	replacer := strings.NewReplacer(args...)
	return replacer.Replace(reExpr)
}

// expandVars replaces <var> to corresponding content regexp such as (a|b|c).
func expandVars(reExpr string, spec *Spec) string {
	if spec == nil {
		return reVar.ReplaceAllString(reExpr, "()")
	}

	return reVar.ReplaceAllStringFunc(reExpr, func(matched string) string {
		// Retrieve var from spec.
		varName := strings.Trim(matched, "<>")
		varVals := (*spec).Vars[varName]

		// Build as RegExp like /(a|b|c)/
		escapedVals := make([]string, len(varVals))
		for i, val := range varVals {
			escapedVals[i] = regexp.QuoteMeta(val)
		}
		return "(" + strings.Join(escapedVals, "|") + ")"
	})
}

// Lookbehind: {...} on the left-side.
func expandLookbehind(reExpr string) (string, string) {
	posExpr := reExpr
	negExpr := ""

	// ^{han}gul
	// │  │   └─ other
	// │  └─ look
	// └─ edge

	// This pattern always matches.
	//  [start, stop, edgeStart, edgeStop, lookStart, lookStop]
	loc := reLookbehind.FindStringSubmatchIndex(posExpr)

	edgeExpr := posExpr[loc[2]:loc[3]]
	otherExpr := posExpr[loc[1]:]

	// edgeExpr is one of "", "^", or "^^...".
	switch edgeExpr {
	case "":
		break
	case "^":
		edgeExpr = `^|\s+`
		break
	default:
		edgeExpr = "^"
	}

	lookExpr := ""
	if loc[5]-loc[4] != 0 {
		lookExpr = posExpr[loc[4]:loc[5]]

		if strings.HasPrefix(lookExpr, "~") {
			// negative lookbehind
			negExpr = fmt.Sprintf(`(%s)%s`, lookExpr[1:], otherExpr)

			lookExpr = ".*" // require greedy matching
		}
	}

	// Replace lookbehind with 2 parentheses:
	//  (^)(han)gul
	posExpr = fmt.Sprintf(`(%s)(%s)%s`, edgeExpr, lookExpr, otherExpr)

	return posExpr, negExpr
}

// Lookahead: {...} on the right-side.
// negExpr should be passed from expandLookbehind.
func expandLookahead(reExpr string, negExpr string) (string, string) {
	posExpr := reExpr

	// han{gul}$
	//  │   │  └─ edge
	//  │   └─ look
	//  └─ other

	// This pattern always matches:
	//  [start, stop, edgeStart, edgeStop, lookStart, lookStop]
	loc := reLookahead.FindStringSubmatchIndex(posExpr)

	otherExpr := posExpr[:loc[0]]
	edgeExpr := posExpr[loc[4]:loc[5]]

	// edgeExpr is one of "", "$", or "$$...".
	switch edgeExpr {
	case "":
		break
	case "$":
		edgeExpr = `$|\s+`
		break
	default:
		edgeExpr = "$"
	}

	lookExpr := ""
	if loc[3]-loc[2] != 0 {
		lookExpr = posExpr[loc[2]:loc[3]]

		// Lookahead can be remaining in the negative regexp
		// the lookbehind determined.
		if negExpr != "" {
			lookaheadLen := len(posExpr) - loc[0]
			negExpr = negExpr[:len(negExpr)-lookaheadLen]
		}

		if strings.HasPrefix(lookExpr, "~") {
			// negative lookahead
			if negExpr != "" {
				negExpr += "|"
			}
			negExpr += fmt.Sprintf(`^%s(%s)`, otherExpr, lookExpr[1:])

			lookExpr = ".*" // require greedy matching
		}
	}

	// Replace lookahead with 2 parentheses:
	//  han(gul)($)
	posExpr = fmt.Sprintf(`%s(%s)(%s)`, otherExpr, lookExpr, edgeExpr)

	return posExpr, negExpr
}

func mustNoZeroWidth(reExpr string) {
	if reZeroWidth.MatchString(reExpr) {
		panic(fmt.Errorf("zero-width group found in middle: %#v", reExpr))
	}
}