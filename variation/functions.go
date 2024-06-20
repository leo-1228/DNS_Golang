package variation

import (
	"errors"
	"fmt"
	"strings"
)

var homoglyphs = map[string][]string{
	"0": {"o"},
	"1": {"l", "i"},
	"2": {"ƻ"},
	"3": {"8"},
	"5": {"ƽ"},
	"6": {"9"},
	"8": {"3"},
	"9": {"6"},
	"a": {"à", "á", "à", "â", "ã", "ä", "å", "ɑ", "ạ", "ǎ", "ă", "ȧ", "ą", "ə"},
	"b": {"d", "lb", "ʙ", "ɓ", "ḃ", "ḅ", "ḇ", "ƅ"},
	"c": {"e", "ƈ", "ċ", "ć", "ç", "č", "ĉ", "ᴄ"},
	"d": {"b", "cl", "dl", "ɗ", "đ", "ď", "ɖ", "ḑ", "ḋ", "ḍ", "ḏ", "ḓ"},
	"e": {"c", "é", "è", "ê", "ë", "ē", "ĕ", "ě", "ė", "ẹ", "ę", "ȩ", "ɇ", "ḛ"},
	"f": {"ƒ", "ḟ"},
	"g": {"q", "ɢ", "ɡ", "ġ", "ğ", "ǵ", "ģ", "ĝ", "ǧ", "ǥ"},
	"h": {"lh", "ĥ", "ȟ", "ħ", "ɦ", "ḧ", "ḩ", "ⱨ", "ḣ", "ḥ", "ḫ", "ẖ"},
	"i": {"1", "l", "í", "ì", "ï", "ı", "ɩ", "ǐ", "ĭ", "ỉ", "ị", "ɨ", "ȋ", "ī", "ɪ"},
	"j": {"ʝ", "ǰ", "ɉ", "ĵ"},
	"k": {"lk", "ik", "lc", "ḳ", "ḵ", "ⱪ", "ķ", "ᴋ"},
	"l": {"1", "i", "ɫ", "ł"},
	"m": {"n", "nn", "rn", "rr", "ṁ", "ṃ", "ᴍ", "ɱ", "ḿ"},
	"n": {"m", "r", "ń", "ṅ", "ṇ", "ṉ", "ñ", "ņ", "ǹ", "ň", "ꞑ"},
	"o": {"0", "ȯ", "ọ", "ỏ", "ơ", "ó", "ö", "ᴏ"},
	"p": {"ƿ", "ƥ", "ṕ", "ṗ"},
	"q": {"g", "ʠ"},
	"r": {"ʀ", "ɼ", "ɽ", "ŕ", "ŗ", "ř", "ɍ", "ɾ", "ȓ", "ȑ", "ṙ", "ṛ", "ṟ"},
	"s": {"ʂ", "ś", "ṣ", "ṡ", "ș", "ŝ", "š", "ꜱ"},
	"t": {"ţ", "ŧ", "ṫ", "ṭ", "ț", "ƫ"},
	"u": {"ᴜ", "ǔ", "ŭ", "ü", "ʉ", "ù", "ú", "û", "ũ", "ū", "ų", "ư", "ů", "ű", "ȕ", "ȗ", "ụ"},
	"v": {"ṿ", "ⱱ", "ᶌ", "ṽ", "ⱴ", "ᴠ"},
	"w": {"vv", "ŵ", "ẁ", "ẃ", "ẅ", "ⱳ", "ẇ", "ẉ", "ẘ", "ᴡ"},
	"x": {"ẋ", "ẍ"},
	"y": {"ʏ", "ý", "ÿ", "ŷ", "ƴ", "ȳ", "ɏ", "ỿ", "ẏ", "ỵ"},
	"z": {"ʐ", "ż", "ź", "ᴢ", "ƶ", "ẓ", "ẕ", "ⱬ"},
}

