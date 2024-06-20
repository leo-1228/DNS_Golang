package api

import (
	"dnscheck/dbs"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
)

type ApiV1Server struct {
	Url    string
	Secret string
	router http.Handler
}

func sendResponse(wr http.ResponseWriter, status int, message string) {
	wr.WriteHeader(status)
	wr.Write([]byte(message))
}

func NewApiV1Server(url, secret string) *ApiV1Server {

	r := chi.NewRouter()

	s := &ApiV1Server{
		Url:    url,
		Secret: secret,
		router: r,
	}

	r.Get(withPrefix("/healthcheck"), s.healthCheck)
	r.Get(withPrefix("/config/{clientId}"), s.getConfig)
	r.Post(withPrefix("/results"), s.setResults)

	return s
}

func (s *ApiV1Server) Serve() error {
	return http.ListenAndServe(s.Url, s.router)
}

func (s *ApiV1Server) handleAuth(req *http.Request) error {
	_, err := authJwt(req.Header.Get("authorization"), []byte(s.Secret))
	return err
}

func (s *ApiV1Server) getConfig(wr http.ResponseWriter, req *http.Request) {
	if err := s.handleAuth(req); err != nil {
		sendResponse(wr, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get config from here
	clientId := chi.URLParam(req, "clientId")
	config := dbs.ClientDB.Get(clientId)
	if err := json.NewEncoder(wr).Encode(config); err != nil {
		sendResponse(wr, http.StatusInternalServerError, "Failed to encode response")
		return
	}
}

func (s *ApiV1Server) setResults(wr http.ResponseWriter, req *http.Request) {
	if err := s.handleAuth(req); err != nil {
		sendResponse(wr, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var setResultsRequest dbs.ResultInfo
	if err := json.NewDecoder(req.Body).Decode(&setResultsRequest); err != nil {
		sendResponse(wr, http.StatusBadRequest, "Could not parse request")
		return
	}

	// logrus.Warn(setResultsRequest.ClientId)

	// Update results
	dbs.ClientDB.AddResult(setResultsRequest)

	wr.WriteHeader(http.StatusOK)
	wr.Write([]byte("OK"))
}

func (s *ApiV1Server) healthCheck(wr http.ResponseWriter, req *http.Request) {
	wr.Header().Add("Cache-Control", "no-store")
	wr.WriteHeader(http.StatusOK)
	wr.Write([]byte("Healthy"))
}
