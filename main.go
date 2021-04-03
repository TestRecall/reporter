package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/testrecall-reporter/reporter"
)

var (
	// updated at compile
	RemoteURL = "http://0.0.0.0:1323"
	Version   = "unknown"
	Date      = "unknown"
	Commit    = "unknown"

	printVersion = flag.Bool("version", false, "print version")

	debug       = flag.Bool("debug", false, "debug log level")
	setExitCode = flag.String("setExitCode", "", "[true]/false', exits 1 if tests failed")
	multi       = flag.String("multi", "", reporter.MultiErrorMessage)

	junitFile = flag.String("file", "", "junit file")
	hostName  = flag.String("host", "", "host name")

	gitBranch = flag.String("branch", "", "git branch")
	gitSHA    = flag.String("sha", "", "git sha")
	gitTag    = flag.String("tag", "", "git tag")
	isPr      = flag.String("pr", "", "true/false/[unknown] is git PR")

	slug        = flag.String("slug", "", "repo slug")
	ciName      = flag.String("ciName", "", "ci runner name")
	buildNumber = flag.String("buildnumber", "", "build number for labeling runs")
	buildURL    = flag.String("buildurl", "", "build url to link back to")
	job         = flag.String("job", "", "build url to link back to")
)

func main() {
	flag.Parse()

	if *printVersion {
		fmt.Printf("Version: %s\nCommit: %s\nBuilt at: %s\n", Version, Commit, Date)
		os.Exit(0)
	}

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})

	if *debug {
		logger.Level = logrus.TraceLevel
	} else {
		logger.Level = logrus.InfoLevel
	}

	flags := map[string]string{}
	flag.VisitAll(func(f *flag.Flag) {
		value := f.Value.String()
		if value != "" {
			flags[f.Name] = value
		}
	})

	payload := reporter.RequestPayload{
		Filename:    *junitFile,
		UploadToken: "",

		RequestData: reporter.RequestData{
			RunData:   [][]byte{},
			Filenames: []string{*junitFile},
			Multi:     *multi,

			Hostname:        *hostName,
			ReporterVersion: Version + "-" + Commit,
			Flags:           flags,

			Branch: *gitBranch,
			SHA:    *gitSHA,
			Tag:    *gitTag,
			PR:     *isPr,

			Slug:        *slug,
			CIName:      *ciName,
			BuildNumber: *buildNumber,
			BuildURL:    *buildURL,
			Job:         *job,
		},

		Logger: logger,
	}

	payload.Setup()

	url := RemoteURL
	if newURL, found := os.LookupEnv("TR_SITE"); found {
		url = newURL
	}

	sender := reporter.NewSender(logger)
	if err := sender.Send(url, payload); err != nil {
		logger.Debug("upload failed!")
		logger.Fatalln(err, payload.RequestData)
	}
	logger.Debug("upload success!")

	fails, xmlValid := payload.FailureCount()
	if shouldExitOnFail(*setExitCode) {
		if !xmlValid {
			logger.Debugf("test xml is invalid")
			os.Exit(1)
		} else if fails > 0 {
			logger.Debugf("exiting with failed tests: %v", fails)
			os.Exit(1)
		}
	}
}

func shouldExitOnFail(s string) bool {
	if s == "false" || s == "f" {
		return false
	}
	return true
}
