package git

import (
	"encoding/json"
	"strings"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/storage"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

func (s *GitStorage) SaveBusinessMomentEntity(customerID string, input platform.HttpInputBusinessMomentEntity) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"customer_id":     customerID,
		"application_id":  input.ApplicationID,
		"environment":     input.Environment,
		"microservice_id": input.MicroserviceID,
		"entity_id":       input.Entity.EntityTypeID,
	})

	// Lookup the microservice
	rawBytes, err := s.GetMicroservice(customerID, input.ApplicationID, input.Environment, input.MicroserviceID)
	if err != nil {
		return storage.ErrNotFound
	}

	var microservice platform.HttpInputBusinessMomentAdaptorInfo
	err = json.Unmarshal(rawBytes, &microservice)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("unmarshall issues")
		return err
	}

	// Confirm its business moment
	if microservice.Kind != platform.MicroserviceKindBusinessMomentsAdaptor {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("not-business-moments-adaptor")
		return storage.ErrNotBusinessMomentsAdaptor
	}

	index := funk.IndexOf(microservice.Extra.Entities, func(entity platform.Entity) bool {
		return entity.EntityTypeID == input.Entity.EntityTypeID
	})

	if index != -1 {
		microservice.Extra.Entities[index] = input.Entity
	} else {
		microservice.Extra.Entities = append(microservice.Extra.Entities, input.Entity)
	}

	// I guess a possible race condition is possible.
	err = s.SaveMicroservice(customerID, input.ApplicationID, input.Environment, microservice.Dolittle.MicroserviceID, microservice)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.SaveMicroservice",
		}).Error("On saving microservice")
		return err
	}
	return nil
}

// Save all is cheaper
func (s *GitStorage) SaveBusinessMoment(customerID string, input platform.HttpInputBusinessMoment) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"customer_id":     customerID,
		"application_id":  input.ApplicationID,
		"environment":     input.Environment,
		"microservice_id": input.MicroserviceID,
		"moment_id":       input.Moment.UUID,
	})

	// Lookup the microservice
	rawBytes, err := s.GetMicroservice(customerID, input.ApplicationID, input.Environment, input.MicroserviceID)
	if err != nil {
		return storage.ErrNotFound
	}

	var microservice platform.HttpInputBusinessMomentAdaptorInfo
	err = json.Unmarshal(rawBytes, &microservice)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("unmarshall issues")
		return err
	}

	// Confirm its business moment
	if microservice.Kind != platform.MicroserviceKindBusinessMomentsAdaptor {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("not-business-moments-adaptor")
		return storage.ErrNotBusinessMomentsAdaptor
	}

	index := funk.IndexOf(microservice.Extra.Moments, func(moment platform.BusinessMoment) bool {
		return moment.UUID == input.Moment.UUID
	})

	if index != -1 {
		microservice.Extra.Moments[index] = input.Moment
	} else {
		microservice.Extra.Moments = append(microservice.Extra.Moments, input.Moment)
	}

	err = s.SaveMicroservice(customerID, input.ApplicationID, input.Environment, microservice.Dolittle.MicroserviceID, microservice)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.SaveMicroservice",
		}).Error("On saving microservice")
		return err
	}

	return nil
}

// TODO We do need to bubble up the microservices
func (s *GitStorage) GetBusinessMoments(customerID string, applicationID string, environment string) (platform.HttpResponseBusinessMoments, error) {
	logContext := s.logContext.WithFields(logrus.Fields{
		"customer_id":    customerID,
		"application_id": applicationID,
		"environment":    environment,
	})

	microservices, err := s.GetMicroservices(customerID, applicationID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("Getting microservices")
	}

	data := platform.HttpResponseBusinessMoments{
		ApplicationID: applicationID,
		Environment:   environment,
		Moments:       make([]platform.HttpInputBusinessMoment, 0),
		Entities:      make([]platform.HttpInputBusinessMomentEntity, 0),
	}

	for _, microservice := range microservices {
		// Filter
		if strings.ToLower(microservice.Environment) != environment {
			continue
		}

		if microservice.Kind != platform.MicroserviceKindBusinessMomentsAdaptor {
			continue
		}

		b, _ := json.Marshal(microservice)
		var businessMomentsAdaptor platform.HttpInputBusinessMomentAdaptorInfo
		err = json.Unmarshal(b, &businessMomentsAdaptor)

		// Add Business moments
		for _, moment := range businessMomentsAdaptor.Extra.Moments {
			withInfo := platform.HttpInputBusinessMoment{
				ApplicationID:  applicationID,
				Environment:    environment,
				MicroserviceID: businessMomentsAdaptor.Dolittle.MicroserviceID,
				Moment:         moment,
			}
			data.Moments = append(data.Moments, withInfo)
		}

		// Add entities
		for _, entity := range businessMomentsAdaptor.Extra.Entities {
			withInfo := platform.HttpInputBusinessMomentEntity{
				ApplicationID:  applicationID,
				Environment:    environment,
				MicroserviceID: businessMomentsAdaptor.Dolittle.MicroserviceID,
				Entity:         entity,
			}
			data.Entities = append(data.Entities, withInfo)
		}
	}

	return data, nil
}

