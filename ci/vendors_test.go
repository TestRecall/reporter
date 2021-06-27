package ci_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/testrecall/reporter/ci"
)

func TestUnknownCI(t *testing.T) {
	os.Clearenv()

	_, found := ci.GetVendor()
	assert.False(t, found)
}
func TestCircleCIBranch(t *testing.T) {
	os.Clearenv()
	setEnv(t, "CIRCLECI", "true")
	setEnv(t, "CIRCLE_BRANCH", "master")
	setEnv(t, "CIRCLE_SHA1", "a177f0f40b26f6196bb972aae3b7c171cdcffed7")
	setEnv(t, "CIRCLE_BUILD_NUM", "18")
	setEnv(t, "CIRCLE_BUILD_URL", "https://circleci.com/gh/KlotzAndrew/tre/18")
	// setEnv(t, "CIRCLE_PULL_REQUEST", "") not set

	vendor, found := ci.GetVendor()
	assert.True(t, found)
	assert.Equal(t, "CircleCI", vendor.GetName())
	assert.Equal(t, "a177f0f40b26f6196bb972aae3b7c171cdcffed7", vendor.GetSHA())
	assert.Equal(t, "18", vendor.GetBuildNumber())
	assert.Equal(t, "https://circleci.com/gh/KlotzAndrew/tre/18", vendor.GetBuildURL())
	assert.Equal(t, "master", vendor.GetBranch())
}

func TestCircleCIPR(t *testing.T) {
	const branch = "pull/3063"
	const name = "CircleCI"
	const sha = "d47fcf2cd85a9c48d1ea12ee8b76c8524e4d2044"
	const buildNumber = "62372"
	const buildURL = "https://circleci.com/gh/cBioPortal/cbioportal-frontend/62372"

	os.Clearenv()
	setEnv(t, "CIRCLECI", "true")
	setEnv(t, "CIRCLE_BRANCH", branch)
	setEnv(t, "CIRCLE_SHA1", sha)
	setEnv(t, "CIRCLE_BUILD_NUM", buildNumber)
	setEnv(t, "CIRCLE_BUILD_URL", buildURL)
	setEnv(t, "CIRCLE_PULL_REQUEST", "https://github.com/cBioPortal/cbioportal-frontend/pull/3063")

	vendor, found := ci.GetVendor()
	assert.True(t, found)
	assert.Equal(t, name, vendor.GetName())
	assert.Equal(t, sha, vendor.GetSHA())
	assert.Equal(t, buildNumber, vendor.GetBuildNumber())
	assert.Equal(t, buildURL, vendor.GetBuildURL())
	assert.Equal(t, branch, vendor.GetBranch())
}

func TestGitlabBranch(t *testing.T) {
	const name = "Gitlab"
	const branch = "master"
	const sha = "d47fcf2cd85a9c48d1ea12ee8b76c8524e4d2044"
	const buildNumber = "7046507"
	const buildURL = "https://gitlab.com/gitlab-examples/ci-debug-trace/-/jobs/379424655"

	os.Clearenv()
	setEnv(t, "GITLAB_CI", "true")
	setEnv(t, "CI_COMMIT_REF_NAME", branch)
	setEnv(t, "CI_COMMIT_SHA", sha)
	setEnv(t, "CI_JOB_ID", buildNumber)
	setEnv(t, "CI_JOB_URL", buildURL)
	// setEnv(t, "CIRCLE_PULL_REQUEST", "") not set

	vendor, found := ci.GetVendor()
	assert.True(t, found)
	assert.Equal(t, name, vendor.GetName())
	assert.Equal(t, sha, vendor.GetSHA())
	assert.Equal(t, buildNumber, vendor.GetBuildNumber())
	assert.Equal(t, buildURL, vendor.GetBuildURL())
	assert.Equal(t, branch, vendor.GetBranch())
}

func TestGitlabMR(t *testing.T) {
	t.Skip()
}

func TestTravisBranch(t *testing.T) {
	os.Clearenv()
	setEnv(t, "TRAVIS", "true")
	setEnv(t, "TRAVIS_BRANCH", "master")
	setEnv(t, "TRAVIS_COMMIT", "a177f0f40b26f6196bb972aae3b7c171cdcffed7")
	setEnv(t, "TRAVIS_BUILD_NUMBER", "30")
	setEnv(t, "TRAVIS_BUILD_WEB_URL", "https://travis-ci.com/KlotzAndrew/tre/builds/149181042")
	setEnv(t, "TRAVIS_PULL_REQUEST", "false")

	vendor, found := ci.GetVendor()
	assert.True(t, found)
	assert.Equal(t, "TravisCI", vendor.GetName())
	assert.Equal(t, "a177f0f40b26f6196bb972aae3b7c171cdcffed7", vendor.GetSHA())
	assert.Equal(t, "30", vendor.GetBuildNumber())
	assert.Equal(t, "https://travis-ci.com/KlotzAndrew/tre/builds/149181042", vendor.GetBuildURL())
	assert.Equal(t, "master", vendor.GetBranch())
}

