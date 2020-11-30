package ci

import "os"

type IVendor interface {
	GetName() string
	GetSHA() string
	GetBuildNumber() string
	GetBuildURL() string
	GetBranch() string

	Active() bool
}

type Vendor struct {
	Name string
	Env  string

	Branch      string
	SHA         string
	BuildNumber string
	BuildURL    string
	PullRequest string
}

func (v Vendor) Active() bool           { _, found := os.LookupEnv(v.Env); return found }
func (v Vendor) GetName() string        { return v.Name }
func (v Vendor) GetSHA() string         { return os.Getenv(v.SHA) }
func (v Vendor) GetBuildNumber() string { return os.Getenv(v.BuildNumber) }
func (v Vendor) GetBuildURL() string    { return os.Getenv(v.BuildURL) }
func (v Vendor) GetBranch() string      { return os.Getenv(v.Branch) }

type Jenkins struct {
	Name string
	Env  string

	Branch      []string
	SHA         []string
	BuildNumber []string
	BuildURL    []string
	PullRequest []string
}

func (v Jenkins) Active() bool           { _, found := os.LookupEnv(v.Env); return found }
func (v Jenkins) GetName() string        { return v.Name }
func (v Jenkins) GetSHA() string         { return guessEnv(v.SHA) }
func (v Jenkins) GetBuildNumber() string { return guessEnv(v.BuildNumber) }
func (v Jenkins) GetBuildURL() string    { return guessEnv(v.BuildURL) }
func (v Jenkins) GetBranch() string      { return guessEnv(v.Branch) }

func guessEnv(envs []string) string {
	for _, env := range envs {
		if v := os.Getenv(env); v != "" {
			return v
		}
	}
	return ""
}

type Travis struct {
	Name string
	Env  string

	Branch      string
	BranchPR    string
	SHA         string
	SHAPR       string
	BuildNumber string
	BuildURL    string
	PullRequest string
}

func (v Travis) Active() bool    { _, found := os.LookupEnv(v.Env); return found }
func (v Travis) GetName() string { return v.Name }
func (v Travis) GetSHA() string {
	if os.Getenv(v.PullRequest) == "false" {
		return os.Getenv(v.SHA)
	}
	return os.Getenv(v.SHAPR)
}
func (v Travis) GetBuildNumber() string { return os.Getenv(v.BuildNumber) }
func (v Travis) GetBuildURL() string    { return os.Getenv(v.BuildURL) }
func (v Travis) GetBranch() string {
	if os.Getenv(v.PullRequest) == "false" {
		return os.Getenv(v.Branch)
	}
	return os.Getenv(v.BranchPR)
}

var vendors = []IVendor{
	Vendor{
		Name:        "CircleCI",
		Env:         "CIRCLECI", // true if circle
		Branch:      "CIRCLE_BRANCH",
		SHA:         "CIRCLE_SHA1",
		BuildNumber: "CIRCLE_BUILD_NUM",
		BuildURL:    "CIRCLE_BUILD_URL",
		PullRequest: "CIRCLE_PULL_REQUEST", // url of pull request
	},
	Vendor{
		Name:        "Gitlab",
		Env:         "GITLAB_CI",
		Branch:      "CI_COMMIT_REF_NAME", // CI_BUILD_REF_NAME
		SHA:         "CI_COMMIT_SHA",      // CI_BUILD_REF
		BuildNumber: "CI_JOB_ID",          // CI_BUILD_ID
		BuildURL:    "CI_JOB_URL",
		PullRequest: "CI_COMMIT_BEFORE_SHA", // url of pull request
	},
	Vendor{
		Name:        "GithubAtions",
		Env:         "GITHUB_ACTIONS",
		Branch:      "GITHUB_REF",        // CI_BUILD_REF_NAME
		SHA:         "GITHUB_SHA",        // CI_BUILD_REF
		BuildNumber: "GITHUB_RUN_NUMBER", // CI_BUILD_ID
		BuildURL:    "GITHUB_API_URL",
		PullRequest: "", // url of pull request
	},
	Jenkins{
		Name:        "Jenkins",
		Env:         "JENKINS_URL",
		Branch:      []string{"ghprbSourceBranch", "BRANCH_NAME"}, // CI_BUILD_REF_NAME
		SHA:         []string{"ghprbActualCommit", "GIT_COMMIT"},  // CI_BUILD_REF
		BuildNumber: []string{"ghprbPullId", "BUILD_NUMBER"},      // CI_BUILD_ID
		BuildURL:    []string{"ghprbPullLink", "BUILD_URL"},
		PullRequest: []string{"ghprbPullId"}, // url of pull request
	},
	Travis{
		Name:        "TravisCI",
		Env:         "TRAVIS",
		Branch:      "TRAVIS_BRANCH",
		BranchPR:    "TRAVIS_PULL_REQUEST_BRANCH",
		SHA:         "TRAVIS_COMMIT",
		SHAPR:       "TRAVIS_PULL_REQUEST_SHA",
		BuildNumber: "TRAVIS_BUILD_NUMBER",
		BuildURL:    "TRAVIS_BUILD_WEB_URL",
		PullRequest: "TRAVIS_PULL_REQUEST", // PR number, or 'false'
	},
}

func GetVendor() (IVendor, bool) {
	for _, vendor := range vendors {
		if active := vendor.Active(); active {
			return vendor, true
		}
	}
	return Vendor{}, false
}
