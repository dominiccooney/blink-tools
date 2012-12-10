package history

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"parse"
	"strings"
	"time"
)

type Bug struct {
	Id int
}

type AttachmentFlag struct {
	Name string
}

var (
	Review = AttachmentFlag{"review?"}
	ReviewGranted = AttachmentFlag{"review+"}
	ReviewDenied = AttachmentFlag{"review-"}
)

type HistoryObserver interface {
	// TODO: Flesh this out
	AttachmentFlag(when time.Time, id int, removed []AttachmentFlag, added []AttachmentFlag)
}

func (b *Bug) HistoryURL() string {
	return fmt.Sprintf("https://bugs.webkit.org/show_activity.cgi?id=%d", b.Id)
}

func get(url string) (result string, error error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error retrieving %s\n%s", url, err.Error()))
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("error retrieving %s\n%s", url, err.Error()))
	}
	// TODO: use the correct encoding
	return string(body), nil
}

func ReplayHistory(b *Bug, observer HistoryObserver) error {
	zone, err := time.LoadLocation("US/Pacific")
	if err != nil {
		panic("could not load US/Pacific timezone")
	}

	body, err := get(b.HistoryURL())
	if err != nil {
		return err
	}

	scan := parse.NewUtf8Scanner(body)
	scan.SkipPast("<th>Added</th>")

	rowspan := 1
	var when time.Time

	for {
		scan.SkipPast("<tr>")

		if rowspan == 1 {
			err = scan.SkipPast("<td rowspan=\"")
			if err != nil {
				break
			}

			rowspan, err = scan.Int()
			if err != nil {
				panic(fmt.Sprintf("rowspan: %s", err.Error()))
			}

			scan.SkipPast("</td>")  // end of 'who'
			scan.SkipPast(">")      // start of 'when'

			year, _ := scan.Int()
			scan.Lit("-")
			month, _ := scan.Int()
			scan.Lit("-")
			day, _ := scan.Int()
			scan.Lit(" ")
			hour, _ := scan.Int()
			scan.Lit(":")
			minute, _ := scan.Int()
			scan.Lit(":")
			second, _ := scan.Int()
			scan.Lit(" PST")

			when = time.Date(year, time.Month(month), day, hour, minute, second, 0, zone)

		} else {
			rowspan--
		}

		scan.SkipPast("<td>")  // 'what'
		scan.SkipSpace()

		if scan.Peek() == '<' {
			// Attachment
			scan.SkipPast("Attachment #")
			attachment, _ := scan.Int()
			scan.SkipPast("</a>")
			scan.SkipSpace()
			if scan.Peek() != 'F' {
				// Attachment ... is obsolete
				continue
			}
			scan.SkipPast("<td>")  // removed flags
			removed := ParseAttachmentFlags(scan)
			scan.SkipPast("<td>")  // added flags
			added := ParseAttachmentFlags(scan)

			observer.AttachmentFlag(when, attachment, removed, added)
		}
	}

	return nil
}

func ParseAttachmentFlags(scan parse.Scanner) []AttachmentFlag {
	field, err := scan.ReadWhile(func (r rune) bool {
		return r != '<'
	})
	if err != nil {
		panic("didn't find < reading flag list")
	}
	flags := make([]AttachmentFlag, 0, 1)
	for _, flag := range([...]AttachmentFlag{Review, ReviewGranted, ReviewDenied}) {
		if strings.Index(field, flag.Name) != -1 {
			flags = append(flags, flag)
		}
	}
	return flags
}
