package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type APIServer struct {
	router *mux.Router
	db *gorm.DB
}

func (a *APIServer) Start() error {
	a.router = mux.NewRouter()
	var err error
	if a.db == nil {
		a.db, err = ConnectToDB()
		if err != nil {
			return fmt.Errorf("unable to connect to database: %w", err)
		}
	}
	a.registerHandlers()
	return nil
}

func (a *APIServer) Serve(port int) {
	go http.ListenAndServe(fmt.Sprintf(":%d", port), a.router)
}

func (a *APIServer) registerHandlers() {
	a.registerAirspaceHandlers()
	a.registerFlightHandlers()
	a.registerFormationHandlers()
	//a.registerPlaneHandlers()
}

func (a *APIServer) respondWithError(w http.ResponseWriter) {
	a.respondWithErrorMessage(w, "Internal server error")
}

func (a *APIServer) respondWithErrorMessage(w http.ResponseWriter, message string) {
	a.respondWithJSON(w, http.StatusInternalServerError, map[string]string{"error": message})
}

func (a *APIServer) respondOKWithJson(w http.ResponseWriter, payload interface{}) {
	a.respondWithJSON(w, http.StatusOK, payload)
}

func (a *APIServer) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Airspace REST handlers, functions and helpers                                                                     //
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type RESTAirspace struct {
	ID uint
	HumanName string
	NetName string
}

func (a *APIServer) registerAirspaceHandlers() {
	a.router.HandleFunc("/airspaces", a.getAirspaces).Methods("GET")
	a.router.HandleFunc("/airspace", a.createAirspace).Methods("POST")
	a.router.HandleFunc("/airspace/{id:[0-9]+}", a.getAirspace).Methods("GET")
	a.router.HandleFunc("/airspace/{id:[0-9]+}", a.updateAirspace).Methods("PUT")
	a.router.HandleFunc("/airspace/{id:[0-9]+}", a.deleteAirspace).Methods("DELETE")
}

func (a *APIServer) getAirspaces(w http.ResponseWriter, r *http.Request) {
	var airspaces []Airspace
	result := a.db.Find(&airspaces)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var restAirspaces []RESTAirspace
	for i, _ := range airspaces {
		restAirspaces = append(restAirspaces, RESTAirspace{
			ID: airspaces[i].ID,
			HumanName: airspaces[i].HumanName,
			NetName: airspaces[i].NetName,
		})
	}
	a.respondOKWithJson(w, restAirspaces)
}

func (a *APIServer) createAirspace(w http.ResponseWriter, r *http.Request) {
	var as RESTAirspace
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&as); err != nil {
		a.respondWithErrorMessage(w, "Invalid Airspace object")
		return
	}
	defer r.Body.Close()
	airspace := Airspace{
		HumanName: as.HumanName,
		NetName: as.NetName,
	}
	result := a.db.Create(&airspace)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	as.ID = airspace.ID
	a.respondWithJSON(w, http.StatusCreated, as)
}

func (a *APIServer) getAirspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var airspace Airspace
	result := a.db.First(&airspace, id)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	a.respondOKWithJson(w, RESTAirspace{
			ID: airspace.ID,
			HumanName: airspace.HumanName,
			NetName: airspace.NetName,
	})
}

func (a *APIServer) updateAirspace(w http.ResponseWriter, r *http.Request) {
	var as RESTAirspace
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&as); err != nil {
		a.respondWithErrorMessage(w, "Invalid Airspace object")
		return
	}
	defer r.Body.Close()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var airspace Airspace
	result := a.db.First(&airspace, id)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	airspace.HumanName = as.HumanName
	airspace.NetName = as.NetName
	result = a.db.Save(&airspace)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	a.respondWithJSON(w, http.StatusCreated, as)
}

func (a *APIServer) deleteAirspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	result := a.db.Delete(&Airspace{}, id)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Flight REST handlers, functions and helpers                                                                       //
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type RESTFlight struct {
	AirspaceID uint
	ID uint
	Name string
}

func (a *APIServer) registerFlightHandlers() {
	a.router.HandleFunc("/flights", a.getFlights).Methods("GET")
	a.router.HandleFunc("/airspace/{aid:[0-9+]}/flights", a.getFlightsInAirspace).Methods("GET")
	a.router.HandleFunc("/flight", a.createFlight).Methods("POST")
	a.router.HandleFunc("/flight/{id:[0-9]+}", a.getFlight).Methods("GET")
	a.router.HandleFunc("/flight/{id:[0-9]+}", a.updateFlight).Methods("PUT")
	a.router.HandleFunc("/flight/{id:[0-9]+}", a.deleteFlight).Methods("DELETE")
}

func (a *APIServer) getFlights(w http.ResponseWriter, r *http.Request) {
	var flights []Flight
	result := a.db.Find(&flights)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var restFlights []RESTFlight
	for i, _ := range flights {
		restFlights = append(restFlights, RESTFlight{
			ID: flights[i].ID,
			Name: flights[i].Name,
		})
	}
	a.respondOKWithJson(w, restFlights)
}

func (a *APIServer) getFlightsInAirspace(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aid, err := strconv.Atoi(vars["aid"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}

	var flights []Flight
	result := a.db.Where("airspace_id = ?", aid).Find(&flights)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var restFlights []RESTFlight
	for i, _ := range flights {
		restFlights = append(restFlights, RESTFlight{
			ID: flights[i].ID,
			Name: flights[i].Name,
		})
	}
	a.respondOKWithJson(w, restFlights)
}

func (a *APIServer) createFlight(w http.ResponseWriter, r *http.Request) {
	var as RESTFlight
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&as); err != nil {
		a.respondWithErrorMessage(w, "Invalid Flight object")
		return
	}
	defer r.Body.Close()
	// TODO: Verify that AirspaceID exists.
	flight := Flight{
		AirspaceID: int(as.AirspaceID),
		Name: as.Name,
	}
	result := a.db.Create(&flight)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	as.ID = flight.ID
	a.respondWithJSON(w, http.StatusCreated, as)
}

