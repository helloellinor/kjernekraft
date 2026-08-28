package handlers

import (
	"testing"
	"time"

	"kjernekraft/handlers/config"
	"kjernekraft/models"
)

// Ein rabatt som hev gjenge ut sluttar aa gjelda i det han gjeng ut.
//
// Datoen vert prøvd kvar gong nokon spør, og ikkje av ein jobb som skal
// hugsa aa slokka honom. Ein jobb som ikkje køyrde hadde late rabatten
// staa — og det er den slags feil ingen oppdagar, av di ho ser ut som
// ingen ting.
func TestRabattenGjengUt(t *testing.T) {
	naa := config.GetInstance().GetCurrentTime()

	for _, p := range []struct {
		namn string
		u    models.User
		vil  bool
	}{
		{"gjeld enno", models.User{
			StudentSenior: true, StudentSeniorGjengUt: naa.AddDate(1, 0, 0)}, true},
		{"gjekk ut i gaar", models.User{
			StudentSenior: true, StudentSeniorGjengUt: naa.AddDate(0, 0, -1)}, false},
		{"utan utgang", models.User{
			StudentSenior: true}, true},
		{"aldri gjeven", models.User{}, false},
		{"honnør kjem av alderen og gjeng aldri ut", models.User{
			Birthdate: naa.AddDate(-70, 0, 0).Format("2006-01-02")}, true},
		{"for ung til honnør", models.User{
			Birthdate: naa.AddDate(-30, 0, 0).Format("2006-01-02")}, false},
	} {
		t.Run(p.namn, func(t *testing.T) {
			if fekk := Kvalifisert(&p.u); fekk != p.vil {
				t.Errorf("Kvalifisert = %v, venta %v", fekk, p.vil)
			}
		})
	}
}

// Ingen brukar er ingen rabatt.
func TestUtanBrukarIngenRabatt(t *testing.T) {
	if Kvalifisert(nil) {
		t.Error("ein som ikkje er logga paa fekk studentrabatt")
	}
}

// Grensa gjeng ved datoen, ikkje ved aaret.
func TestGrensaGjengVedDatoen(t *testing.T) {
	naa := config.GetInstance().GetCurrentTime()
	knapt := models.User{StudentSenior: true, StudentSeniorGjengUt: naa.Add(time.Hour)}
	nettupp := models.User{StudentSenior: true, StudentSeniorGjengUt: naa.Add(-time.Hour)}

	if !Kvalifisert(&knapt) {
		t.Error("ein time att, og rabatten gjeld ikkje")
	}
	if Kvalifisert(&nettupp) {
		t.Error("ein time yver, og rabatten gjeld framleis")
	}
}
