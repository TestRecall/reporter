package reporter_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/testrecall/reporter/reporter"
)

func testLogger() *logrus.Logger {
	l := logrus.New()
	l.Level = logrus.TraceLevel
	return l
}

func TestGitBranchFromInfo(t *testing.T) {
	for _, tt := range []struct {
		line string
		want string
	}{
		{line: "HEAD, master", want: "master"},
		{line: "HEAD -> master", want: "master"},
		{line: "(HEAD -> master, origin/master, origin/HEAD)", want: "master"},
		{line: "HEAD, origin/master, master", want: "master"},
	} {
		got := reporter.GitBranchFromInfo(tt.line)

		assert.Equal(t, tt.want, got, tt.line)
	}
}

func TestFailureCount(t *testing.T) {
	for _, tt := range []struct {
		filename string
		fails    int
		validXML bool
	}{
		{"rspec_success.xml", 0, true},
		{"rspec_malformed.xml", 0, false},
		{"golang_success.xml", 0, true},
		{"golang_fail.xml", 1, true},
	} {
		report := reporter.RequestPayload{
			Logger: testLogger(),
			RequestData: reporter.RequestData{
				RunData: [][]byte{getFixture(tt.filename)},
			},
		}
		fails, xmlOK := report.FailureCount()
		assert.Equal(t, tt.validXML, xmlOK)
		assert.Equal(t, tt.fails, fails)
	}
}

func getFixture(filename string) []byte {
	fp := filepath.Join("fixtures", filename)
	b, err := os.ReadFile(fp)
	if err != nil {
		fmt.Print(err)
	}

	return b
}

func TestSetup(t *testing.T) {
	os.Setenv("TR_UPLOAD_TOKEN", "abc123")

	payload := reporter.RequestPayload{
		Filename: "./fixtures/hello.txt",
		RequestData: reporter.RequestData{
			Branch:          "branch1",
			SHA:             "sha1",
			BuildNumber:     "10",
			BuildURL:        "https://testrecall.com",
			ReporterVersion: "10",
		},
		Logger: testLogger(),
	}
	payload.Setup()

	assert.Less(t, 1, len(payload.RequestData.ReporterVersion), payload.RequestData.ReporterVersion)
}

func gitConfig(t *testing.T, dir string) {
	out, err := runCmd(dir, `git config commit.gpgsign false`)
	assert.NoError(t, err, string(out))
	_, err = runCmd(dir, `git config user.email "you@example.com"`)
	assert.NoError(t, err)
	_, err = runCmd(dir, `git config user.name testuser`)
	assert.NoError(t, err)
	_, err = runCmd(dir, `git config --local -l`)
	assert.NoError(t, err)
}

func TestGetBranch(t *testing.T) {
	_, err := exec.LookPath("git")
	assert.NoError(t, err, "git needs to be installed")

	dir, cleanup := newRepo(t)
	defer cleanup()

	fmt.Println(dir)

	err = os.Chdir(dir)
	assert.NoError(t, err)
	gitConfig(t, dir)

	err = createBuzzFile(dir)
	assert.NoError(t, err)
	out, err := gitAdd(dir)
	assert.NoError(t, err, string(out))
	out, err = gitCommit(dir, "zero commit")
	fmt.Println(string(out))
	assert.NoError(t, err, string(out))

	out, _ = exec.Command("git", "log").CombinedOutput()
	fmt.Println(string(out))
	fmt.Println("master branch") // master branch
	assert.Equal(t, "master", reporterBranch())

	fmt.Println("detached head master") // detached head master
	sha, err := gitSha(dir)
	assert.NoError(t, err)

	out, err = gitCheckout(dir, strings.TrimSpace(string(sha)))
	assert.NoError(t, err, string(out))
	assert.Equal(t, "master", reporterBranch(), string(out))

	fmt.Println("new branch") // new branch
	out, err = gitNewBrach(dir, "foobranch")
	assert.NoError(t, err, string(out))

	out, err = gitBrach(dir)
	assert.NoError(t, err)
	assert.Equal(t, "foobranch", reporterBranch(), string(out))

	fmt.Println("detached head branch") // detached head branch
	err = createBarFile(dir)
	assert.NoError(t, err)
	out, err = gitAdd(dir)
	assert.NoError(t, err, string(out))
	out, err = gitCommit(dir, "second commit")
	assert.NoError(t, err, string(out))

	sha, err = gitSha(dir)
	assert.NoError(t, err)

	out, err = gitCheckout(dir, strings.TrimSpace(string(sha)))
	assert.NoError(t, err, string(out))
	assert.Equal(t, "foobranch", reporterBranch(), string(out))

	fmt.Println("detatched fork") // detatched fork
	newDir := dir + "_cloned"
	_, cleanupNewDir := gitClone(t, dir, newDir)
	defer cleanupNewDir()

	curDir, err := os.Getwd()
	assert.NoError(t, err)
	defer assert.NoError(t, os.Chdir(curDir))
	err = os.Chdir(newDir)
	assert.NoError(t, err)
	gitConfig(t, dir)

	out, err = gitDetachtedBranch(newDir, "master")
	assert.NoError(t, err, string(out))
	gitConfig(t, dir)

	out, err = gitBrach(newDir)
	assert.NoError(t, err)
	assert.Equal(t, "master", reporterBranch(), string(out))

}