func (a *APIServer) getFlight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var flight Flight
	result := a.db.First(&flight, id)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	a.respondOKWithJson(w, RESTFlight{
		AirspaceID: uint(flight.AirspaceID),
		ID:         flight.ID,
		Name:       flight.Name,
	})
}

func (a *APIServer) updateFlight(w http.ResponseWriter, r *http.Request) {
	var as RESTFlight
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&as); err != nil {
		a.respondWithErrorMessage(w, "Invalid Flight object")
		return
	}
	defer r.Body.Close()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var flight Flight
	result := a.db.First(&flight, id)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	flight.Name = as.Name
	result = a.db.Save(&flight)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	a.respondWithJSON(w, http.StatusCreated, as)
}

func (a *APIServer) deleteFlight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	result := a.db.Delete(&Flight{}, id)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Formation REST handlers, functions and helpers                                                                    //
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

type RESTFormation struct {
	FlightID int
	ID uint
	Name string
	CPU int
	RAM int
	Disk int
	BaseName string
	Domain string
	TargetCount int
}

func (a *APIServer) registerFormationHandlers() {
	a.router.HandleFunc("/formations", a.getFormations).Methods("GET")
	a.router.HandleFunc("/flight/{fid:[0-9+]}/formations", a.getFormationsInFlight).Methods("GET")
	a.router.HandleFunc("/formation", a.createFormation).Methods("POST")
	a.router.HandleFunc("/formation/{id:[0-9]+}", a.getFormation).Methods("GET")
	a.router.HandleFunc("/formation/{id:[0-9]+}", a.updateFormation).Methods("PUT")
	a.router.HandleFunc("/formation/{id:[0-9]+}", a.deleteFormation).Methods("DELETE")
}

func (a *APIServer) getFormations(w http.ResponseWriter, r *http.Request) {
	var formations []Formation
	result := a.db.Find(&formations)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var restFormations []RESTFormation
	for i, _ := range formations {
		restFormations = append(restFormations, RESTFormation{
			ID: formations[i].ID,
			Name: formations[i].Name,
			CPU: formations[i].CPU,
			RAM: formations[i].RAM,
			Disk: formations[i].Disk,
			BaseName: formations[i].BaseName,
			Domain: formations[i].Domain,
			TargetCount: formations[i].TargetCount,
			FlightID: formations[i].FlightID,
		})
	}
	a.respondOKWithJson(w, restFormations)
}

func (a *APIServer) getFormationsInFlight(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	aid, err := strconv.Atoi(vars["fid"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}

	var formations []Formation
	result := a.db.Where("flight_id = ?", aid).Find(&formations)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var restFormations []RESTFormation
	for i, _ := range formations {
		restFormations = append(restFormations, RESTFormation{
			ID: formations[i].ID,
			Name: formations[i].Name,
			CPU: formations[i].CPU,
			RAM: formations[i].RAM,
			Disk: formations[i].Disk,
			BaseName: formations[i].BaseName,
			Domain: formations[i].Domain,
			TargetCount: formations[i].TargetCount,
			FlightID: formations[i].FlightID,
		})
	}
	a.respondOKWithJson(w, restFormations)
}

func (a *APIServer) createFormation(w http.ResponseWriter, r *http.Request) {
	var as RESTFormation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&as); err != nil {
		a.respondWithErrorMessage(w, "Invalid Formation object")
		return
	}
	defer r.Body.Close()
	// TODO: Verify that FlightID exists.
	formation := Formation{
		FlightID: as.FlightID,
		Name: as.Name,
		CPU: as.CPU,
		RAM: as.RAM,
		Disk: as.Disk,
		BaseName: as.BaseName,
		Domain: as.Domain,
		TargetCount: as.TargetCount,
	}
	result := a.db.Create(&formation)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	as.ID = formation.ID
	a.respondWithJSON(w, http.StatusCreated, as)
}

func (a *APIServer) getFormation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var as Formation
	result := a.db.First(&as, id)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	a.respondOKWithJson(w, RESTFormation{
		FlightID: as.FlightID,
		Name: as.Name,
		CPU: as.CPU,
		RAM: as.RAM,
		Disk: as.Disk,
		BaseName: as.BaseName,
		Domain: as.Domain,
		TargetCount: as.TargetCount,
	})
}

func (a *APIServer) updateFormation(w http.ResponseWriter, r *http.Request) {
	var as RESTFormation
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&as); err != nil {
		a.respondWithErrorMessage(w, "Invalid Formation object")
		return
	}
	defer r.Body.Close()
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	var formation Formation
	result := a.db.First(&formation, id)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	// Only the TargetCount can be changed. When rolling releases are implemented in the future some values like CPU and RAM will also be able to be changed.
	formation.TargetCount = as.TargetCount
	result = a.db.Save(&formation)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	a.respondWithJSON(w, http.StatusCreated, as)
}

func (a *APIServer) deleteFormation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		log.Error().Stack().Err(err).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
	result := a.db.Delete(&Formation{}, id)
	if result.Error != nil {
		log.Error().Stack().Err(result.Error).Msgf("unable to process request %s", r.RequestURI)
		a.respondWithError(w)
		return
	}
}