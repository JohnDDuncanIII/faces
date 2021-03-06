package faces

import (
	//b64 "encoding/base64"
	"math"
	"regexp"
	"strings"
	"strconv"
)

/* python: https://www.apt-browse.org/browse/ubuntu/trusty/universe/all/xpn/1.2.6-5/file/usr/share/xpn/xpn_src/XFace.py
javascript (png): https://github.com/JohnDDuncanIII/MessageFaces/blob/master/chrome/content/xface.js
javascript (bmp): https://bug20417.bmoattachments.org/attachment.cgi?id=184433 */

var LENGTH = 48
var PIXELS = (LENGTH * LENGTH)

var FIRSTPRINT = []rune("!")[0]
var LASTPRINT  = []rune("~")[0]
var NUMPRINTS = (LASTPRINT - FIRSTPRINT + 1)

var BITSPERWORD = 8
var WORDCARRY = (1 << uint(BITSPERWORD))
var WORDMASK = (WORDCARRY - 1)
var MAXWORDS = math.Floor(float64(((PIXELS * 2 + BITSPERWORD - 1) / BITSPERWORD)))

var BLACK = 0
var GREY = 1
var WHITE = 2

type BType struct {
	b_first int
	b_words int
	b_word []int
}

type Levels struct {
	p_offset int
	p_range int
}

type Gen struct {
	g_00 []int
	g_01 []int
	g_02 []int
	g_10 []int
	g_20 []int
	g_30 []int
	g_40 []int
	g_11 []int
	g_21 []int
	g_31 []int
	g_41 []int
	g_12 []int
	g_22 []int
	g_32 []int
	g_42 []int
}

var B BType = BType{b_first: 0, b_words: 0, b_word: make([]int, int(MAXWORDS))}
var F = make([]int, PIXELS)

var levels = [][]Levels {
	[]Levels {Levels{p_offset: 255, p_range:1}, Levels{p_offset: 0, p_range:251}, Levels{p_offset:251, p_range:4 }},
	[]Levels {Levels{p_offset: 255, p_range:1}, Levels{p_offset: 0, p_range:200}, Levels{p_offset:200, p_range:55 }},
	[]Levels {Levels{p_offset: 223, p_range:33}, Levels{p_offset: 0, p_range:159}, Levels{p_offset:159, p_range:64 }},
	[]Levels {Levels{p_offset: 0, p_range:131}, Levels{p_offset: 0, p_range:0}, Levels{p_offset:131, p_range:125 }},
}

/* At the bottom of the octree 2x2 elements are considered black if any
 * pixel is black.  The probabilities below give the distribution of the
 * 16 possible 2x2 patterns.  All white is not really a possibility and
 * has a probability range of zero.  Again, experimentally derived data */
var freqs = []Levels {
	Levels{p_offset:0,   p_range:0 }, Levels{p_offset:0,   p_range:38}, Levels{p_offset:38,  p_range:38}, Levels{p_offset:152, p_range:13},
	Levels{p_offset:76,  p_range:38}, Levels{p_offset:165, p_range:13}, Levels{p_offset:178, p_range:13}, {p_offset:230, p_range:6 },
	Levels{p_offset:114, p_range:38}, Levels{p_offset:191, p_range:13}, Levels{p_offset:204, p_range:13}, {p_offset:236, p_range:6 },
	Levels{p_offset:217, p_range:13}, Levels{p_offset:242, p_range:6 }, Levels{p_offset:248, p_range:5 }, Levels{p_offset:253, p_range:3 },
}