func reporterBranch() string {
	fmt.Println("getting branch")
	payload := reporter.RequestPayload{Logger: testLogger()}
	payload.GetBranch()
	return payload.RequestData.Branch
}

func gitClone(t *testing.T, dir, newDir string) (string, func()) {
	cmd := exec.Command("git", "clone", dir, newDir)
	cmd.Dir = "/tmp"

	out, err := cmd.CombinedOutput()
	fmt.Println("out", string(out))
	assert.NoError(t, err, string(out))

	return dir, func() { os.RemoveAll(newDir) }
}

func newRepo(t *testing.T) (string, func()) {
	dir, cleanup := tempDir(t)

	out, err := gitInit(dir)
	assert.NoError(t, err, string(out))
	gitConfig(t, dir)

	err = createFooFile(dir)
	assert.NoError(t, err)

	out, err = gitAdd(dir)
	assert.NoError(t, err, string(out))

	out, err = gitCommit(dir, "first commit")
	assert.NoError(t, err, string(out))

	return dir, cleanup
}

func tempDir(t *testing.T) (string, func()) {
	dir, err := os.MkdirTemp("", "tr-git-tests-")
	assert.NoError(t, err)
	return dir, func() { os.RemoveAll(dir) }
}

func gitCheckout(dir, branch string) ([]byte, error) {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func gitSha(dir string) ([]byte, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func gitBrach(dir string) ([]byte, error) {
	cmd := exec.Command("git", "branch")
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func gitNewBrach(dir, branch string) ([]byte, error) {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func gitDetachtedBranch(dir, branch string) ([]byte, error) {
	cmd := exec.Command("git", "checkout", "-q", "-B", "master")
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func gitInit(dir string) ([]byte, error) {
	cmd := exec.Command("git", "init")
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func gitAdd(dir string) ([]byte, error) {
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func gitCommit(dir, msg string) ([]byte, error) {
	cmd := exec.Command("git", "commit", "-m", msg)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func createFooFile(dir string) error {
	d1 := []byte("hello\ngo\n")
	err := os.WriteFile(dir+"/foo.txt", d1, 0644)
	return err
}

func createBarFile(dir string) error {
	d1 := []byte("hello\ngo\n")
	err := os.WriteFile(dir+"/bar.txt", d1, 0644)
	return err
}

func createBuzzFile(dir string) error {
	d1 := []byte("hello\ngo\n")
	err := os.WriteFile(dir+"/buzz.txt", d1, 0644)
	return err
}

func runCmd(dir, c string) ([]byte, error) {
	args := strings.Split(c, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func TestGetRunData(t *testing.T) {
	d, err := os.MkdirTemp("", "s-")
	assert.NoError(t, err)
	defer os.RemoveAll(d)

	for _, tt := range []struct {
		file    string
		pattern string
		match   int
	}{
		{"/a/f.xml", "/a/f.xml", 1},
		{"/f.xml", "*", 1},
		{"/f.xml", "/*.xml", 1},
		{"/a/f.xml", "/a/*.xml", 1},
		{"/a/b/f.xml", "/*/*/*.xml", 1},
		{"/a/b/c/f-oo.xml", "/*/*/f*xml", 1},
		{"/a/b/c/d/f-oo.xml", "/*/*/*/*/f-*.xml", 1},
	} {
		file := d + tt.file
		pattern := d + tt.pattern
		fmt.Println(file, pattern)

		err := os.MkdirAll(filepath.Dir(file), 0755)
		assert.NoError(t, err)

		_, err = os.Create(file)
		assert.NoError(t, err)

		matches, err := reporter.SearchReportFiles(pattern)
		assert.NoError(t, err)

		if tt.match != len(matches) {
			t.Errorf("wanted %v got %v, with path %s pattern %s", tt.match, len(matches), file, pattern)
		}
	}
}
