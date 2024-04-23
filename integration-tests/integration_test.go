package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/testrecall/reporter/reporter"
	"golang.org/x/sys/unix"
)

func TestUpload(t *testing.T) {
	expected := reporter.RequestData{
		Branch:      "master",
		CIName:      "Gitlab",
		SHA:         "sha123",
		BuildNumber: "1",
		BuildURL:    "https://localhost:7788",
	}

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	router := http.NewServeMux()
	router.HandleFunc("/runs", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		params := new(reporter.RequestData)
		err = json.Unmarshal(body, params)
		assert.NoError(t, err)

		idempotenceKey := r.Header.Get(reporter.IdempotencyKeyHeader)
		assert.Regexp(t, regexp.MustCompile("^[a-zA-Z0-9-_]{20,40}"), idempotenceKey, idempotenceKey)

		assert.Equal(t, expected.Branch, params.Branch)
		w.WriteHeader(http.StatusCreated)
	})
	server := &http.Server{
		Addr:     ":7788",
		Handler:  router,
		ErrorLog: logger,
	}
	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			assert.NoError(t, err)
		}
	}()
	defer func() {
		err := server.Shutdown(context.Background())
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	setEnv(t, "TR_SITE", "http://0.0.0.0:7788")
	setEnv(t, "TR_UPLOAD_TOKEN", "123")

	setEnv(t, "GITLAB_CI", "true")
	setEnv(t, "CI_COMMIT_REF_NAME", "master")
	setEnv(t, "CI_COMMIT_SHA", "sha123")
	setEnv(t, "CI_JOB_ID", "1")
	setEnv(t, "CI_JOB_URL", "https://localhost:7788")

	u := unix.Utsname{}
	unix.Uname(&u)
	machine := strings.Trim(string(u.Machine[:]), "\x00")

	executable_path := "./dist/linux_linux_amd64_v1/reporter"
	if machine == "arm64" {
		executable_path = "./dist/macos_darwin_arm64/reporter"
	}

	fmt.Println("mahein is", machine)
	fmt.Println("executable_path is", executable_path)

	out, err := runCmd("..", fmt.Sprintf("%s -file integration-tests/fixtures/small.xml -debug true", executable_path))
	assert.NoError(t, err, string(out))
}

func runCmd(dir, c string) ([]byte, error) {
	args := strings.Split(c, " ")
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}

func setEnv(t *testing.T, key, value string) {
	assert.NoError(t, os.Setenv(key, value))
}