var qwerty = map[string]string{
	"1": "2q", "2": "3wq1", "3": "4ew2", "4": "5re3", "5": "6tr4", "6": "7yt5", "7": "8uy6", "8": "9iu7", "9": "0oi8", "0": "po9",
	"q": "12wa", "w": "3esaq2", "e": "4rdsw3", "r": "5tfde4", "t": "6ygfr5", "y": "7uhgt6", "u": "8ijhy7", "i": "9okju8", "o": "0plki9", "p": "lo0",
	"a": "qwsz", "s": "edxzaw", "d": "rfcxse", "f": "tgvcdr", "g": "yhbvft", "h": "ujnbgy", "j": "ikmnhu", "k": "olmji", "l": "kop",
	"z": "asx", "x": "zsdc", "c": "xdfv", "v": "cfgb", "b": "vghn", "n": "bhjm", "m": "njk",
}
var qwertz = map[string]string{
	"1": "2q", "2": "3wq1", "3": "4ew2", "4": "5re3", "5": "6tr4", "6": "7zt5", "7": "8uz6", "8": "9iu7", "9": "0oi8", "0": "po9",
	"q": "12wa", "w": "3esaq2", "e": "4rdsw3", "r": "5tfde4", "t": "6zgfr5", "z": "7uhgt6", "u": "8ijhz7", "i": "9okju8", "o": "0plki9", "p": "lo0",
	"a": "qwsy", "s": "edxyaw", "d": "rfcxse", "f": "tgvcdr", "g": "zhbvft", "h": "ujnbgz", "j": "ikmnhu", "k": "olmji", "l": "kop",
	"y": "asx", "x": "ysdc", "c": "xdfv", "v": "cfgb", "b": "vghn", "n": "bhjm", "m": "njk",
}
var azerty = map[string]string{
	"1": "2a", "2": "3za1", "3": "4ez2", "4": "5re3", "5": "6tr4", "6": "7yt5", "7": "8uy6", "8": "9iu7", "9": "0oi8", "0": "po9",
	"a": "2zq1", "z": "3esqa2", "e": "4rdsz3", "r": "5tfde4", "t": "6ygfr5", "y": "7uhgt6", "u": "8ijhy7", "i": "9okju8", "o": "0plki9", "p": "lo0m",
	"q": "zswa", "s": "edxwqz", "d": "rfcxse", "f": "tgvcdr", "g": "yhbvft", "h": "ujnbgy", "j": "iknhu", "k": "olji", "l": "kopm", "m": "lp",
	"w": "sxq", "x": "wsdc", "c": "xdfv", "v": "cfgb", "b": "vghn", "n": "bhj",
}
var keyboards = []map[string]string{qwerty, qwertz, azerty}

var vowels = "aeiou"

type DInfo struct {
	Prefix     string
	MainDomain string
	Suffix     string
}

func (d DInfo) Original() string {
	p := ""
	if len(d.Prefix) > 0 {
		p = d.Prefix + "."
	}
	return fmt.Sprintf("%s%s%s", p, d.MainDomain, d.Suffix)
}
func (d DInfo) NewDomainFormat(newMainDomain string) string {
	p := ""
	if len(d.Prefix) > 0 {
		p = d.Prefix + "."
	}
	return fmt.Sprintf("%s%s%s", p, newMainDomain, d.Suffix)
}
func (d DInfo) NewSpaceFormat(space string) string {
	p := ""
	if len(d.Prefix) > 0 {
		p = d.Prefix + "."
	}
	return fmt.Sprintf("%s%s%s", p, d.MainDomain, space)
}
func (d DInfo) GetInAllSpaces(spaces []string) []string {

	result := make([]string, 0)

	for _, sp := range spaces {

		newDomain := d.NewSpaceFormat(sp)
		result = append(result, newDomain)
	}

	return result
}
func DomainInfo(domain string, spaces []string) (DInfo, error) {
	domain = strings.TrimSpace(domain)
	//

	space := ""
	for _, v := range spaces {
		if strings.HasSuffix(strings.ToLower(domain), strings.ToLower(v)) {
			space = v
			break
		}
	}
	if space == "" {
		return DInfo{}, errors.New("space not supported")
	}

	parts := strings.Split(domain, ".")

	if len(parts[0]) == 0 {
		return DInfo{}, errors.New("invalid domain")
	}

	if len(parts) <= 3 {
		return DInfo{
			Prefix:     "",
			MainDomain: parts[0],
			Suffix:     space,
		}, nil
	}
	return DInfo{
		Prefix:     strings.Join(parts[:len(parts)-3], "."),
		MainDomain: parts[len(parts)-3],
		Suffix:     space,
	}, nil
}

