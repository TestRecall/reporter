package reporter_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testrecall-reporter/reporter"
)

const uploadToken = "abc123"

func TestSendOK(t *testing.T) {
	const fileName = "report.xml"

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "", username)
		assert.Equal(t, uploadToken, password)

		body, err := ioutil.ReadAll(r.Body)
		assert.NoError(t, err)

		data := new(reporter.RequestData)
		err = json.Unmarshal(body, data)
		assert.NoError(t, err)

		assert.Equal(t, reporter.RequestData{
			RunData:   [][]byte{[]byte("foo")},
			Filenames: []string{fileName},
		}, *data)

		w.WriteHeader(http.StatusCreated)
	})

	s, teardown := testingHTTPClient(h)
	defer teardown()

	sender := reporter.NewSender(testLogger())
	err := sender.Send(s.URL, reporter.RequestPayload{
		UploadToken: uploadToken,
		RequestData: reporter.RequestData{
			Filenames: []string{fileName},
			RunData:   [][]byte{[]byte("foo")},
		},
	})

	assert.NoError(t, err)
}

func TestSendRecover(t *testing.T) {
	counter := 0
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		counter++
		if counter <= 3 {
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("badbadbad"))
			assert.NoError(t, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})

	s, teardown := testingHTTPClient(h)
	defer teardown()

	sender := reporter.NewSender(testLogger())
	err := sender.Send(s.URL, reporter.RequestPayload{})

	assert.NoError(t, err, err)
	assert.Equal(t, 4, counter)
}

func TestSendErr(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte(""))
		assert.NoError(t, err)
	})

	s, teardown := testingHTTPClient(h)
	defer teardown()

	sender := reporter.NewSender(testLogger())
	err := sender.Send(s.URL, reporter.RequestPayload{})

	assert.Error(t, err)
}

func testingHTTPClient(handler http.Handler) (*httptest.Server, func()) {
	s := httptest.NewServer(handler)

	return s, s.Close
}