// table for GenFunc()
var G = Gen{
	g_00: []int {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
		1,1,1,0,0,0,1,1,1,1,0,1,1,1,1,1,0,0,0,0,0,1,0,1,0,0,0,1,0,1,1,1,0,0,0,0,0,1,0,1,0,0,0,0,1,1,1,1,
		0,0,0,0,0,0,0,0,0,0,0,1,1,0,1,1,0,0,0,0,1,1,1,1,1,1,0,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,1,0,0,0,0,1,1,1,1,0,0,0,0,0,0,1,1,0,1,1,1,1,1,1,1,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,1,1,1,0,1,
		0,1,0,0,0,1,0,1,0,0,1,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,1,
		0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,
		0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,1,0,0,1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,
		0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,1,1,0,0,1,0,0,1,0,0,0,0,1,1,1,1,0,0,0,0,1,1,1,1,
		1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
		0,0,0,1,1,0,1,1,0,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,0,0,1,1,1,1,0,1,0,1,0,1,0,0,
		0,0,0,0,0,1,1,1,0,0,0,1,1,1,1,1,0,1,0,1,0,1,1,1,0,1,0,0,0,1,1,1,1,1,0,1,0,1,1,1,0,0,1,1,1,1,0,1,
		1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,0,1,1,1,1,1,0,0,0,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
		0,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,0,0,0,0,0,1,0,1,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,1,0,0,0,0,1,1,1,1,
		0,0,0,0,1,1,1,1,0,1,0,1,1,1,1,1,1,0,0,1,1,0,1,1,1,1,0,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
		0,1,0,1,1,1,1,1,0,0,0,1,1,1,0,1,0,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,0,0,0,1,1,1,1,1,
		0,0,0,0,1,1,1,1,0,1,0,1,1,1,1,1,0,0,0,0,0,0,1,1,0,0,0,1,1,1,1,1,0,1,0,0,1,1,1,1,0,1,0,1,1,1,1,1,
		1,1,1,1,0,1,1,1,0,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,1,1,0,1,0,0,0,0,1,1,1,1,
		1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,0,1,1,1,1,1,1,0,0,0,0,1,1,1,1,0,1,0,0,1,1,1,1,
		1,1,0,1,0,1,1,1,0,0,1,1,1,1,1,1,0,1,0,0,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
		0,1,1,0,0,1,1,1,1,0,1,1,1,1,1,1,0,1,0,1,0,1,1,0,0,0,1,0,0,1,0,1,0,0,0,1,1,1,1,1,0,1,1,1,1,1,1,1,
		1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,
		0,1,0,1,1,1,1,1,0,1,1,1,1,1,1,1,0,0,0,0,0,0,0,1,1,1,0,1,1,1,1,1,0,0,0,1,0,1,0,0,0,0,0,0,0,0,0,0,
		0,0,0,0,0,1,0,1,0,0,0,0,1,1,1,1,0,0,0,0,0,1,1,1,1,0,1,0,0,0,1,0,0,0,0,0,1,0,0,1,0,0,0,0,1,1,1,1,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,0,1,0,1,1,1,1,1,
		0,0,0,1,1,0,0,0,1,1,0,1,0,1,1,1,1,0,0,1,0,1,0,0,0,1,1,1,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,
		0,0,0,1,1,1,1,1,1,0,1,1,0,1,1,1,0,0,0,0,1,1,0,0,0,0,0,0,0,1,1,1,0,0,0,0,1,1,1,1,0,0,0,0,1,1,1,1,
		0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,0,0,0,0,1,1,1,1,0,0,0,1,1,1,1,1,1,0,0,0,0,1,0,0,1,0,0,0,1,1,1,1,
		0,0,0,0,0,1,0,1,0,0,0,1,0,1,0,1,0,0,0,0,0,1,0,1,0,0,0,0,1,1,1,1,0,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,
		1,0,0,0,0,1,1,1,1,1,0,1,1,1,1,1,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,1,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,
		0,0,0,0,1,1,1,1,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,1,0,1,0,0,0,0,0,1,0,0,
		0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,1,0,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,0,0,0,1,1,1,1,
		0,1,0,0,1,0,1,0,0,1,0,0,0,0,0,0,0,1,0,1,1,1,1,1,0,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,
		1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,
		1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,1,1,1,1,1,0,1,
		1,1,0,1,0,1,1,1,0,1,0,1,1,1,1,1,0,1,0,0,1,1,1,1,0,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,
		1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,0,1,1,1,
		1,1,0,1,1,1,1,1,0,1,1,1,1,1,1,1,0,1,0,0,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
		0,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,
		0,0,0,0,1,1,1,1,0,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,1,1,1,0,1,1,1,1,1,1,1,1,1,
		0,0,0,0,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,0,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,
		1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,1,1,0,1,0,0,0,0,1,1,1,1,
		0,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,1,0,0,0,0,0,1,0,1,0,0,0,0,0,0,1,0,0,0,0,0,0,0,1,0,0,0,0,0,1,1,1,1,
		0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,1,1,0,
		0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,0,0,1,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
		0,0,0,0,0,1,0,1,0,0,0,0,1,1,1,1,0,1,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,1,1,0,0,0,0,0,0,1,1,1,1,
		0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
		1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,1,0,1,
		0,0,0,0,0,0,0,1,0,0,0,1,0,1,0,1,1,0,1,0,1,1,1,1,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,
		0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,1,1,0,0,0,0,0,1,1,0,0,
		0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,0,0,0,1,0,1,0,1,
		1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,
		0,1,1,1,1,1,1,1,0,1,0,1,1,1,1,1,1,1,0,1,1,0,1,1,1,1,1,1,1,1,1,1,0,1,0,0,1,1,1,1,0,0,1,1,1,1,1,0,
		0,0,0,0,0,1,0,1,0,0,0,0,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,0,0,1,0,1,0,1,0,1,0,0,1,1,1,1,
		0,0,0,0,1,1,0,1,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,1,0,0,0,0,1,1,1,1,0,1,0,0,1,1,1,1,0,1,0,1,1,1,1,1,
		1,0,0,1,1,1,1,1,1,1,0,1,1,1,1,1,0,0,1,0,0,1,0,1,0,0,0,0,1,1,1,0,0,0,0,0,1,1,0,1,0,0,0,0,1,1,0,1,
		0,1,0,0,1,1,1,1,0,1,1,1,1,1,1,1,1,0,0,0,1,1,1,1,0,0,0,0,1,1,1,1,0,0,0,0,1,1,1,1,1,1,1,1,1,0,1,0,
		0,0,0,0,0,1,0,0,0,1,0,0,1,1,1,1,0,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,0,1,1,1,0,1,1,1,
		0,1,0,0,0,1,1,1,1,1,1,0,1,1,0,1,0,0,0,0,0,1,0,1,0,0,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
		1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,0,0,1,1,1,1,0,1,1,0,1,1,1,1,1,1,0,1,1,0,0,0,0,1,0,1,1,1,1,1,
		0,0,0,0,1,1,1,1,0,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,0,1,0,1,1,1,1,1,0,0,0,0,0,1,1,1,0,0,0,0,1,1,1,1,
		1,0,0,1,0,1,0,0,0,0,0,0,1,1,0,1,0,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,1,0,0,0,1,1,0,0,1,0,1,0,1,1,1,
		0,0,0,0,0,0,0,1,0,0,0,0,1,1,0,1,0,0,0,0,0,0,0,1,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,1,1,1,1,
		0,1,0,0,0,1,1,1,0,1,1,0,1,1,0,0,0,0,0,0,1,1,0,1,0,0,0,0,1,1,1,1,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,1,0,1,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,
		0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,0,1,0,0,1,0,1,0,1,0,0,0,0,0,0,0,1,0,0,0,0,1,1,1,1,0,1,1,1,1,1,1,1,
		0,0,0,0,1,1,0,0,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,1,0,0,0,0,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
		0,0,0,0,1,1,1,1,0,1,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,1,0,0,1,0,0,
		0,0,0,0,1,1,0,1,0,0,0,0,1,1,1,1,0,0,0,0,1,1,1,1,0,1,1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,0,1,1,1,1,1,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,1,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,0,1,0,0,1,1,0,1,1,0,0,1,1,1,1,0,0,0,0,0,1,0,1,
		1,1,0,0,1,1,1,1,0,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,0,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
		0,0,0,1,0,1,1,1,0,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
		0,1,0,0,0,1,1,0,0,0,0,0,1,0,0,1,0,1,0,0,1,1,1,1,0,1,0,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,
		1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,1,0,1,0,1,0,0,0,1,0,0,0,1,0,1,0,0,1,1,1,0,1,1,1,1,1,1,1,
		0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,0,0,0,0,0,1,0,0,
		1,1,0,1,1,1,1,1,0,1,1,1,1,1,1,1,0,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,
		0,0,0,0,1,1,1,0,1,1,1,0,0,1,1,0,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
		1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,1,1,0,1,1,0,0,1,0,0,0,1,1,1,1,0,1,0,0,1,1,1,1,
		0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,1,0,0,1,1,1,1,
		1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,1,0,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
		0,0,0,0,0,0,1,1,0,0,0,0,1,1,0,0,1,0,0,1,1,1,0,1,0,0,0,0,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,
		1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1},
	g_01: []int {0,0,1,1,0,1,1,1,0,1,1,1,0,0,1,1,0,0,0,0,0,0,0,0,0,0,0,1,1,0,0,1,0,1,0,1,0,1,1,1,0,1,1,1,1,1,1,1,
		1,1,1,1,0,1,0,1,1,1,1,1,1,0,1,1,0,1,1,1,0,0,0,0,0,0,1,1,0,0,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,0,0,1,
		0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1},
	g_02: []int {0,1,0,1},
	g_10: []int {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,
		1,1,1,1,0,0,1,1,0,1,0,1,1,1,1,1,1,0,0,0,0,1,0,0,0,0,0,0,0,1,0,0,0,0,0,1,0,1,1,1,1,0,0,1,1,1,1,1,
		0,0,0,0,0,1,0,0,0,0,1,0,0,0,1,1,0,0,0,0,0,1,0,1,1,1,1,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,1,1,0,0,0,0,0,0,1,1,0,0,1,1,0,0,1,1,1,1,0,1,0,1,1,1,
		0,0,0,0,0,1,0,1,0,0,0,0,0,0,1,1,0,1,0,1,1,1,1,1,0,0,1,1,1,1,1,1,0,0,0,1,0,1,1,1,0,0,1,1,0,0,1,1,
		1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,1,0,0,
		0,0,0,1,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,1,0,1,0,1,0,1,1,1,0,0,0,0,0,1,0,1,0,0,1,0,0,1,0,1,
		0,0,0,0,0,1,0,1,0,0,0,0,0,0,1,1,0,0,1,1,0,1,0,1,1,0,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,
		0,0,0,0,0,1,1,1,0,1,1,0,1,1,1,1,0,0,1,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,1,0,1,1,1,0,0,0,0,0,1,1,0,
		1,1,1,1,1,0,1,0,1,1,1,0,1,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,1,1,1,0,0,0,1,1,1,1,1,1,0,0,1,1,1,1,1,
		0,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1},
	g_20: []int {0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,1,0,1,0,0,0,0,1,1,0,0,1,0,1,1,1,0,
		1,1,1,1,1,1,1,1,0,0,1,1,1,1,1,1},
	g_30: []int {0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,1,0,1,0,0,0,1,0,0,0,1,0,0,0,1,
		0,0,0,1,0,0,1,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,1,1,0,0,0,1,0,0,0,1,
		0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,1,1,0,0,1,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,1,1,0,0,0,1,0,0,0,1,
		0,0,0,1,0,0,1,1,0,0,0,1,0,0,1,1,0,0,0,1,0,0,1,1,0,0,0,1,0,0,1,1,0,0,1,1,0,0,0,1,0,0,1,1,0,0,0,1,
		0,0,0,1,0,0,0,1,0,0,0,0,0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,0,0,1,0,1,1,1,0,0,0,1,0,0,0,1,0,0,0,1,
		0,0,0,1,0,0,0,1,0,1,1,1,0,1,0,1},
	g_40: []int {0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,1,
		0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,1,0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,1,0,0,1,1,1,0,
		1,1,1,0,0,1,0,0,0,0,0,0,1,1,0,1,0,0,0,1,0,0,0,0,0,0,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,
		0,1,0,0,0,1,0,0,0,1,0,0,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,0,0,0,0,0,1,1,1,1,0,0,0,0,1,1,1,1,
		1,0,1,0,1,1,1,0,1,0,1,0,1,1,1,1,0,1,0,0,0,1,0,1,0,1,1,1,1,1,1,1,1,1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,
		0,0,0,0,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,1,0,0,0,0,0,0,0,1,0,0,0,1,0,0,0,1,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,1,1,1,0,0,1,1,0,1,1,1,0,1,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,1,
		0,0,0,0,0,0,0,0,1,1,1,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,1,1,0,1,
		0,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,0,1,0,0,1,1,1,1,0,1,0,1,1,1,1,1,0,0,1,1,1,1,0,1,1,1,1,1,1,1,1,1,
		1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,1,1,1,0,0,1,1,1,1,1,1,1,1,
		1,1,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,0,1,1,0,1,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,1,0,1,0,0,0,0,0,0,0,1,0,0,0,0,0,1,1,1,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,1,0,0,0,0,1,1,1,1,1,0,0,0,0,0,0,0,1,0,0,0,1,0,0,0,1,
		0,0,0,0,0,1,0,1,0,1,1,1,1,1,1,1,0,0,0,0,0,0,0,0,0,0,0,1,1,1,1,1,0,1,0,0,0,0,0,1,0,1,0,1,0,1,1,1,
		0,0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,0,1,0,1,0,1,1,1,0,1,1,1,0,0,0,0,1,1,0,1,0,1,0,1,1,1,1,1,
		0,1,0,0,1,1,0,1,1,1,1,1,1,1,1,1,0,1,0,0,1,1,1,1,1,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,1,1,1,1,1,1,1,
		0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,1,
		0,0,0,0,0,1,0,1,0,1,1,1,1,1,0,1,0,0,0,1,0,0,0,0,0,0,0,1,0,1,0,1,0,0,1,0,1,1,1,1,1,1,1,1,1,1,1,1,
		0,1,0,0,0,0,0,0,0,1,0,1,0,0,0,0,0,0,0,0,1,1,0,1,1,1,1,1,1,1,0,1,0,0,0,0,0,1,0,0,0,0,0,0,1,1,1,1,
		0,0,0,0,0,1,1,1,0,0,0,1,1,1,1,1,0,0,0,0,0,1,1,1,0,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,0,1,1,1,1,1,1,
		0,0,0,0,1,1,0,1,0,1,1,1,1,1,1,1,0,0,0,0,1,1,1,1,1,1,1,1,1,1,1,1,0,1,0,0,1,1,0,1,0,1,1,1,1,1,0,1,
		0,0,0,0,1,1,1,1,1,1,1,1,1,1,1,1},
	g_11: []int {0,0,0,0,0,0,0,1,0,0,0,1,0,0,1,1,0,0,0,0,0,0,1,1,0,1,1,1,1,1,1,1},
	g_21: []int {0,0,0,1,0,1,1,1},
	g_31: []int {0,1,0,1,0,1,0,1,0,1,0,1,0,1,1,1,0,1,0,1,0,1,1,1,0,1,1,1,1,1,1,1},
	g_41: []int {0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,1,0,0,0,0,0,0,0,1,0,0,0,1,1,1,1,1,0,0,0,0,0,0,1,1,0,0,0,1,1,1,1,1,
		0,0,1,1,1,1,1,1,1,1,1,1,1,1,1,1},
	g_12: []int {0,1},
	g_22: []int {0},
	g_32: []int {0,0,0,1},
	g_42: []int {0,0,0,1},
}

