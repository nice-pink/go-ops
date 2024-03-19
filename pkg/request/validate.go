package request

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	RedirectContains string = "redirect-contains:"
	BodyContains     string = "body-contains:"
)

func Validate(resp *http.Response, validate string) bool {
	fmt.Print("Validate:", validate, " ")

	isValid := true
	if strings.HasPrefix(validate, RedirectContains) {
		// redirect contains string
		short, _ := strings.CutPrefix(validate, RedirectContains)
		loc, err := resp.Location()
		if loc == nil || err != nil {
			isValid = false
			fmt.Println("❌ Is redirected. Status:", resp.StatusCode, "and does not contain", short)
		} else {
			isValid = strings.Contains(resp.Header.Get("Location"), short)
			if !isValid {
				fmt.Println("❌ Is redirected. Status:", resp.StatusCode, "and does not contain", short)
			}
		}

	} else if strings.HasPrefix(validate, BodyContains) {
		// body contains
		short, _ := strings.CutPrefix(validate, RedirectContains)
		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			isValid = false
			if !isValid {
				fmt.Println("❌ Can't read body.")
			}
		}
		body := string(b)
		isValid = strings.Contains(body, short)
		if !isValid {
			fmt.Println("❌ Body does not contain:", short)
		}
	}

	if isValid {
		fmt.Println(" ✅")
	}

	return isValid
}