func TestTravisPR(t *testing.T) {
	const branch = "tbranch"
	const name = "TravisCI"
	const sha = "fcff3fa2c0ef8f1c6e113ce8c338681cdeb48f85"
	const buildNumber = "30"
	const buildURL = "https://travis-ci.com/KlotzAndrew/tre/builds/149181042"

	os.Clearenv()
	setEnv(t, "TRAVIS", "true")
	setEnv(t, "TRAVIS_BRANCH", "master")
	setEnv(t, "TRAVIS_PULL_REQUEST_BRANCH", branch)
	setEnv(t, "TRAVIS_COMMIT", "wrong-4cbe8bfee972e36592e8b037253a09ed45")
	setEnv(t, "TRAVIS_PULL_REQUEST_SHA", sha)
	setEnv(t, "TRAVIS_BUILD_NUMBER", buildNumber)
	setEnv(t, "TRAVIS_BUILD_WEB_URL", buildURL)
	setEnv(t, "TRAVIS_PULL_REQUEST", "1")

	vendor, found := ci.GetVendor()
	assert.True(t, found)
	assert.Equal(t, name, vendor.GetName())
	assert.Equal(t, sha, vendor.GetSHA())
	assert.Equal(t, buildNumber, vendor.GetBuildNumber())
	assert.Equal(t, buildURL, vendor.GetBuildURL())
	assert.Equal(t, branch, vendor.GetBranch())
}

func TestGitlab(t *testing.T) {
	t.Skip("tbd")
}

func TestGithubAtions(t *testing.T) {
	t.Skip("tbd")
}

// https://github.com/jenkinsci/ghprb-plugin
func TestJenkinsGHBranch(t *testing.T) {
	const name = "Jenkins"
	const branch = "tbranch"
	const sha = "fcff3fa2c0ef8f1c6e113ce8c338681cdeb48f85"
	const buildNumber = "30"
	const buildURL = "https://jenkins.io/KlotzAndrew/tre/builds/149181042"

	os.Clearenv()
	setEnv(t, "JENKINS_URL", "https://jenkins.io/job/tr/PR-123")
	setEnv(t, "ghprbSourceBranch", branch) // GIT_BRANCH, BRANCH_NAME
	setEnv(t, "ghprbActualCommit", sha)    // GIT_COMMIT
	setEnv(t, "ghprbPullId", buildNumber)  // BUILD_NUMBER
	setEnv(t, "ghprbPullLink", buildURL)   // BUILD_URL
	// setEnv(t, "ghprbPullId", "2300")       // CHANGE_ID

	vendor, found := ci.GetVendor()
	assert.True(t, found)
	assert.Equal(t, name, vendor.GetName())
	assert.Equal(t, sha, vendor.GetSHA())
	assert.Equal(t, buildNumber, vendor.GetBuildNumber())
	assert.Equal(t, buildURL, vendor.GetBuildURL())
	assert.Equal(t, branch, vendor.GetBranch())
}

func TestJenkinsBranch(t *testing.T) {
	const name = "Jenkins"
	const branch = "tbranch"
	const sha = "fcff3fa2c0ef8f1c6e113ce8c338681cdeb48f85"
	const buildNumber = "30"
	const buildURL = "https://jenkins.io/KlotzAndrew/tre/builds/149181042"

	os.Clearenv()
	setEnv(t, "JENKINS_URL", "https://jenkins.io/job/tr/PR-123")
	setEnv(t, "BRANCH_NAME", branch) // GIT_BRANCH
	setEnv(t, "GIT_COMMIT", sha)
	setEnv(t, "BUILD_NUMBER", buildNumber)
	setEnv(t, "BUILD_URL", buildURL)

	vendor, found := ci.GetVendor()
	assert.True(t, found)
	assert.Equal(t, name, vendor.GetName())
	assert.Equal(t, sha, vendor.GetSHA())
	assert.Equal(t, buildNumber, vendor.GetBuildNumber())
	assert.Equal(t, buildURL, vendor.GetBuildURL())
	assert.Equal(t, branch, vendor.GetBranch())
}

func setEnv(t *testing.T, key, value string) {
	assert.NoError(t, os.Setenv(key, value))
}