func BigMul(a int) { // multiply B.b_word by a (B.b_word[0]=LSB)
	var i int
	if a == 1 || B.b_words == 0 {
		return
	}
	/* Treat this as a == WORDCARRY and just shift everything left a WORD */
	if a == 0 {
		B.b_words++
		if B.b_first > 0 {
			B.b_first--
		} else {
			for i = B.b_words - 1; i >= 1; i-- {
				B.b_word[i] = B.b_word[i - 1]
			}
		}
		B.b_word[B.b_first] = 0
		return
	}

	var c = 0
	var last = B.b_words + B.b_first - 1
	var word = B.b_word
	for i = B.b_first; i <= last; i++ {
		c += word[i] * a
		word[i] = (c & WORDMASK)
		c >>= uint(BITSPERWORD)
	}
	if c != 0 {
		word[B.b_first + B.b_words] = c & WORDMASK
		B.b_words++
	}
}

func BigAdd(a int) { // add a to B.b_word
	var i int
	a &= WORDMASK
	if a == 0 {
		return
	}

	var c = a
	var last = B.b_words + B.b_first - 1
	for i = B.b_first; i <= last; i++ {
		c += B.b_word[i]
		B.b_word[i] = c & WORDMASK
		c >>= uint(BITSPERWORD)
		if c == 0 {
			break
		}
	}
	if ((i > last) && c != 0) {
		if i >= len(B.b_word) { // i hate js
			B.b_word = append(B.b_word, c & WORDMASK)
		} else {
			B.b_word[i] = c & WORDMASK
		}

		B.b_words++
	}
}

