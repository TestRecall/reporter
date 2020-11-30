package reporter

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/sirupsen/logrus"
)

func newClient(logger *logrus.Logger) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.Logger = logger
	client.Backoff = retryablehttp.LinearJitterBackoff
	client.RetryWaitMin = 800 * time.Millisecond
	client.RetryWaitMax = 1200 * time.Millisecond
	client.RetryMax = 4
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler
	return client
}

type sender struct {
	client *retryablehttp.Client
	Logger *logrus.Logger
}

func NewSender(logger *logrus.Logger) sender {
	return sender{newClient(logger), logger}
}

func (s sender) Send(remoteURL string, payload RequestPayload) error {
	jsonValue, err := json.Marshal(payload.RequestData)
	if err != nil {
		s.Logger.Fatalln(err)
	}
	s.Logger.Debug("outgoing data: ", string(jsonValue))

	url := remoteURL + "/runs"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		s.Logger.Fatalln("err posting")
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add(IdempotencyKeyHeader, payload.IdempotencyKey)

	req.SetBasicAuth("", payload.UploadToken)

	rreq, err := retryablehttp.FromRequest(req)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(rreq)
	if err != nil {
		s.Logger.Errorln(err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			s.Logger.Errorln(err, resp.StatusCode)
		}
		s.Logger.Debugln(string(body))
		return errors.New(string(body))
	}

	return err
}
