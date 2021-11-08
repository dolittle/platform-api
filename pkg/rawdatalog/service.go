package rawdatalog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

type service struct {
	logContext               logrus.FieldLogger
	repo                     Repo
	uriPrefix                string
	pathToMicroserviceConfig string
	topic                    string
	tenantID                 string
	applicationID            string
	environment              string
	allowedUriSuffixes       map[string]platform.RawDataLogIngestorWebhookConfig
}

func NewService(logContext logrus.FieldLogger, uriPrefix string, pathToMicroserviceConfig string, topic string, repo Repo, tenantID string, applicationID string, environment string) *service {
	s := &service{
		logContext:               logContext,
		uriPrefix:                uriPrefix,
		pathToMicroserviceConfig: pathToMicroserviceConfig,
		topic:                    topic,
		repo:                     repo,
		tenantID:                 tenantID,
		applicationID:            applicationID,
		environment:              environment,
	}

	s.loadAllowedUriSuffixes()
	go s.watchAndLoadAllowedUriSuffixes()
	return s
}

func (s *service) loadAllowedUriSuffixes() {
	// TODO add watch to the file
	b, err := ioutil.ReadFile(s.pathToMicroserviceConfig)

	var data platform.HttpInputRawDataLogIngestorInfo
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":                    err,
			"pathToMicroserviceConfig": s.pathToMicroserviceConfig,
		}).Fatal("loading microservice config")
	}

	err = json.Unmarshal(b, &data)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":                    err,
			"pathToMicroserviceConfig": s.pathToMicroserviceConfig,
		}).Fatal("loading microservice config")
	}

	allowedUriSuffixes := map[string]platform.RawDataLogIngestorWebhookConfig{}
	for _, webhook := range data.Extra.Webhooks {
		allowedUriSuffixes[webhook.UriSuffix] = webhook
	}
	s.allowedUriSuffixes = allowedUriSuffixes
	s.logContext.WithFields(logrus.Fields{
		"webhooks": s.allowedUriSuffixes,
	}).Info("allowedUriSuffix updated")
}

func (s *service) watchAndLoadAllowedUriSuffixes() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Add(s.pathToMicroserviceConfig)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("error whilst starting to watch file change")
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				s.logContext.WithFields(logrus.Fields{
					"ok":    ok,
					"event": event,
				}).Info("Event")
				return
			}

			s.logContext.WithFields(logrus.Fields{
				"event": event,
			}).Info("Event")

			// https://martensson.io/go-fsnotify-and-kubernetes-configmaps/
			if event.Op == fsnotify.Remove {
				watcher.Remove(event.Name)
				watcher.Add(s.pathToMicroserviceConfig)
				s.loadAllowedUriSuffixes()
				continue
			}

			if event.Op&fsnotify.Write == fsnotify.Write {
				s.logContext.Info("file written")
				s.loadAllowedUriSuffixes()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			s.logContext.WithFields(logrus.Fields{
				"error": err,
			}).Fatal("error whilst watching file change")
		}
	}
}

func (s *service) Webhook(w http.ResponseWriter, r *http.Request) {
	topic := s.topic
	tenantID := s.tenantID
	applicationID := s.applicationID
	environment := s.environment

	pathname := r.URL.Path
	pathname = strings.ToLower(pathname)
	pathname = strings.TrimPrefix(pathname, s.uriPrefix)
	pathname = strings.TrimPrefix(pathname, "/")
	pathname = strings.TrimSuffix(pathname, "/")

	// TODO read from config-files
	webhook, ok := s.allowedUriSuffixes[pathname]
	if !ok {
		s.logContext.WithFields(logrus.Fields{
			"error":            fmt.Sprintf("uriSuffix not on the list: %s", pathname),
			"webhookUriSuffix": pathname,
			"webhookUriPrefix": s.uriPrefix,
			"context":          "verify webhook configured",
		}).Error("webhook")
		utils.RespondWithError(w, http.StatusForbidden, "Webhook not supported due to unknown uri suffix")
		return
	}

	if r.Header.Get("Authorization") != webhook.Authorization {
		s.logContext.WithFields(logrus.Fields{
			"error":                "authorization failed",
			"headerAuthorization":  r.Header.Get("Authorization"),
			"webhookAuthorization": webhook.Authorization,
			"webhookUriSuffix":     webhook.UriSuffix,
			"context":              "checking authorization",
		}).Error("webhook")
		utils.RespondWithError(w, http.StatusForbidden, "Webhook not supported, failed authorization")
		return
	}

	parts := strings.Split(pathname, "/")

	labels := map[string]string{
		"uriSuffix": pathname,
	}

	for index, part := range parts {
		indexStr := fmt.Sprintf("uri-%d", index)
		labels[indexStr] = part
	}

	kind := webhook.Kind
	metadata := RawMomentMetadata{
		TenantID:      tenantID,
		ApplicationID: applicationID,
		Environment:   environment,
		Labels:        map[string]string{},
	}
	metadata.Labels = labels

	var dst interface{}
	dec := json.NewDecoder(r.Body)

	err := dec.Decode(&dst)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":   err,
			"context": "incoming payload",
		}).Error("webhook")
		utils.RespondWithError(w, http.StatusBadRequest, "Failed to pass payload")
		return
	}

	moment := RawMoment{
		Kind:     kind,
		When:     time.Now().UTC().Unix(),
		Data:     dst,
		Metadata: metadata,
	}

	err = s.repo.Write(topic, moment)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"error":   err,
			"context": "writing to log",
		}).Error("webhook")
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to write to log")
		return
	}

	utils.RespondNoContent(w, http.StatusOK)
	return
}