func BigPop(p []Levels) int { // p is freqs oder levels[lev]
    var r = B.b_word[B.b_first] // r = LSB; B >> 8
    B.b_first++
    B.b_words--

    var i = 0
    for ((r < p[i].p_offset) || (r >= p[i].p_range + p[i].p_offset)) {
	i++
    }
    BigMul(p[i].p_range)
    BigAdd(r - p[i].p_offset)
    return i
}

func PopGreys(off int, len int) {
    if (len > 3) {
	len /= 2
	PopGreys(off,  len)
	PopGreys(off + len, len)
	PopGreys(off + LENGTH * len, len)
	PopGreys(off + LENGTH * len + len, len)
    } else {
	len = BigPop(freqs)
	if (len & 1) != 0 {
	    F[off] = 1
	}
	if (len & 2) != 0 {
	    F[off + 1] = 1
	}
	if (len & 4) != 0 {
	    F[off + LENGTH] = 1
	}
	if (len & 8) != 0 {
	    F[off + LENGTH + 1] = 1
	}
    }
}

func UnCompress(off int, len int, lev int) {
	switch (BigPop(levels[lev])) {
	case WHITE:
		return
	case BLACK:
		PopGreys(off, len)
		return
	default:
		len /= 2
		lev++
		UnCompress(off, len, lev)
		UnCompress(off + len, len, lev)
		UnCompress(off + len * LENGTH, len, lev)
		UnCompress(off + len * LENGTH + len, len, lev)
		return
	}
}

