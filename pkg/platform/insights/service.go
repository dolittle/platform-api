package insights

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/dolittle/platform-api/pkg/platform"
	"github.com/dolittle/platform-api/pkg/platform/mongo"
	"github.com/dolittle/platform-api/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func NewService(logContext logrus.FieldLogger, k8sDolittleRepo platform.K8sRepo, lokiHost string) service {
	return service{
		logContext:      logContext,
		k8sDolittleRepo: k8sDolittleRepo,
		lokiHost:        lokiHost,
	}
}

func (s *service) GetRuntimeV1(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	applicationID := vars["applicationID"]
	environment := strings.ToLower(vars["environment"])

	userID := r.Header.Get("User-ID")
	tenantID := r.Header.Get("Tenant-ID")
	allowed := s.k8sDolittleRepo.CanModifyApplicationWithResponse(w, tenantID, applicationID, userID)
	if !allowed {
		return
	}

	logContext := s.logContext.WithFields(logrus.Fields{
		"applicationID": applicationID,
		"tenantID":      tenantID,
		"userID":        userID,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	mongoURI := mongo.GetMongoURI(applicationID, environment)

	client, err := mongo.SetupMongo(ctx, mongoURI)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"mongoURI": mongoURI,
			"error":    err,
		}).Error("Connecting to mongo")
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	dbs := mongo.GetEventStoreDatabases(ctx, client)

	latestEvents := make(map[string]platform.RuntimeLatestEvent, 0)
	latestEventsPerEventType := make(map[string][]platform.RuntimeLatestEvent, 0)
	eventLogCounts := make(map[string]int64, 0)
	runtimeStates := make(map[string][]platform.RuntimeState, 0)

	for _, db := range dbs {
		key := fmt.Sprintf("%s", db)

		_latestEvents, err := mongo.GetLatestEvent(ctx, client, db)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"database": db,
				"error":    err,
				"method":   "mongo.GetLatestEvent",
			}).Info("Skipping db: Failed to get latest event, skipping db")
			continue
		}

		latestEvents[key] = _latestEvents

		_latestEventsPerEventType, err := mongo.GetLatestEventPerEventType(ctx, client, db)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"database": db,
				"error":    err,
				"method":   "mongo.GetLatestEventPerEventType",
			}).Info("Skipping db: Failed to get latest event per event type")
			continue
		}
		latestEventsPerEventType[key] = _latestEventsPerEventType

		eventLogCount, err := mongo.GetEventLogCount(ctx, client, db)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"database": db,
				"error":    err,
				"method":   "mongo.GetEventLogCount",
			}).Info("Skipping db: Failed to get event log count")
			continue
		}
		eventLogCounts[key] = eventLogCount

		_runtimeStates, err := mongo.GetRuntimeStates(ctx, client, db)
		if err != nil {
			logContext.WithFields(logrus.Fields{
				"database": db,
				"error":    err,
				"method":   "mongo.GetRuntimeStates",
			}).Info("Skipping db: Failed to get runtime states")
			continue
		}
		runtimeStates[key] = _runtimeStates
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"applicationID":            applicationID,
		"environment":              environment,
		"latestEvents":             latestEvents,
		"latestEventsPerEventType": latestEventsPerEventType,
		"eventLogCounts":           eventLogCounts,
		"runtimeStates":            runtimeStates,
	})
	return
}

// ProxyLoki
func (s *service) ProxyLoki(w http.ResponseWriter, r *http.Request) {
	tenantID := r.Header.Get("Tenant-ID")
	r.Header.Del("Tenant-ID")
	r.Header.Del("User-ID")
	r.Header.Del("x-shared-secret")
	r.Header.Set("X-Scope-OrgId", fmt.Sprintf("tenant-%s", tenantID))
	// Remove prefix
	parts := strings.Split(r.URL.Path, "/loki")

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	r = r.WithContext(ctx)

	r.URL.Path = strings.TrimPrefix(r.URL.Path, parts[0])
	serveReverseProxy(s.lokiHost, w, r)
}

func serveReverseProxy(host string, res http.ResponseWriter, req *http.Request) {
	url, _ := url.Parse("/")
	url.Host = host
	// Hard coding to http for now
	url.Scheme = "http"
	proxy := httputil.NewSingleHostReverseProxy(url)

	// Update the request
	req.Host = url.Host
	req.URL.Host = url.Host
	req.URL.Scheme = url.Scheme

	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	// Note that ServeHttp is non blocking and uses a go routine under the hood
	proxy.ServeHTTP(res, req)
}
