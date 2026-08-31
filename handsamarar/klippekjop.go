package handsamarar

import (
	"fmt"

	"kjernekraft/models"
)

// The purchase section on the klippekort page.
//
// The packages used to be written into the HTML — twenty-two cards with
// name, price and "save this much" — while the database had its own five.
// Two truths about the same thing, and they disagreed: the HTML offered
// five clips for 1100 kroner where the database said 499. Anyone changing
// the price in Pricing changed nothing a customer could see.
//
// So everything comes from here now, and the categories are the ones that
// *exist*: the HTML had six, the database has two, and the four that did not
// exist could not be bought at all — their cards had no package id, so the
// button ended in "could not find package ID".
type Category struct {
	Name     string
	Key      string // trygg for id-ar og spurnadsstrengen
	Packages []models.KlippekortPackage
}

// Categories groups the packages in the order the database gives them.
// GetAllKlippekortPackages sorts on category and then price, so the grouping
// only has to follow that order.
func Categories(pakkar []models.KlippekortPackage) []Category {
	var ut []Category
	// The key ends up as an id on the page, and two identical ids is an id that
	// 	// does not point. Names differing only in punctuation — "Yoga" and
	// 	// "Yoga!" — become the same string, and a name without letters becomes
	// 	// nothing. So we number instead: the key must always point at one
	// 	// section.
	sedd := map[string]bool{}
	for _, p := range pakkar {
		if len(ut) == 0 || ut[len(ut)-1].Name != p.Category {
			n := slugify(p.Category)
			if n == "" {
				n = "bolk"
			}
			grunn := n
			for i := 2; sedd[n]; i++ {
				n = fmt.Sprintf("%s-%d", grunn, i)
			}
			sedd[n] = true
			ut = append(ut, Category{Name: p.Category, Key: n})
		}
		i := len(ut) - 1
		ut[i].Packages = append(ut[i].Packages, p)
	}
	return ut
}

// slugify turns a category name into something that survives standing in an
// id and in a query string. "Reformer/Apparatus" would otherwise become two
// path segments in a URL.
func slugify(s string) string {
	ut := make([]rune, 0, len(s))
	fyrre := '-'
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			ut = append(ut, r)
			fyrre = r
		case r >= 'A' && r <= 'Z':
			ut = append(ut, r+32)
			fyrre = r
		case r == 'æ' || r == 'Æ':
			ut = append(ut, 'a', 'e')
			fyrre = 'e'
		case r == 'ø' || r == 'Ø':
			ut = append(ut, 'o', 'e')
			fyrre = 'e'
		case r == 'å' || r == 'Å':
			ut = append(ut, 'a', 'a')
			fyrre = 'a'
		default:
			if fyrre != '-' {
				ut = append(ut, '-')
				fyrre = '-'
			}
		}
	}
	for len(ut) > 0 && ut[len(ut)-1] == '-' {
		ut = ut[:len(ut)-1]
	}
	return string(ut)
}
