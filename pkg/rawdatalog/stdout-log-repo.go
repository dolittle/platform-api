package rawdatalog

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
)

type stdoutLogRepo struct {
	logContext logrus.FieldLogger
}

func NewStdoutLogRepo(logContext logrus.FieldLogger) Repo {
	return &stdoutLogRepo{
		logContext: logContext,
	}
}

func (r *stdoutLogRepo) Write(topic string, moment RawMoment) error {
	data, _ := json.Marshal(moment)
	fmt.Println(string(data))
	//r.logContext.WithFields(logrus.Fields{
	//	"topic":    topic,
	//	"raw-data": string(data),
	//}).Info("entry")
	return nil
}
