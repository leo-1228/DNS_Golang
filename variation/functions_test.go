package variation

import (
	"testing"
	"time"
)

func TestGo(t *testing.T) {

	ch := make(chan int, 1000)

	go func() {
		for c := range ch {
			t.Log(c)
			time.Sleep(time.Millisecond * 500)
		}
	}()

	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
		}
		close(ch)
		t.Log("channel closed")
	}()

	time.Sleep(time.Second * 5)
}

func TestHomoglyph(t *testing.T) {

	// result, err := Homoglyph("aaaaircon.net.au", true)
	// t.Log(err)
	// t.Log(len(result))

	// result2, err := Addition("aaaaircon.net.au")
	// t.Log(err)
	// t.Log(result2)

	// result3, err := Bitsquatting("aaaaircon.net.au")
	// t.Log(err)
	// t.Log(result3)

	// result4, err := Hyphenation("aaaaircon.net.au")
	// t.Log(err)
	// t.Log(result4)

	// result4, err := Insertion("aaaaircon.net.au")
	// t.Log(err)
	// t.Log(result4)

	// result4, err := Omission("aaaaircon.net.au")
	// t.Log(err)
	// t.Log(result4)

	// 	result4, err := Repetition("aaaaircon.net.au")
	// 	t.Log(err)
	// 	t.Log(result4)

	// result4, err := Replacement("aaaaircon.net.au")
	// t.Log(err)
	// t.Log(result4)

	// result4, err := Transposition("aaaaircon.net.au")
	// t.Log(err)
	// t.Log(result4)

	// result4, err := VowelSwap("aaaaircon.net.au")
	// t.Log(err)
	// t.Log(result4)

	dic := []string{
		"mail",
		"my",
		"online",
		"pay",
		"payment",
		"payments",
		"portal",
		"recovery",
	}
	va, err := New(Config{
		ValidSpaces:         []string{".com.au", ".net.au", ".org.au", ".asn.au", ".id.au", ".au"},
		HomoglyphMethod:     true,
		HomoglyphNormal:     false,
		AdditionMethod:      true,
		BitsquattingMethod:  true,
		HyphenationMethod:   true,
		InsertionMethod:     true,
		OmissionMethod:      true,
		RepetitionMethod:    true,
		ReplacementMethod:   true,
		TranspositionMethod: true,
		VowelSwapMethod:     true,
		DictionaryMethod:    true,
		DictionaryData:      dic,
	})
	if err != nil {
		t.Fatal(err)
	}

	result4, err := va.Run([]string{
		"001.net.au",
		"007.net.au",
		"100.net.au",
		"121creative.net.au",
		"127.net.au",
		"12free.net.au",
		"1300854745.net.au",
		"1300catcher.net.au",
		"1300phonenumbers.net.au",
		"13quest.net.au",
		"13telco.net.au",
		"16888.net.au",
		"1882.net.au",
		"1leopard.net.au",
		"1stoprenovations.net.au",
		"1wilshire2230.net.au",
		"200.net.au",
		"2-0.net.au",
		"234.net.au",
		"24-7.net.au",
		"2birds.net.au",
		"306090.net.au",
		"33waystoselfsabotage.net.au",
	})
	t.Log(err)
	total := 0
	if len(result4) > 0 {
		for i := 0; i < len(result4); i++ {
			total += len(result4[i].Variations)
			t.Log(len(result4[i].Variations))
		}
	}
	t.Log("total:", total)
}

func TestDomainInfo(t *testing.T) {
	info, err := DomainInfo("test.google.com.au", []string{".com.au", ".net.au", ".org.au", ".asn.au", ".id.au", ".au", ".com"})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(info.GetInAllSpaces([]string{".com.au", ".net.au", ".org.au", ".asn.au", ".id.au", ".au", ".com"}))
}
