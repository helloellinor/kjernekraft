package handsamarar

import (
	"net/http"
	"strconv"
	"strings"
)

// What the two price lists share.
//
// Memberships and klippekort keep separate handlers on purpose: one has
// binding and a student rate, the other clips and validity days, and
// only memberships have rows that must not be writable (Skjult, the
// Black card). Forcing both through one handler would be a shared shell
// over two different things. The form, however, is the same form — it
// posts every row, marks deletions with "slett-<id>", adds rows with
// "namn-ny…", and lets the server diff. These three are that part.

// reintNamn strips control characters and surrounding space.
//
// A name that is not a name must not be storable whatever sends the
// form: the dock once wrote a NUL into an empty field, and three
// memberships called NUL reached the database that way.
func reintNamn(s string) string {
	reint := strings.Map(func(r rune) rune {
		if r < 32 || r == 127 || r == 0xFFFD {
			return -1
		}
		return r
	}, s)
	return strings.TrimSpace(reint)
}

// prisTal reads a number from the price form, falling back to standard
// when the field is empty, unparseable or negative. Putting the default
// in the call reads as "unchanged if nobody typed anything".
func prisTal(r *http.Request, nykel string, standard int) int {
	if s := r.FormValue(nykel); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 0 {
			return n
		}
	}
	return standard
}

// nyeRadene gives the rows the form wants to add: the suffix their
// other fields carry, and the name typed. A row without a name is one
// somebody changed their mind about, and is left out.
func nyeRadene(r *http.Request) map[string]string {
	ut := map[string]string{}
	for nykel, verdiar := range r.Form {
		if !strings.HasPrefix(nykel, "namn-ny") || len(verdiar) == 0 {
			continue
		}
		namn := reintNamn(verdiar[0])
		if namn == "" {
			continue
		}
		ut[strings.TrimPrefix(nykel, "namn-")] = namn
	}
	return ut
}
