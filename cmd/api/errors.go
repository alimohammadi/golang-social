package main

import (
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	// use r for log
	app.logger.Errorw("internal server error", "method", r.Method, "path", r.URL.Path, "error", err)
	writeJSONError(w, http.StatusInternalServerError, "The server encountered a problem")
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	// use r for log
	app.logger.Warnf("bad request", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request, err error) {
	// use r for log
	app.logger.Errorw("conflict response", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusNotFound, "Not found")
}

func (app *application) conflictResponse(w http.ResponseWriter, r *http.Request, err error) {
	// use r for log
	// log.Printf("conflict error: %s path: %s error: %s", r.Method, r.URL.Path, err.Error())
	app.logger.Warnf("not found error", "method", r.Method, "path", r.URL.Path, "error", err)

	writeJSONError(w, http.StatusConflict, err.Error())
}
