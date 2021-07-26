package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/red-gold/telar-core/pkg/log"
)

// GitHub provider
type GitHub struct {
	Client *http.Client
}

// NewGitHub create a new GitHub API provider
func NewGitHub(c *http.Client) *GitHub {
	return &GitHub{
		Client: c,
	}
}

// GetProfile returns a profile for a user from GitHub
func (gh *GitHub) GetProfile(accessToken string) (*Profile, error) {
	var err error
	var githubProfile GitHubProfile
	profile := &Profile{}

	req, reqErr := http.NewRequest(http.MethodGet, "https://api.github.com/user", nil)
	req.Header.Add("Authorization", "token "+accessToken)
	if reqErr != nil {
		return profile, reqErr
	}

	res, err := gh.Client.Do(req)
	if err != nil {
		return profile, reqErr
	}

	if res.StatusCode != http.StatusOK {
		return profile, fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	if res.Body != nil {
		defer res.Body.Close()

		bytesOut, _ := ioutil.ReadAll(res.Body)
		unmarshalErr := json.Unmarshal(bytesOut, &githubProfile)
		if unmarshalErr != nil {
			return profile, unmarshalErr
		}
	}
	log.Info(" github user unmarshal %v \n", githubProfile)
	profile.TwoFactor = githubProfile.TwoFactor
	profile.Name = githubProfile.Name
	profile.Email = githubProfile.Email
	profile.Avatar = githubProfile.Avatar
	profile.CreatedAt = githubProfile.CreatedAt
	profile.Login = githubProfile.Login
	profile.ID = fmt.Sprint(githubProfile.ID)

	if profile.Email == "" {
		email, emailErr := gh.GetGithubEmail(accessToken)
		if emailErr != nil {
			return profile, emailErr
		}
		profile.Email = email
	}

	if profile.Name == "" {
		profile.Name = strings.Split(profile.Email, "@")[0]
	}
	return profile, err
}

// GetGithubEmail get user email from github
func (gh *GitHub) GetGithubEmail(accessToken string) (string, error) {
	var err error
	var githubEmails []GitHubEmail

	req, reqErr := http.NewRequest(http.MethodGet, "https://api.github.com/user/emails", nil)
	req.Header.Add("Authorization", "token "+accessToken)
	if reqErr != nil {
		return "", reqErr
	}

	res, err := gh.Client.Do(req)
	if err != nil {
		return "", reqErr
	}

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code: %d", res.StatusCode)
	}

	if res.Body != nil {
		defer res.Body.Close()

		bytesOut, _ := ioutil.ReadAll(res.Body)
		unmarshalErr := json.Unmarshal(bytesOut, &githubEmails)
		if unmarshalErr != nil {
			return "", unmarshalErr
		}
	}

	if githubEmails != nil {
		for _, email := range githubEmails {
			if email.Primary {
				return email.Email, nil
			}
		}
	}

	return "", fmt.Errorf("No primary email found in Github.")
}

// GitHubProfile represents a GitHub profile
type GitHubProfile struct {
	ID        int       `json:"id"`
	Login     string    `json:"login"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar_url"`
	TwoFactor bool      `json:"two_factor_authentication"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GitHubEmail represents an email registered in Github
type GitHubEmail struct {
	Verified bool   `json:"verified"`
	Primary  bool   `json:"primary"`
	Email    string `json:"email"`
}
