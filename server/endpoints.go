package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mleone896/inventory/db"
	"github.com/mleone896/inventory/models"
)

// TagsRequest ...
type TagsRequest struct {
	Role        string `json:"primary_role"`
	Environment string `json:"environment"`
	SubnetID    string `json:"subnet_id"`
	Pool        string `json:"pool,omitempty"`
	Color       string `json:"color,omitempty"`
	Name        string `json:"name"`
	Owner       string `json:"owner,omitempty"`
	err         error
}

// IsValid ...
func (h *TagsRequest) IsValid() bool {
	if h.Role == "" {
		h.err = fmt.Errorf("primary_role must be valid")
		return false
	}

	if h.Environment == "" {
		h.err = fmt.Errorf("environment must be valid")
		return false
	}

	if h.SubnetID == "" {
		h.err = fmt.Errorf("subnet_id must be valid")
		return false
	}

	if h.Pool == "" {
		h.err = fmt.Errorf("pool must be valid")
		return false
	}

	return true
}

// APIError struct represents a json return erorr type
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

// Error ...
func Error(w http.ResponseWriter, status int, reason, detail string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	errorJSON, _ := json.Marshal(APIError{
		Code:    status,
		Message: reason,
		Detail:  detail})

	_, err := w.Write(errorJSON)

	if err != nil {
		log.Printf("could not write api error  %s", reason)
	}

}

// ListHostAttrsByColor ...
func (ctx *APIContext) ListHostAttrsByColor(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id := vars["id"]

	if id == "" {
		Error(w, http.StatusBadRequest, "bad color sent in request %s", "key not found")
		return
	}

	instance := models.NewInstance(models.WithColorTag(id))

	if err := instance.GetByID(ctx.dao.Conn); err != nil {
		Error(w, http.StatusBadRequest, "could not retrieve instance info: %s", err.Error())
		return
	}

	res := convertInstanceToTagsReq(instance)

	if err := json.NewEncoder(w).Encode(res); err != nil {
		Error(w, http.StatusInternalServerError,
			"instanceByColor: could not write json response to http handler",
			err.Error())
		return
	}

	return

}

// NewTagsReq ...
func (ctx *APIContext) NewTagsReq(w http.ResponseWriter, r *http.Request) {
	var hreq *TagsRequest
	// decode the req
	if err := json.NewDecoder(r.Body).Decode(&hreq); err != nil {
		Error(w, http.StatusBadRequest, "could not read body, please send valid req", err.Error())
		return
	}

	// make sure all fields have values
	if !hreq.IsValid() {
		Error(w, http.StatusBadRequest, "could not read body, please send valid req", hreq.err.Error())
		return
	}

	response, err := generateNewHostTags(hreq, ctx.dao)

	if err != nil {
		Error(w, http.StatusInternalServerError, "could not generate correct host tags", hreq.err.Error())
		return

	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		Error(w, http.StatusInternalServerError, "could not write json response to http handler", hreq.err.Error())
		return
	}

	return

}

// ListColors returns all the available colors not in use
func (ctx *APIContext) ListColors(w http.ResponseWriter, r *http.Request) {

	obj := models.Colors()
	rows, err := ctx.dao.FindAll(obj)

	if err != nil {
		Error(w, http.StatusInternalServerError, "could not find colors", err.Error())
		return
	}

	colors, err := obj.UnpackRows(rows)

	if err != nil {
		Error(w, http.StatusInternalServerError, "could not unpack colors", err.Error())
		return
	}

	if err := json.NewEncoder(w).Encode(colors); err != nil {
		Error(w, http.StatusInternalServerError, "failed to marshal", err.Error())
		return

	}

}

func generateNewHostTags(treq *TagsRequest, db *db.DataObj) (*TagsRequest, error) {

	res := new(TagsRequest)
	// get the subnet from the id sent in payload

	subnet, err := models.NewSubnet(models.WithSubnetID(treq.SubnetID))

	if err != nil {
		return nil, err
	}

	// get the subnet from the db and populate the object
	if err := db.Read(subnet); err != nil {
		return nil, err
	}

	color, err := models.NewColor(models.WithDefault())

	// get a random unused color from db and populate color object
	if err := db.Read(color); err != nil {
		return nil, err
	}

	// this is no bueno
	nameTag := formatHostName(
		treq.Environment,
		treq.Role,
		treq.Pool,
		color.Name,
		factorAvailabiltyZone(subnet.AZ))

	res.Name = nameTag
	res.Color = color.Name
	res.Owner = "TBD"
	res.Role = treq.Role
	res.SubnetID = treq.SubnetID
	res.Pool = treq.Pool
	res.Environment = treq.Environment

	return res, nil
}

// gets the identifier out of an aws az
func factorAvailabiltyZone(az string) string {
	// NOTE(mlcrsi): this is prone to error if AvailabilityZone does not meet
	// the standard AWS $LETTERS-IDENTIFIER-AZTHING
	parts := strings.SplitN(az, "-", 3)
	azIdentifier := parts[2]

	return azIdentifier
}

// essentially takes a req + subnet identifier and returns the Name tag
func formatHostName(env, role, pool, color, azIdentifier string) string {
	// Determine the hostname
	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		string(env[0]),
		role,
		string(pool[0]),
		color,
		azIdentifier,
	)
}

func convertInstanceToTagsReq(i *models.Instance) *TagsRequest {
	tr := &TagsRequest{}
	tr.Role = i.Tags.Map["role"].String
	tr.Color = i.Tags.Map["color"].String
	tr.SubnetID = i.SubnetID
	tr.Environment = i.Tags.Map["environment"].String
	tr.Name = i.Tags.Map["Name"].String

	return tr
}