func (v *Variation) Homoglyph(domain string) ([]VariationRecord, error) {
	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	mix := func(domain string) map[string]bool {
		result := map[string]bool{}
		final := map[string]bool{}

		for w := 1; w < len(domain); w++ {
			for i := 0; i < len(domain)-w+1; i++ {
				pre := domain[:i]
				win := domain[i : i+w]
				suf := domain[i+w:]
				for _, c := range win {
					for _, g := range homoglyphs[string(c)] {
						t := pre + strings.Replace(win, string(c), g, -1) + suf
						if _, exist := result[t]; !exist {
							result[t] = true
							final[dInfo.NewDomainFormat(t)] = true
						}
					}
				}
			}
		}
		return final
	}
	result1 := mix(domain)

	if !v.homoglyphNormal {
		for r := range result1 {
			result2 := mix(r)
			for r2 := range result2 {
				result1[r2] = true
			}
		}
	}

	finalResult := make([]VariationRecord, 0, len(result1))
	for key := range result1 {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "Homoglyph", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
func (v *Variation) Addition(domain string) ([]VariationRecord, error) {
	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	result := map[string]bool{}
	final := map[string]bool{}

	for i := 48; i <= 122; i++ {
		if (i >= 48 && i <= 57) || (i >= 97 && i <= 122) {
			t := fmt.Sprintf("%s%c", domain, i)
			if _, exist := result[t]; !exist {
				result[t] = true
				final[dInfo.NewDomainFormat(t)] = true
			}
		}
	}

	finalResult := make([]VariationRecord, 0, len(final))
	for key := range final {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "Addition", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
func (v *Variation) Bitsquatting(domain string) ([]VariationRecord, error) {
	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	result := map[string]bool{}
	final := map[string]bool{}

	masks := []byte{1, 2, 4, 8, 16, 32, 64, 128}
	chars := "abcdefghijklmnopqrstuvwxyz0123456789-"
	for i := 0; i < len(domain); i++ {
		for _, mask := range masks {
			b := domain[i] ^ mask
			if strings.ContainsRune(chars, rune(b)) {
				mixed := []byte(domain)
				mixed[i] = b
				t := string(mixed)

				if _, exist := result[t]; !exist {
					result[t] = true
					final[dInfo.NewDomainFormat(t)] = true
				}

			}
		}
	}

	finalResult := make([]VariationRecord, 0, len(final))
	for key := range final {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "Bitsquatting", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
func (v *Variation) Hyphenation(domain string) ([]VariationRecord, error) {

	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	result := map[string]bool{}
	final := map[string]bool{}

	for i := 1; i < len(domain); i++ {
		t := domain[:i] + "-" + domain[i:]
		if _, exist := result[t]; !exist {
			result[t] = true
			final[dInfo.NewDomainFormat(t)] = true
		}
	}

	finalResult := make([]VariationRecord, 0, len(final))
	for key := range final {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "Hyphenation", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
func (v *Variation) Insertion(domain string) ([]VariationRecord, error) {

	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	result := map[string]bool{}
	final := map[string]bool{}

	for i := 1; i < len(domain)-1; i++ {
		prefix, orig_c, suffix := domain[:i], domain[i], domain[i+1:]
		for _, keyboard := range keyboards {
			if chars, ok := keyboard[string(orig_c)]; ok {
				for _, c := range chars {
					t := prefix + string(c) + string(orig_c) + suffix
					if _, exist := result[t]; !exist {
						result[t] = true
						final[dInfo.NewDomainFormat(t)] = true
					}
					t = prefix + string(orig_c) + string(c) + suffix
					if _, exist := result[t]; !exist {
						result[t] = true
						final[dInfo.NewDomainFormat(t)] = true
					}
				}
			}
		}
	}

	finalResult := make([]VariationRecord, 0, len(final))
	for key := range final {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "Insertion", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
func (v *Variation) Omission(domain string) ([]VariationRecord, error) {
	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	result := map[string]bool{}
	final := map[string]bool{}

	for i := 0; i < len(domain); i++ {
		t := domain[:i] + domain[i+1:]
		if _, exist := result[t]; !exist {
			result[t] = true
			final[dInfo.NewDomainFormat(t)] = true
		}
	}

	finalResult := make([]VariationRecord, 0, len(final))
	for key := range final {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "Omission", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
func (v *Variation) Repetition(domain string) ([]VariationRecord, error) {
	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	result := map[string]bool{}
	final := map[string]bool{}

	for i, c := range domain {
		t := domain[:i] + string(c) + domain[i:]
		if _, exist := result[t]; !exist {
			result[t] = true
			final[dInfo.NewDomainFormat(t)] = true
		}
	}

	finalResult := make([]VariationRecord, 0, len(final))
	for key := range final {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "Repetition", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
func (v *Variation) Replacement(domain string) ([]VariationRecord, error) {
	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	result := map[string]bool{}
	final := map[string]bool{}

	for i, c := range domain {
		pre := domain[:i]
		suf := domain[i+1:]
		for _, layout := range keyboards {
			if repl, ok := layout[string(c)]; ok {
				for _, r := range repl {
					t := pre + string(r) + suf

					if _, exist := result[t]; !exist {
						result[t] = true
						final[dInfo.NewDomainFormat(t)] = true
					}
				}
			}
		}
	}

	finalResult := make([]VariationRecord, 0, len(final))
	for key := range final {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "Replacement", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
func (v *Variation) Transposition(domain string) ([]VariationRecord, error) {

	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	result := map[string]bool{}
	final := map[string]bool{}

	for i := 0; i < len(domain)-1; i++ {
		t := domain[:i] + string(domain[i+1]) + string(domain[i]) + domain[i+2:]

		if _, exist := result[t]; !exist {
			result[t] = true
			final[dInfo.NewDomainFormat(t)] = true
		}

	}

	finalResult := make([]VariationRecord, 0, len(final))
	for key := range final {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "Transposition", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
func (v *Variation) VowelSwap(domain string) ([]VariationRecord, error) {

	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	result := map[string]bool{}
	final := map[string]bool{}

	for i := 0; i < len(domain); i++ {
		if strings.Contains(vowels, string(domain[i])) {
			for _, vowel := range vowels {
				t := domain[:i] + string(vowel) + domain[i+1:]
				if _, exist := result[t]; !exist {
					result[t] = true
					final[dInfo.NewDomainFormat(t)] = true
				}
			}
		}
	}

	finalResult := make([]VariationRecord, 0, len(final))
	for key := range final {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "VowelSwap", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
func (v *Variation) Dictionary(domain string) ([]VariationRecord, error) {

	dInfo, err := DomainInfo(domain, v.validSpaces)
	if err != nil {
		return nil, err
	}
	domain = dInfo.MainDomain

	result := map[string]bool{}
	final := map[string]bool{}

	for _, word := range v.dictionaryData {
		if !(strings.HasPrefix(domain, word) && strings.HasSuffix(domain, word)) {
			t := domain + "-" + word
			if _, exist := result[t]; !exist {
				result[t] = true
				final[dInfo.NewDomainFormat(t)] = true
			}
			t = domain + word
			if _, exist := result[t]; !exist {
				result[t] = true
				final[dInfo.NewDomainFormat(t)] = true
			}
			t = word + "-" + domain
			if _, exist := result[t]; !exist {
				result[t] = true
				final[dInfo.NewDomainFormat(t)] = true
			}
			t = word + domain
			if _, exist := result[t]; !exist {
				result[t] = true
				final[dInfo.NewDomainFormat(t)] = true
			}
		}
	}
	if strings.Contains(domain, "-") {
		parts := strings.Split(domain, "-")
		for _, word := range v.dictionaryData {
			t := strings.Join(parts[:len(parts)-1], "-") + "-" + word
			if _, exist := result[t]; !exist {
				result[t] = true
				final[dInfo.NewDomainFormat(t)] = true
			}
			t = word + "-" + strings.Join(parts[1:], "-")
			if _, exist := result[t]; !exist {
				result[t] = true
				final[dInfo.NewDomainFormat(t)] = true
			}
		}
	}

	finalResult := make([]VariationRecord, 0, len(final))
	for key := range final {
		if v.checkMainDomainDuplication {
			if _, ext := v.mainDomains[key]; ext {
				continue
			}
		}
		finalResult = append(finalResult, VariationRecord{Variant: key, How: "Dictionary", MainDomain: dInfo.Original()})
	}
	return finalResult, nil
}
