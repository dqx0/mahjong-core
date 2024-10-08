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

func (s *Shoupai) Zimo(p string) error {
	return s.ZimoWithCheck(p, true)
}

func (s *Shoupai) ZimoWithCheck(p string, check bool) error {
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

func (s *Shoupai) Dapai(p string) error {
	return s.DapaiWithCheck(p, true)
}

func (s *Shoupai) DapaiWithCheck(p string, check bool) error {
	if check && s.zimo == "" {
		return errors.New("Error: " + p)
	}
	if !validPai(p) {
		return errors.New("Invalid pai: " + p)
	}
	suit := string(p[0])
	n, _ := strconv.Atoi(string(p[1]))
	if err := s.Decrease(suit, n); err != nil {
		return err
	}
	s.zimo = ""
	if strings.HasSuffix(p, "*") {
		s.lizhi = true
	}
	return nil
}

func (s *Shoupai) Fulou(m string, check bool) error {
	if check && s.zimo != "" {
		return errors.New("Error: " + m)
	}
	if m != validMianzi(m) {
		return errors.New("Invalid mianzi: " + m)
	}
	if matched, _ := regexp.MatchString(`\d{4}$`, m); matched {
		return errors.New("Error: " + m)
	}
	if matched, _ := regexp.MatchString(`\d{3}[\+\=\-]\d$`, m); matched {
		return errors.New("Error: " + m)
	}
	suit := string(m[0])
	numbers := regexp.MustCompile(`\d(?![\+\=\-])`).FindAllString(m, -1)
	for _, nStr := range numbers {
		n, _ := strconv.Atoi(nStr)
		if err := s.Decrease(suit, n); err != nil {
			return err
		}
	}
	s.fulou = append(s.fulou, m)
	if matched, _ := regexp.MatchString(`\d{4}`, m); !matched {
		s.zimo = m
	}
	return nil
}

func (s *Shoupai) Gang(m string, check bool) error {
	if check && s.zimo == "" {
		return errors.New("Error: " + m)
	}
	if check && len(s.zimo) > 2 {
		return errors.New("Error: " + m)
	}
	if m != validMianzi(m) {
		return errors.New("Invalid mianzi: " + m)
	}
	suit := string(m[0])
	if matched, _ := regexp.MatchString(`\d{4}$`, m); matched {
		numbers := regexp.MustCompile(`\d`).FindAllString(m, -1)
		for _, nStr := range numbers {
			n, _ := strconv.Atoi(nStr)
			if err := s.Decrease(suit, n); err != nil {
				return err
			}
		}
		s.fulou = append(s.fulou, m)
	} else if matched, _ := regexp.MatchString(`\d{3}[\+\=\-]\d$`, m); matched {
		m1 := m[:5]
		i := -1
		for index, m2 := range s.fulou {
			if m1 == m2 {
				i = index
				break
			}
		}
		if i < 0 {
			return errors.New("Error: " + m)
		}
		s.fulou[i] = m
		n, _ := strconv.Atoi(string(m[len(m)-1]))
		if err := s.Decrease(suit, n); err != nil {
			return err
		}
	} else {
		return errors.New("Error: " + m)
	}
	s.zimo = ""
	return nil
}

func (s *Shoupai) Menqian() bool {
	re := regexp.MustCompile(`[\+\=\-]`)
	for _, m := range s.fulou {
		if re.MatchString(m) {
			return false
		}
	}
	return true
}

func (s *Shoupai) Lizhi() bool {
	return s.lizhi
}

func (s *Shoupai) GetDapai(check bool) []string {
	if s.zimo == "" {
		return nil
	}

	deny := make(map[string]bool)
	if check && len(s.zimo) > 2 {
		m := s.zimo
		suit := string(m[0])
		n := 5
		if match := regexp.MustCompile(`\d`).FindString(m[1:2]); match != "" {
			n, _ = strconv.Atoi(match)
		}
		deny[suit+strconv.Itoa(n)] = true
		if !regexp.MustCompile(`^[mpsz](\d)\d\d`).MatchString(strings.Replace(m, "0", "5", 1)) {
			if n < 7 && regexp.MustCompile(`^[mps]\d-\d\d$`).MatchString(m) {
				deny[suit+strconv.Itoa(n+3)] = true
			}
			if n > 3 && regexp.MustCompile(`^[mps]\d\d\d-$`).MatchString(m) {
				deny[suit+strconv.Itoa(n-3)] = true
			}
		}
	}

	var dapai []string
	if !s.lizhi {
		for _, suit := range []string{"m", "p", "s", "z"} {
			bingpai := s.bingpai[suit]
			for n := 1; n < len(bingpai); n++ {
				if bingpai[n] == 0 {
					continue
				}
				if deny[suit+strconv.Itoa(n)] {
					continue
				}
				if suit+strconv.Itoa(n) == s.zimo && bingpai[n] == 1 {
					continue
				}
				if suit == "z" || n != 5 {
					dapai = append(dapai, suit+strconv.Itoa(n))
				} else {
					if bingpai[0] > 0 && (suit+"0" != s.zimo || bingpai[0] > 1) {
						dapai = append(dapai, suit+"0")
					}
					if bingpai[0] < bingpai[5] {
						dapai = append(dapai, suit+strconv.Itoa(n))
					}
				}
			}
		}
	}
	if len(s.zimo) == 2 {
		dapai = append(dapai, s.zimo+"_")
	}
	return dapai
}

func (s *Shoupai) GetChiMianzi(p string, check bool) ([]string, error) {
	if s.zimo != "" {
		return nil, nil
	}
	if !validPai(p) {
		return nil, errors.New("Invalid pai: " + p)
	}

	var mianzi []string
	suit := string(p[0])
	n := 5
	if p[1] != '0' {
		n, _ = strconv.Atoi(string(p[1]))
	}
	d := regexp.MustCompile(`[\+\=\-]$`).FindString(p)
	if d == "" {
		return nil, errors.New("Invalid pai: " + p)
	}
	if suit == "z" || d != "-" {
		return mianzi, nil
	}
	if s.lizhi {
		return mianzi, nil
	}

	bingpai := s.bingpai[suit]
	if n >= 3 && bingpai[n-2] > 0 && bingpai[n-1] > 0 {
		if !check || (n > 3 && bingpai[n-3]+bingpai[n] < 14-(len(s.fulou)+1)*3) {
			if n-2 == 5 && bingpai[0] > 0 {
				mianzi = append(mianzi, suit+"067-")
			}
			if n-1 == 5 && bingpai[0] > 0 {
				mianzi = append(mianzi, suit+"406-")
			}
			if (n-2 != 5 && n-1 != 5) || bingpai[0] < bingpai[5] {
				mianzi = append(mianzi, suit+strconv.Itoa(n-2)+strconv.Itoa(n-1)+string(p[1])+d)
			}
		}
	}
	if n >= 2 && n <= 8 && bingpai[n-1] > 0 && bingpai[n+1] > 0 {
		if !check || bingpai[n] < 14-(len(s.fulou)+1)*3 {
			if n-1 == 5 && bingpai[0] > 0 {
				mianzi = append(mianzi, suit+"06-7")
			}
			if n+1 == 5 && bingpai[0] > 0 {
				mianzi = append(mianzi, suit+"34-0")
			}
			if (n-1 != 5 && n+1 != 5) || bingpai[0] < bingpai[5] {
				mianzi = append(mianzi, suit+strconv.Itoa(n-1)+string(p[1])+d+strconv.Itoa(n+1))
			}
		}
	}
	if n <= 7 && bingpai[n+1] > 0 && bingpai[n+2] > 0 {
		var bingpaiNPlus3 int
		if n < 7 {
			bingpaiNPlus3 = bingpai[n+3]
		} else {
			bingpaiNPlus3 = 0
		}
		if !check || bingpai[n]+bingpaiNPlus3 < 14-(len(s.fulou)+1)*3 {
			if n+1 == 5 && bingpai[0] > 0 {
				mianzi = append(mianzi, suit+"4-06")
			}
			if n+2 == 5 && bingpai[0] > 0 {
				mianzi = append(mianzi, suit+"3-40")
			}
			if (n+1 != 5 && n+2 != 5) || bingpai[0] < bingpai[5] {
				mianzi = append(mianzi, suit+string(p[1])+d+strconv.Itoa(n+1)+strconv.Itoa(n+2))
			}
		}
	}
	return mianzi, nil
}

func (s *Shoupai) GetPengMianzi(p string) ([]string, error) {
	if s.zimo != "" {
		return nil, nil
	}
	if !validPai(p) {
		return nil, errors.New("Invalid pai: " + p)
	}

	var mianzi []string
	suit := string(p[0])
	n := 5
	if p[1] != '0' {
		n, _ = strconv.Atoi(string(p[1]))
	}
	d := regexp.MustCompile(`[\+\=\-]$`).FindString(p)
	if d == "" {
		return nil, errors.New("Invalid pai: " + p)
	}
	if s.lizhi {
		return mianzi, nil
	}

	bingpai := s.bingpai[suit]
	if bingpai[n] >= 2 {
		if n == 5 && bingpai[0] >= 2 {
			mianzi = append(mianzi, suit+"00"+string(p[1])+d)
		}
		if n == 5 && bingpai[0] >= 1 && bingpai[5]-bingpai[0] >= 1 {
			mianzi = append(mianzi, suit+"50"+string(p[1])+d)
		}
		if n != 5 || bingpai[5]-bingpai[0] >= 2 {
			mianzi = append(mianzi, suit+strconv.Itoa(n)+strconv.Itoa(n)+string(p[1])+d)
		}
	}
	return mianzi, nil
}

func (s *Shoupai) GetGangMianzi(p string) ([]string, error) {
	var mianzi []string
	if p != "" {
		if s.zimo != "" {
			return nil, nil
		}
		if !validPai(p) {
			return nil, errors.New("Invalid pai: " + p)
		}

		suit := string(p[0])
		n := 5
		if p[1] != '0' {
			n, _ = strconv.Atoi(string(p[1]))
		}
		d := regexp.MustCompile(`[\+\=\-]$`).FindString(p)
		if d == "" {
			return nil, errors.New("Invalid pai: " + p)
		}
		if s.lizhi {
			return mianzi, nil
		}

		bingpai := s.bingpai[suit]
		if bingpai[n] == 3 {
			if n == 5 {
				mianzi = []string{suit + strings.Repeat("5", 3-bingpai[0]) + strings.Repeat("0", bingpai[0]) + string(p[1]) + d}
			} else {
				mianzi = []string{suit + strconv.Itoa(n) + strconv.Itoa(n) + strconv.Itoa(n) + strconv.Itoa(n) + d}
			}
		}
	} else {
		if s.zimo == "" {
			return nil, nil
		}
		if len(s.zimo) > 2 {
			return nil, nil
		}
		p = strings.Replace(s.zimo, "0", "5", 1)

		for _, suit := range []string{"m", "p", "s", "z"} {
			bingpai := s.bingpai[suit]
			for n := 1; n < len(bingpai); n++ {
				if bingpai[n] == 0 {
					continue
				}
				if bingpai[n] == 4 {
					if s.lizhi && suit+strconv.Itoa(n) != p {
						continue
					}
					if n == 5 {
						mianzi = append(mianzi, suit+strings.Repeat("5", 4-bingpai[0])+strings.Repeat("0", bingpai[0]))
					} else {
						mianzi = append(mianzi, suit+strconv.Itoa(n)+strconv.Itoa(n)+strconv.Itoa(n)+strconv.Itoa(n))
					}
				} else {
					if s.lizhi {
						continue
					}
					for _, m := range s.fulou {
						if strings.Replace(m, "0", "5", -1)[:4] == suit+strconv.Itoa(n)+strconv.Itoa(n)+strconv.Itoa(n) {
							if n == 5 && bingpai[0] > 0 {
								mianzi = append(mianzi, m+"0")
							} else {
								mianzi = append(mianzi, m+strconv.Itoa(n))
							}
						}
					}
				}
			}
		}
	}
	return mianzi, nil
}

func main() {
	shoupai := FromString("m123p456s789z34567")
	err := shoupai.Dapai("m1")
	if err != nil {
		fmt.Println(err)
	}

	println(shoupai.Clone().ToString())
}
