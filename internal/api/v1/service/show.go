package service

import (
	"net/http"

	"github.com/epinio/epinio/helpers/kubernetes"
	"github.com/epinio/epinio/internal/api/v1/response"
	"github.com/epinio/epinio/internal/organizations"
	"github.com/epinio/epinio/internal/services"
	"github.com/epinio/epinio/pkg/api/core/v1/models"
	"github.com/julienschmidt/httprouter"

	. "github.com/epinio/epinio/pkg/api/core/v1/errors"
)

// Show handles the API end point /orgs/:org/services/:service
// It returns the detail information of the named service instance
func (sc Controller) Show(w http.ResponseWriter, r *http.Request) APIErrors {
	ctx := r.Context()
	params := httprouter.ParamsFromContext(ctx)
	org := params.ByName("org")
	serviceName := params.ByName("service")

	cluster, err := kubernetes.GetCluster(ctx)
	if err != nil {
		return InternalError(err)
	}

	exists, err := organizations.Exists(ctx, cluster, org)
	if err != nil {
		return InternalError(err)
	}
	if !exists {
		return OrgIsNotKnown(org)
	}

	service, err := services.Lookup(ctx, cluster, org, serviceName)
	if err != nil {
		if err.Error() == "service not found" {
			return ServiceIsNotKnown(serviceName)
		}
		if err != nil {
			return InternalError(err)
		}
	}

	serviceDetails, err := service.Details(ctx)
	if err != nil {
		return InternalError(err)
	}

	responseData := map[string]string{}
	for key, value := range serviceDetails {
		responseData[key] = value
	}

	err = response.JSON(w, models.ServiceShowResponse{
		Username: service.User(),
		Details:  responseData,
	})
	if err != nil {
		return InternalError(err)
	}

	return nil
}
