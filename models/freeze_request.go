package models

import (
	"database/sql"
	"time"
)

type FreezeRequest struct {
	MembershipID      int64          `json:"membership_id"`
	UserID            int64          `json:"user_id"`
	Status            string         `json:"status"`
	StartDate         string         `json:"start_date"`
	RenewalDate       string         `json:"renewal_date"`
	EndDate           sql.NullString `json:"end_date"`
	BindingEnd        sql.NullString `json:"binding_end"`
	LastBilled        sql.NullString `json:"last_billed"`
	CreatedAt         time.Time      `json:"created_at"`
	UserName          string         `json:"user_name"`
	UserEmail         string         `json:"user_email"`
	UserPhone         string         `json:"user_phone"`
	MembershipName    string         `json:"membership_name"`
	MembershipPrice   float64        `json:"membership_price"`
	CommitmentMonths  int            `json:"commitment_months"`
}