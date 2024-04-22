package reporter

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	junit "github.com/joshdk/go-junit"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/testrecall/reporter/ci"
)

var branchCommand = strings.Fields("git log -n 1 --pretty=%D HEAD")

const MultiErrorMessage = `
valid values: before/partial/after

multi allows TestRecall to group multiple or parallel test reports together.
Otherwise each partial report will be considered its on build.

 - 'before' must be sent before any results are uploaded, to let TestRecall
know that multiple reports will be uploaded
 - 'partial' must be sent with any results that will be grouped together
 - 'after' must be called after all uploads are finished, to prevent
acidenitally combining extra files`

const noTokenMessage = `
	TR_UPLOAD_TOKEN must be set in the environment,
	find the token for this project here:
	app.testrecall.com/projects/<my-project>/integrations
`

func noFileMessage(filename string) string {
	return fmt.Sprintf(`
		unable to find file to upload results at: %s
		check if the path is correct and being passed to the reporter
	`, filename)
}

const (
	MultiBefore  = "before"
	MutliPartial = "partial"
	MultiAfter   = "after"
)

const IdempotencyKeyHeader = "Idempotency-Key"

type RequestPayload struct {
	IdempotencyKey string
	UploadToken    string
	Filename       string

	RequestData RequestData

	Logger *logrus.Logger

	Vendor ci.IVendor
}

type RequestData struct {
	RunData   [][]byte `json:"run"`
	Filenames []string `json:"file_names"`
	Multi     string   `json:"multi"`

	Hostname        string            `json:"hostname"`
	ReporterVersion string            `json:"reporter_version"`
	Flags           map[string]string `json:"flags"`

	Branch string `json:"branch"`
	SHA    string `json:"sha"`
	Tag    string `json:"tag"`
	PR     string `json:"pr"`

	Slug        string `json:"slug"`
	CIName      string `json:"ci_name"`
	BuildNumber string `json:"build_number"`
	BuildURL    string `json:"build_url"`
	Job         string `json:"job"`
}

func (r *RequestPayload) Setup() {
	r.IdempotencyKey = newIdempotencyKey()

	r.GetUploadToken()
	r.GetHostname()
	r.GetRunData()

	r.GetVendor()

	r.GetSHA()
	r.GetBranch()
	r.GetBuildNumber()
	r.GetBuildURL()
}

func newIdempotencyKey() string {
	now := time.Now().UnixNano()
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return fmt.Sprintf("%v_%v", now, base64.URLEncoding.EncodeToString(buf)[:6])
}

func (r RequestPayload) isVendorKnown() bool {
	return r.RequestData.CIName != ""
}

func (r RequestPayload) FailureCount() (int, bool) {
	if mutli, _ := r.isMulti(); mutli {
		return 0, true
	}

	run, err := junit.Ingest(r.RequestData.RunData[0])
	if err != nil {
		r.Logger.Debug(err)
		return 0, false
	}

	count := 0
	for _, s := range run {
		count += s.Totals.Failed
	}
	return count, true
}

func (r *RequestPayload) GetVendor() {
	if vendor, found := ci.GetVendor(); found {
		r.Vendor = vendor
		r.RequestData.CIName = r.Vendor.GetName()
	}
}

func (r *RequestPayload) GetUploadToken() {
	r.UploadToken = os.Getenv("TR_UPLOAD_TOKEN")
	if r.UploadToken == "" {
		r.Logger.Fatal(noTokenMessage)
	}
}

func (r *RequestPayload) GetBranch() {
	if r.RequestData.Branch != "" {
		return
	}

	if r.isVendorKnown() {
		r.RequestData.Branch = r.Vendor.GetBranch()
		if r.RequestData.Branch != "" {
			return
		}
	}

	if _, err := exec.LookPath("git"); err != nil {
		r.Logger.Fatal("-sha is a required field")
	}

	// NOTE: ci may be in a detached head
	out, err := exec.Command(branchCommand[0], branchCommand[1:]...).CombinedOutput()
	r.Logger.Debugln("branch: ", string(out))
	if err != nil {
		r.Logger.Fatal("git error checking for detached head ", err)
	}
	rawOut := string(out)
	r.RequestData.Branch = GitBranchFromInfo(rawOut)

	r.Logger.Debugln(r.RequestData.Branch, rawOut)
}