func (s *GitStorage) DeleteBusinessMoment(customerID string, applicationID string, environment string, microserviceID string, momentID string) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"customer_id":     customerID,
		"application_id":  applicationID,
		"environment":     environment,
		"microservice_id": microserviceID,
		"moment_id":       momentID,
	})

	rawBytes, err := s.GetMicroservice(customerID, applicationID, environment, microserviceID)
	if err != nil {
		return storage.ErrNotFound
	}

	var microservice platform.HttpInputBusinessMomentAdaptorInfo
	err = json.Unmarshal(rawBytes, &microservice)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("unmarshall issues")
		return err
	}

	// Confirm its business moment
	if microservice.Kind != platform.MicroserviceKindBusinessMomentsAdaptor {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("not-business-moments-adaptor")
		return storage.ErrNotBusinessMomentsAdaptor
	}

	index := funk.IndexOf(microservice.Extra.Moments, func(moment platform.BusinessMoment) bool {
		return moment.UUID == momentID
	})

	if index == -1 {
		return nil
	}

	microservice.Extra.Moments = append(microservice.Extra.Moments[:index], microservice.Extra.Moments[index+1:]...)
	err = s.SaveMicroservice(customerID, applicationID, environment, microservice.Dolittle.MicroserviceID, microservice)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.SaveMicroservice",
		}).Error("On saving microservice")
		return err
	}

	return nil
}

// TODO this is not good enough
func (s *GitStorage) DeleteBusinessMomentEntity(customerID string, applicationID string, environment string, microserviceID string, entityID string) error {
	logContext := s.logContext.WithFields(logrus.Fields{
		"customer_id":     customerID,
		"application_id":  applicationID,
		"environment":     environment,
		"microservice_id": microserviceID,
		"entity_id":       entityID,
	})

	rawBytes, err := s.GetMicroservice(customerID, applicationID, environment, microserviceID)
	if err != nil {
		return storage.ErrNotFound
	}

	var microservice platform.HttpInputBusinessMomentAdaptorInfo
	err = json.Unmarshal(rawBytes, &microservice)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("unmarshall issues")
		return err
	}

	// Confirm its business moment
	if microservice.Kind != platform.MicroserviceKindBusinessMomentsAdaptor {
		logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("not-business-moments-adaptor")
		return storage.ErrNotBusinessMomentsAdaptor
	}

	index := funk.IndexOf(microservice.Extra.Entities, func(entity platform.Entity) bool {
		return entity.EntityTypeID == entityID
	})

	if index == -1 {
		return nil
	}
	// Remove business moments
	cleanedMoments := funk.Filter(microservice.Extra.Moments, func(moment platform.BusinessMoment) bool {
		return moment.EntityTypeID != entityID
	}).([]platform.BusinessMoment)
	microservice.Extra.Moments = cleanedMoments

	// Remove from entity
	microservice.Extra.Entities = append(microservice.Extra.Entities[:index], microservice.Extra.Entities[index+1:]...)

	err = s.SaveMicroservice(customerID, applicationID, environment, microservice.Dolittle.MicroserviceID, microservice)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"error":  err,
			"method": "s.SaveMicroservice",
		}).Error("On saving microservice")
		return err
	}

	return nil
}
