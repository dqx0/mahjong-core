package main

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Shoupai struct {
	bingpai map[string][]int
	fulou   []string
	zimo    string
	lizhi   bool
}

func validPai(p string) bool {
	match, _ := regexp.MatchString(`^(?:[mps]\d|z[1-7])_?\*?[\+\=\-]?$`, p)
	return match
}

func validMianzi(m string) string {
	if strings.Contains(m, "z") && strings.ContainsAny(m, "089") {
		return ""
	}
	h := strings.ReplaceAll(m, "0", "5")
	if match, _ := regexp.MatchString(`^[mpsz](\d)\d\d[\+\=\-]?\d?$`, h); match {
		if h[1] == h[2] && h[2] == h[3] {
			return strings.ReplaceAll(m, "05", "50")
		}
	} else if match, _ := regexp.MatchString(`^[mpsz](\d)\d\d\d[\+\=\-]?$`, h); match {
		digits := regexp.MustCompile(`\d`).FindAllString(h, -1)
		filteredDigits := []string{}
		for _, digit := range digits {
			if !strings.ContainsAny(digit, "+=-") {
				filteredDigits = append(filteredDigits, digit)
			}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(filteredDigits)))
		suffix := regexp.MustCompile(`\d[\+\=\-]$`).FindString(m)
		return string(m[0]) + strings.Join(filteredDigits, "") + suffix
	} else if match, _ := regexp.MatchString(`^[mps]\d+\-\d*$`, h); match {
		hongpai := strings.Contains(m, "0")
		nn := regexp.MustCompile(`\d`).FindAllString(h, -1)
		if len(nn) != 3 || (nn[0][0]+1 != nn[1][0]) || (nn[1][0]+1 != nn[2][0]) {
			return ""
		}
		h = h[0:1] + strings.Join(regexp.MustCompile(`\d[\+\=\-]?`).FindAllString(h, -1), "")
		if hongpai {
			return strings.Replace(h, "5", "0", 1)
		}
		return h
	}
	return ""
}

func NewShoupai() *Shoupai {
	s := &Shoupai{
		bingpai: map[string][]int{
			"_": {0},
			"m": make([]int, 10),
			"p": make([]int, 10),
			"s": make([]int, 10),
			"z": make([]int, 8),
		},
		fulou: []string{},
		zimo:  "",
		lizhi: false,
	}

	return s
}

func NewShoupaiWithQipai(qipai []string) (*Shoupai, error) {
	s := &Shoupai{
		bingpai: map[string][]int{
			"_": {0},
			"m": make([]int, 10),
			"p": make([]int, 10),
			"s": make([]int, 10),
			"z": make([]int, 8),
		},
		fulou: []string{},
		zimo:  "",
		lizhi: false,
	}

	for _, p := range qipai {
		if p == "_" {
			s.bingpai["_"][0]++
			continue
		}
		if !validPai(p) {
			return nil, errors.New("Invalid pai: " + p)
		}
		suit := string(p[0])
		n, _ := strconv.Atoi(string(p[1]))
		if s.bingpai[suit][n] == 4 {
			return nil, errors.New("Too many pai: " + p)
		}
		s.bingpai[suit][n]++
		if suit != "z" && n == 0 {
			s.bingpai[suit][5]++
		}
	}

	return s, nil
}

func FromString(paistr string) *Shoupai {
	fulou := strings.Split(paistr, ",")
	bingpai := fulou[0]
	fulou = fulou[1:]

	qipai := regexp.MustCompile(`_`).FindAllString(bingpai, -1)
	if qipai == nil {
		qipai = []string{}
	}
	for _, suitstr := range regexp.MustCompile(`[mpsz]\d+`).FindAllString(bingpai, -1) {
		s := string(suitstr[0])
		for _, n := range regexp.MustCompile(`\d`).FindAllString(suitstr, -1) {
			if s == "z" && (n < "1" || n > "7") {
				continue
			}
			qipai = append(qipai, s+n)
		}
	}
	// fulouスライスの中で真の値を持つ要素の数を数える
	count := 0
	for _, x := range fulou {
		if x != "" {
			count++
		}
	}

	// qipaiスライスの長さを調整
	newLength := 14 - count*3
	if newLength < 0 {
		newLength = 0
	}
	if newLength > len(qipai) {
		newLength = len(qipai)
	}
	qipai = qipai[:newLength]
	var zimo string
	if (len(qipai)-2)%3 == 0 && len(qipai) > 0 {
		zimo = qipai[len(qipai)-1]
		qipai = qipai[:len(qipai)-1]
	}
	shoupai, err := NewShoupaiWithQipai(qipai)
	if err != nil {
		return nil
	}

	var last string
	for _, m := range fulou {
		if m == "" {
			shoupai.zimo = last
			break
		}
		m = validMianzi(m)
		if m != "" {
			shoupai.fulou = append(shoupai.fulou, m)
			last = m
		}
	}

	if shoupai.zimo == "" {
		if zimo != "" {
			shoupai.zimo = zimo
		} else {
			shoupai.zimo = ""
		}
	}
	shoupai.lizhi = strings.HasSuffix(bingpai, "*")

	return shoupai
}

