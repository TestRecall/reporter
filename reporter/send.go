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

// https://github.com/hashicorp/go-retryablehttp/issues/93
type leveledLogrus struct{ *logrus.Logger }

func (l *leveledLogrus) fields(keysAndValues ...interface{}) map[string]interface{} {
	fields := make(map[string]interface{})
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		fields[keysAndValues[i].(string)] = keysAndValues[i+1]
	}
	return fields
}
func (l *leveledLogrus) Error(msg string, keysAndValues ...interface{}) {
	l.WithFields(l.fields(keysAndValues...)).Error(msg)
}
func (l *leveledLogrus) Info(msg string, keysAndValues ...interface{}) {
	l.WithFields(l.fields(keysAndValues...)).Info(msg)
}
func (l *leveledLogrus) Debug(msg string, keysAndValues ...interface{}) {
	l.WithFields(l.fields(keysAndValues...)).Debug(msg)
}
func (l *leveledLogrus) Warn(msg string, keysAndValues ...interface{}) {
	l.WithFields(l.fields(keysAndValues...)).Warn(msg)
}

func newClient(logger *logrus.Logger) *retryablehttp.Client {
	client := retryablehttp.NewClient()
	client.Logger = retryablehttp.LeveledLogger(&leveledLogrus{logger})
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
		return err
	}
	s.Logger.Debug("outgoing data: ", string(jsonValue))

	url := remoteURL + "/runs"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return err
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
