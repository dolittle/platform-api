package rawdatalog

import (
	"encoding/json"
	"fmt"
)

type stdoutLogRepo struct{}

func NewStdoutLogRepo() Repo {
	return &stdoutLogRepo{}
}

func (r *stdoutLogRepo) Write(topic string, moment RawMoment) error {
	data, _ := json.Marshal(moment)
	fmt.Println(string(data))
	return nil
}