func UnCompAll(fbuf string) {
	var i int
	B.b_words = 0
	B.b_first = 0
	// convert base 94 to base 256
	var kl = len(fbuf)
	for i = 0; i < kl; i++ {
		BigMul(int(NUMPRINTS))
		cCode := []rune(fbuf)[i]
		BigAdd(int(cCode) - int(FIRSTPRINT))
	}
	// empty icon
	for i = 0; i < PIXELS; i++ {
		F[i] = 0
	}
	// uncompress
	UnCompress(0, 16, 0)
	UnCompress(16, 16, 0)
	UnCompress(32, 16, 0)
	UnCompress(768, 16, 0)
	UnCompress(784, 16, 0)
	UnCompress(800, 16, 0)
	UnCompress(1536, 16, 0)
	UnCompress(1552, 16, 0)
	UnCompress(1568, 16, 0)
}

func GenFunc() {
	var m, l, k, j, i, h int
	for j = 0; j < LENGTH;  j++ {
		for i = 0; i < LENGTH;  i++ {
			k = 0
			for l = i - 2; l <= i + 2; l++ {
				for m = j - 2; m <= j; m++ {
					if (l >= i) && (m == j) {
						continue
					}
					if (l > 0) && (l <= LENGTH) && (m > 0) {
						if F[l + m * LENGTH] != 0 {
							k = k * 2 + 1
						} else {
							k = k * 2
						}
					}
				}
			}
			switch (i) {
			case 1 :
				switch (j) {
				case 1 :
					F[h] ^= G.g_22[k]
					break
				case 2 :
					F[h] ^= G.g_21[k]
					break
				default:
					F[h] ^= G.g_20[k]
					break
				}
				break
			case 2 :
				switch (j) {
				case 1 :
					F[h] ^= G.g_12[k]
					break
				case 2 :
					F[h] ^= G.g_11[k]
					break
				default:
					F[h] ^= G.g_10[k]
					break
				}
				break
			case LENGTH - 1 :
				switch (j) {
				case 1 :
					F[h] ^= G.g_42[k]
					break
				case 2 :
					F[h] ^= G.g_41[k]
					break
				default:
					F[h] ^= G.g_40[k]
					break
				}
				break
			case LENGTH :
				switch (j) {
				case 1 :
					F[h] ^= G.g_32[k]
					break
				case 2 :
					F[h] ^= G.g_31[k]
					break
				default:
					F[h] ^= G.g_30[k]
					break
				}
				break
				default :
				switch (j) {
				case 1 :
					F[h] ^= G.g_02[k]
					break
				case 2 :
					F[h] ^= G.g_01[k]
					break
				default:
					F[h] ^= G.g_00[k]
					break
				}
				break
			}
			h++
		}
	}
}

