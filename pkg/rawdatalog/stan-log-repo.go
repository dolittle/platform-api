package rawdatalog

import (
	"encoding/json"

	"github.com/nats-io/stan.go"
)

type stanLogRepo struct {
	sc stan.Conn
}

func NewStanLogRepo(sc stan.Conn) Repo {
	return &stanLogRepo{
		sc: sc,
	}
}

// topic == subject == stream
func (r *stanLogRepo) Write(topic string, moment RawMoment) error {
	msg, _ := json.Marshal(moment)
	err := r.sc.Publish(topic, msg)
	if err != nil {
		return err
	}
	return nil
}
