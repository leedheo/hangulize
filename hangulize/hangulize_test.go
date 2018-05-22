package hangulize

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The most basic case of Hangulize.
func TestItaGloria(t *testing.T) {
	gloria := Hangulize("ita", "gloria")
	assert.Equal(t, "글로리아", gloria)
}

func TestItaGloriaFromSource(t *testing.T) {
	r := strings.NewReader(`
lang:
    id      = "ita"
    codes   = "it", "ita"
    english = "Italian"
    korean  = "이탈리아어"
    script  = "roman"

config:
    author  = "Brian Jongseong Park"
    stage   = "beta"
    markers = ",", ";"

vars:
    # 모음
    "@" = "a", "e", "i", "o", "u"

rewrite:
    # 의도 의도
    "'" -> ""

    "^gli$"     -> "li"
    "^glia$"    -> "g.lia"
    "^glioma$"  -> "g.lioma"
    "^gli{@}"   -> "li"
    "{@}gli"    -> "li"
    "gn{@}"     -> "nJ"
    "gn"        -> "n"
    "gg"        -> "g"
    "gi{a|o|u}" -> "G"
    "g{e|i}"    -> "G"

    "ss"        -> "s"
    "sce"       -> "sJe"
    "sci"       -> "si"

    "cc"        -> "c"
    "ci{a|o|u}" -> "z"
    "c{e|i}"    -> "z"
    "c"         -> "k"

    "qq"        -> "q"
    "quo"       -> "kuo"
    "qu{a|e|i}" -> "kW"
    "quy"       -> "kWi"
    "q"         -> "k"

    "xx" -> "x"
    "x"  -> "ks"
    "tt" -> "t"
    "ts" -> "z"
    "w"  -> "v"

    "aa" -> "a"
    "bb" -> "b"
    "dd" -> "d"
    "ee" -> "e"
    "ff" -> "f"
    "hh" -> "h"

    "{@}h{@}" -> "."
    "h"       -> ""

    "ii"      -> "i"
    "jj"      -> "j"
    "^j{@}"   -> "J"
    "{@}j{@}" -> "J"
    "j"       -> "i"

    "kk" -> "k"
    "ll" -> "l"

    "{@}mm{@}" -> "m,m"
    "mm"       -> "m"
    "{@}nn{@}" -> "n,n"

    "nn"    -> "n"
    "oo"    -> "o"
    "pp"    -> "p"
    "rr"    -> "r"
    "tt"    -> "t"
    "uu"    -> "u"
    "vv"    -> "v"
    "^y{@}" -> "J"
    "yy"    -> "y"
    "y"     -> "i"
    "zz"    -> "z"

    "{@}k{p|s|t|z}" -> "k,"
    "{@}p{k|s|t|z}" -> "p,"

    "^l" -> "l;"
    "^m" -> "m;"
    "^n" -> "n;"

    "l$" -> "l,"
    "m$" -> "m,"
    "n$" -> "n,"

    "l{@|m,|n,}" -> "l;"
    "{,}l"       -> "l;"
    "m{@}"       -> "m;"
    "n{@|J}"     -> "n;"

    "l" -> "l,"
    "m" -> "m,"
    "n" -> "n,"

    ",," -> ","
    ",;" -> ""

    ",l," -> "l,"
    ",m," -> "m,"
    ",n," -> "n,"

    "l{mn}" -> "l,"

    ";" -> ""
    
    # 예시(원본에 없었음)
    "a" -> "aa", "A" #  표기로 분기

hangulize:
    "b"    -> "ㅂ"
    "d"    -> "ㄷ"
    "f"    -> "ㅍ"
    "g"    -> "ㄱ"
    "G"    -> "ㅈ"
    "k,"   -> "-ㄱ"
    "k"    -> "ㄱ"
    "^l"   -> "ㄹ"
    "{,}l" -> "ㄹ"
    "l,"   -> "-ㄹ"
    "l"    -> "-ㄹㄹ"
    "m,"   -> "-ㅁ"
    "m"    -> "ㅁ"
    "n,"   -> "-ㄴ"
    "n"    -> "ㄴ"
    "p,"   -> "-ㅂ"
    "p"    -> "ㅍ"
    "r"    -> "ㄹ"
    "s"    -> "ㅅ"
    "t"    -> "ㅌ"
    "v"    -> "ㅂ"
    "z"    -> "ㅊ"
    "Ja"   -> "ㅑ"
    "Je"   -> "ㅖ"
    "Ji"   -> "ㅣ"
    "Jo"   -> "ㅛ"
    "Ju"   -> "ㅠ"
    "Wa"   -> "ㅘ"
    "We"   -> "ㅞ"
    "Wi"   -> "ㅟ"
    "a"    -> "ㅏ"
    "e"    -> "ㅔ"
    "i"    -> "ㅣ"
    "o"    -> "ㅗ"
    "u"    -> "ㅜ"

test:
    # 제1항 "gl"
    # "i" 앞에서는 "ㄹㄹ"로 적고, 그 밖의 경우에는 "글ㄹ"로 적는다.
    "paglia" -> "팔리아"
    "egli"   -> "엘리"
    "gloria" -> "글로리아"
    "glossa" -> "글로사"

    # 제2항 "gn"
    # 뒤따르는 모음과 합쳐 "냐", "녜", "뇨", "뉴", "니"로 적는다.
    "montagna" -> "몬타냐"
    "gneiss"   -> "녜이스"
    "gnocco"   -> "뇨코"
    "gn"       -> "뉴"
    "ogni"     -> "오니"
	`)
	spec, err := ParseSpec(r)
	if err != nil {
		panic(err)
	}

	hangulizer := NewHangulizer(spec)

	gloria := hangulizer.Hangulize("gloria")
	assert.Equal(t, "글로리아", gloria)
}