func GitBranchFromInfo(info string) string {
	trimmed := strings.TrimSuffix(info, "\n")

	forward := strings.Split(trimmed, "->")
	var branch string
	if len(forward) >= 2 {
		branches := strings.Split(forward[1], ",")
		branch = branches[0]
	} else {
		branches := strings.Split(trimmed, ",")
		branch = branches[len(branches)-1]
	}
	return strings.TrimSpace(branch)
}

func (r *RequestPayload) GetSHA() {
	if r.RequestData.SHA != "" {
		return
	}

	if r.isVendorKnown() {
		r.RequestData.SHA = r.Vendor.GetSHA()
		if r.RequestData.SHA != "" {
			return
		}
	}

	if _, err := exec.LookPath("git"); err != nil {
		r.Logger.Fatal("-sha is a required field")
	}

	out, err := exec.Command("git", "rev-parse", "HEAD").CombinedOutput()
	if err != nil {
		r.Logger.Fatal("git error using rev-parse", err)
	}
	rawOut := string(out)
	rawOut = strings.TrimSuffix(rawOut, "\n")
	r.RequestData.SHA = rawOut
}

func (r *RequestPayload) GetHostname() {
	if r.RequestData.Hostname != "" {
		return
	}

	h, err := os.Hostname()
	if err != nil {
		r.Logger.Fatal("unable to detect hostname", err)
	}
	r.RequestData.Hostname = h
}

func (r *RequestPayload) GetBuildNumber() {
	if r.RequestData.BuildNumber != "" {
		return
	}

	r.Logger.Debugf("vendor: %v", r.isVendorKnown())
	if r.isVendorKnown() {
		r.Logger.Debugf("vendor build number: %v", r.Vendor.GetBuildNumber())
		r.RequestData.BuildNumber = r.Vendor.GetBuildNumber()
		if r.RequestData.BuildNumber != "" {
			return
		}
	}
}

func (r *RequestPayload) GetBuildURL() {
	if r.RequestData.BuildURL != "" {
		return
	}

	if r.isVendorKnown() {
		r.RequestData.BuildURL = r.Vendor.GetBuildURL()
		if r.RequestData.BuildURL != "" {
			return
		}
	}
}

func (r *RequestPayload) isMulti() (bool, error) {
	multi := r.RequestData.Multi
	switch multi {
	case "":
		return false, nil
	case MultiBefore, MutliPartial, MultiAfter:
		return true, nil
	default:
		return false, errors.New(multi)
	}
}

func (r *RequestPayload) GetRunData() {
	multi, err := r.isMulti()
	if err != nil {
		r.Logger.Fatal(err, MultiErrorMessage)
	}

	fs := afero.NewOsFs()
	files, err := SearchReportFiles(fs, r.Filename)
	if err != nil {
		r.Logger.Fatal(err)
	}
	r.RequestData.Filenames = files

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			r.Logger.Fatal(err)
		}
		r.RequestData.RunData = append(r.RequestData.RunData, data)
	}

	if r.Filename == "" && len(r.RequestData.RunData) == 0 {
		// multi allows uploading no data
		if !multi {
			r.Logger.Fatal("-file is a required field")
		}
	}
}

var defaultPatterns = []string{
	"./*/*/TEST-*.xml",
	"./*/*/*/TEST-*.xml",
	"./*/*/*/*/TEST-*.xml",
	"./*/*/*/*/*/TEST-*.xml",

	"junit*.xml",
	"rspec*.xml",
	"report*.xml",

	"./reports/junit*.xml",
	"./reports/rspec*.xml",
	"./reports/report*.xml",

	"./test-results/junit*.xml",
	"./test-results/rspec*.xml",
	"./test-results/report*.xml",

	"/tmp/test-results/junit*.xml",
	"/tmp/test-results/rspec*.xml",
	"/tmp/test-results/report*.xml",
}

func SearchReportFiles(fs afero.Fs, pattern string) ([]string, error) {
	// check for suplied pattern
	if pattern != "" {
		files, err := afero.Glob(fs, pattern)
		if err != nil {
			return []string{}, err
		}

		if len(files) == 0 {
			return []string{}, errors.New(noFileMessage(pattern))
		}

		return files, nil
	}

	for _, p := range defaultPatterns {
		if matched, err := afero.Glob(fs, p); err != nil || len(matched) > 0 {
			return matched, err
		}
	}

	// no files found
	return []string{}, errors.New(noFileMessage(pattern))
}