func (s *Shoupai) ToString() string {
	paistr := strings.Repeat("_", s.bingpai["_"][0])
	if s.zimo == "_" {
		paistr = strings.Repeat("_", s.bingpai["_"][0]-1)
	}

	for _, suit := range []string{"m", "p", "s", "z"} {
		suitstr := suit
		bingpai := s.bingpai[suit]
		nHongpai := 0
		if suit != "z" {
			nHongpai = bingpai[0]
		}
		for n := 1; n < len(bingpai); n++ {
			nPai := bingpai[n]
			if s.zimo != "" {
				if suit+fmt.Sprint(n) == s.zimo {
					nPai--
				}
				if n == 5 && suit+"0" == s.zimo {
					nPai--
					nHongpai--
				}
			}
			for i := 0; i < nPai; i++ {
				if n == 5 && nHongpai > 0 {
					suitstr += "0"
					nHongpai--
				} else {
					suitstr += fmt.Sprint(n)
				}
			}
		}
		if len(suitstr) > 1 {
			paistr += suitstr
		}
	}

	if s.zimo != "" && len(s.zimo) <= 2 {
		paistr += s.zimo
	}
	if s.lizhi {
		paistr += "*"
	}

	for _, m := range s.fulou {
		paistr += "," + m
	}
	if s.zimo != "" && len(s.zimo) > 2 {
		paistr += ","
	}

	return paistr
}

func (s *Shoupai) Clone() *Shoupai {
	clone := &Shoupai{
		bingpai: map[string][]int{
			"_": append([]int{}, s.bingpai["_"]...),
			"m": append([]int{}, s.bingpai["m"]...),
			"p": append([]int{}, s.bingpai["p"]...),
			"s": append([]int{}, s.bingpai["s"]...),
			"z": append([]int{}, s.bingpai["z"]...),
		},
		fulou: append([]string{}, s.fulou...),
		zimo:  s.zimo,
		lizhi: s.lizhi,
	}
	return clone
}

func (s *Shoupai) FromString(paistr string) *Shoupai {
	shoupai := FromString(paistr)

	s.bingpai = map[string][]int{
		"_": append([]int{}, shoupai.bingpai["_"]...),
		"m": append([]int{}, shoupai.bingpai["m"]...),
		"p": append([]int{}, shoupai.bingpai["p"]...),
		"s": append([]int{}, shoupai.bingpai["s"]...),
		"z": append([]int{}, shoupai.bingpai["z"]...),
	}
	s.fulou = append([]string{}, shoupai.fulou...)
	s.zimo = shoupai.zimo
	s.lizhi = shoupai.lizhi

	return s
}

func (s *Shoupai) Decrease(suit string, n int) error {
	bingpai := s.bingpai[suit]
	if bingpai[n] == 0 || (n == 5 && bingpai[0] == bingpai[5]) {
		if s.bingpai["_"][0] == 0 {
			return errors.New("Error: " + suit + strconv.Itoa(n))
		}
		s.bingpai["_"][0]--
	} else {
		bingpai[n]--
		if n == 0 {
			bingpai[5]--
		}
	}
	return nil
}

func (s *Shoupai) Zimo(p string, check bool) error {
	if check && s.zimo != "" {
		return errors.New("Error: " + p)
	}
	if p == "_" {
		s.bingpai["_"][0]++
		s.zimo = p
	} else {
		if !validPai(p) {
			return errors.New("Invalid pai: " + p)
		}
		suit := string(p[0])
		n, _ := strconv.Atoi(string(p[1]))
		bingpai := s.bingpai[suit]
		if bingpai[n] == 4 {
			return errors.New("Too many pai: " + p)
		}
		bingpai[n]++
		if n == 0 {
			if bingpai[5] == 4 {
				return errors.New("Too many pai: " + p)
			}
			bingpai[5]++
		}
		s.zimo = suit + strconv.Itoa(n)
	}
	return nil
}

func main() {
	shoupai := NewShoupai().FromString("m123p456s789z4567")
	shoupai.zimo = "z7"
	println(shoupai.Clone().ToString())
}
