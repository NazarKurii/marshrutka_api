package google

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maryan_api/config"
	rfc7807 "maryan_api/pkg/problem"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserInfoOAUTH struct {
	FirstName   string
	LastName    string
	Email       string
	DateOfBirth time.Time
}

var cfg = oauth2.Config{
	ClientID:     config.GoogleClientID(),
	ClientSecret: config.GoogleSecretKey(),
	RedirectURL:  config.FrontendURL(),
	Endpoint:     google.Endpoint,
	Scopes:       []string{"profile", "email", "https://www.googleapis.com/auth/user.birthday.read"},
}

func GetCredentialsByCode(code string, ctx context.Context, client *http.Client) (UserInfoOAUTH, error) {
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		fmt.Println(err.Error())
		return UserInfoOAUTH{}, invalidCode(err)
	}

	clientOAUTH := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(token),
			Base:   client.Transport,
		},
		Timeout: client.Timeout,
	}

	res, err := clientOAUTH.Get("https://people.googleapis.com/v1/people/me?personFields=birthdays,names,emailAddresses")
	if err != nil {

		return UserInfoOAUTH{}, badGateway(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {

		return UserInfoOAUTH{}, rfc7807.Internal("Failed reading Google response", err.Error())
	}

	var parsed struct {
		Names []struct {
			FamilyName string `json:"familyName"`
			GivenNAme  string `json:"givenName"`
		} `json:"names"`
		EmailAddresses []struct {
			Value string `json:"value"`
		} `json:"emailAddresses"`
		Birthdays []struct {
			Date struct {
				Year  int `json:"year"`
				Month int `json:"month"`
				Day   int `json:"day"`
			} `json:"date"`
		} `json:"birthdays"`
	}

	err = json.Unmarshal(body, &parsed)
	if err != nil {
		return UserInfoOAUTH{}, rfc7807.Internal("Parsing Google Response Error", err.Error())
	}

	var dob time.Time
	if len(parsed.Birthdays) > 0 {
		d := parsed.Birthdays[0].Date
		dob = time.Date(d.Year, time.Month(d.Month), d.Day, 0, 0, 0, 0, time.UTC)
	} else {
		dob = time.Now().AddDate(-18, 0, 0)
	}

	return UserInfoOAUTH{
		FirstName:   parsed.Names[0].GivenNAme,
		LastName:    parsed.Names[0].FamilyName,
		Email:       parsed.EmailAddresses[0].Value,
		DateOfBirth: dob,
	}, nil
}