// array of one byte strings, initialized to zero bytes
var png [2444]string
var crc32_table [256]int

func PNGFace() {
	var i int
	for i = 0; i < 2444; i++ {
		png[i] = "\x00"
	}

	Insert(0, "\x00\x00\x00\rIHDR\x00\x00\x000\x00\x00\x000\b\x03")
	Insert(25, "\x00\x00\x00\x06PLTE")
	Insert(43, "\x00\x00\x00\x02tRNS")
	Insert(57, "\x00\x00\t;IDATxÚ\x010\tÏö")
	Insert(2432, "\x00\x00\x00\x00IEND")

	/* Table of CRCs of all 8-bit messages. */
	var n int
	for n = 0; n < 256; n++ {
		var c = n
		var k int
		for k = 0; k < 8; k++ {
			if (c & 1) != 0 {
				c = -306674912 ^ ((c >> 1) & 0x7fffffff)
			} else {
				c = (c >> 1) & 0x7fffffff
			}
		}
		crc32_table[n] = c
	}
}

func Color(index int, asRGBA string) {
	var oRGBA = strings.Split(asRGBA,",")
	v0, _ := strconv.Atoi(oRGBA[0])
	v1, _ := strconv.Atoi(oRGBA[1])
	v2, _ := strconv.Atoi(oRGBA[2])
	v3, _ := strconv.Atoi(oRGBA[3])
	png[33+index*3+0] = string(v0)
	png[33+index*3+1] = string(v1)
	png[33+index*3+2] = string(v2)
	png[51+index] = string(v3 * 255)
}

