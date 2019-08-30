package upstreams

import (
	"os"
	"testing"

	"github.com/google/go-github/github"
	"gopkg.in/yaml.v3"
)

func TestGetClient(t *testing.T) {
	var client *github.Client
	currentAccessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	os.Unsetenv("GITHUB_ACCESS_TOKEN")
	client = getClient()
	if client == nil {
		t.Errorf("Could not get a Github client (without access token)")
	}
	os.Setenv("GITHUB_ACCESS_TOKEN", "test")
	client = getClient()
	if client == nil {
		t.Errorf("Could not get a Github client (with \"test\" access token)")
	}
	os.Setenv("GITHUB_ACCESS_TOKEN", currentAccessToken)
}

func TestUnserialiseGithub(t *testing.T) {
	validYamls := []string{
		"flavour: github\nurl: helm/helm\nconstraints: <1.0.0",
	}

	for _, valid := range validYamls {
		var u Github
		err := yaml.Unmarshal([]byte(valid), &u)
		if err != nil {
			t.Errorf("Failed to deserialise valid yaml:\n%s", valid)
		}
	}
}

func TestInvalidValues(t *testing.T) {
	var err error
	invalidURL := "test"
	gh := Github{
		URL: invalidURL,
	}
	_, err = gh.LatestVersion()
	if err == nil {
		t.Errorf("Should fail on invalid URL:\n%s", invalidURL)
	}

	invalidConstraint := "invalid-constraint"
	gh2 := Github{
		URL:         "test/test",
		Constraints: invalidConstraint,
	}
	_, err = gh2.LatestVersion()
	if err == nil {
		t.Errorf("Should fail on invalid Constraint:\n%s", invalidConstraint)
	}
}

func TestWrongRepository(t *testing.T) {
	gh := Github{
		URL: "Pluies/doesnotexist",
	}
	_, err := gh.LatestVersion()
	if err == nil {
		t.Errorf("Should fail on repository that does not exist")
	}
}

func TestHappyPath(t *testing.T) {
	gh := Github{
		URL: "helm/helm",
	}
	latestVersion, err := gh.LatestVersion()
	if err != nil {
		t.Errorf("Failed Github happy path test: %v", err)
	}
	if latestVersion == "" {
		t.Errorf("Got an empty latestVersion")
	}
}