func Insert(offs int, str string) {
	var j int
	for j = 0; j < len([]rune(str)); j++ {
		png[offs] = string([]rune(str)[j])
		offs++
	}
}

func Insert4(offs int, w int) {
	Insert(offs, (string((w>>24)&255) + string((w>>16)&255) + string((w>>8)&255) + string(w&255)))
}

func CRC32(offs int, size int) {
	var crc = -1 // initialize crc
	var i int
	for i = 4; i < size - 4; i++ {
		crc = crc32_table[(crc ^ int([]rune(png[offs + i])[0])) & 0xff] ^ ((crc >> 8) & 0x00ffffff)
	}
	Insert4(offs + size - 4, crc ^ -1)
}

func PNGFaceURL (xface []int) {
	// compute adler32 of output pixels + row filter bytes
	var BASE = 65521 /* largest prime smaller than 65536 */
	var NMAX = 5552  /* NMAX is the largest n such that 255n(n+1)/2 + (n+1)(BASE-1) <= 2^32-1 */
	var s1 = 1
	var s2 = 0
	var n = NMAX
	var y int
	for y = 0; y < 48; y++ {
		var x int
		for x = -1; x < 48; x++ {
			var i = y * 49 + x + 73
			if (x >= 0) {
				png[i] = string(xface[x + y * 48]) // set X-Face dot
			}
			if len([]rune(png[i])) > 0 { // custom
				s1 += int([]rune(png[i])[0])
			}
			s2 += s1
			n--
			if n == 0 {
				s1 %= BASE
				s2 %= BASE
				n = NMAX
			}
		}
	}

	s1 %= BASE
	s2 %= BASE

	Insert4(2424, (s2 << 16) | s1)

	CRC32(0, 25)
	CRC32(25, 18)
	CRC32(43, 14)
	CRC32(57, 2375)
	CRC32(2432, 12)
}

// output a PNG string, Base64 encoded (basically, javascript btoa())
func DoBase64 (s string) string {
	var ch = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	var c1, c2, c3 rune
	var e1, e2, e3, e4 uint
	var l = len([]rune(s))
	var r = ""
	var i int
	for i=0; i < l; i+=3 {
		c1 = []rune(s)[i]
		e1 = uint(c1) >> 2
		if i < l-1 {
			c2 = []rune(s)[i+1]
			e2 = ((uint(c1) & 3) << 4) | (uint(c2) >> 4)
		} else {
			e2 = ((uint(c1) & 3) << 4)
		}
		if i < l-1 {
			c3 = []rune(s)[i+2]
		}
		if (l < i+2) {
			e3 = 64
		} else {
			e3 = (uint(c2 & 0xf) << 2) | (uint(c3) >> 6)
		}
		if (l < i+3) {
			e4 = 64
		} else {
			e4 = uint(c3) & 0x3f
		}
		r+= string([]rune(ch)[e1]) + string([]rune(ch)[e2]) + string([]rune(ch)[e3]) + string([]rune(ch)[e4])
	}

	return r
}

func DoXFace(xface string) string {
	PNGFace()
	var re = regexp.MustCompile(`[^!-~]`)
	s := re.ReplaceAllString(xface, "")
	UnCompAll(s) // eliminate illegal chars
	GenFunc()

	Color(0, "255,255,255,1")
	Color(1, "0,0,0,1")

	PNGFaceURL(F)
	return DoBase64(string(0x89)+"PNG\r\n"+string(0x1a)+"\n"+strings.Join(png[:],""))
	//return "data:image/png;base64," +  b64.StdEncoding.EncodeToString([]byte(string(0x89)+"PNG\r\n"+string(0x1a)+"\n"+strings.Join(png[:],""))) // this is not working
}
