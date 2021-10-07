package application

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/epinio/epinio/helpers/kubernetes"
	"github.com/epinio/epinio/internal/api/v1/response"
	"github.com/epinio/epinio/internal/application"
	"github.com/epinio/epinio/internal/cli/server/requestctx"
	"github.com/epinio/epinio/internal/organizations"
	"github.com/epinio/epinio/internal/services"
	"github.com/epinio/epinio/pkg/api/core/v1/models"

	"github.com/julienschmidt/httprouter"

	. "github.com/epinio/epinio/pkg/api/core/v1/errors"
)

// Create handles the API endpoint POST /namespaces/:org/applications
// It creates a new and empty application. I.e. without a workload.
func (hc Controller) Create(w http.ResponseWriter, r *http.Request) APIErrors {
	ctx := r.Context()
	params := httprouter.ParamsFromContext(ctx)
	org := params.ByName("org")
	username := requestctx.User(ctx)

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

	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return InternalError(err)
	}

	var createRequest models.ApplicationCreateRequest
	err = json.Unmarshal(bodyBytes, &createRequest)
	if err != nil {
		return BadRequest(err)
	}

	appRef := models.NewAppRef(createRequest.Name, org)
	found, err := application.Exists(ctx, cluster, appRef)
	if err != nil {
		return InternalError(err, "failed to check for app resource")
	}
	if found {
		return AppAlreadyKnown(createRequest.Name)
	}

	// Sanity check the services, if any. IOW anything to be bound
	// has to exist now.  We will check again when the application
	// is deployed, to guard against bound services being removed
	// from now till then. While it should not be possible through
	// epinio itself (*), external editing of the relevant
	// resources cannot be excluded from consideration.
	//
	// (*) `epinio service delete S` on a bound service S will
	//      either reject the operation, or, when forced, unbind S
	//      from the app.

	var theIssues []APIError

	for _, serviceName := range createRequest.Configuration.Services {
		_, err := services.Lookup(ctx, cluster, org, serviceName)
		if err != nil {
			if err.Error() == "service not found" {
				theIssues = append(theIssues, ServiceIsNotKnown(serviceName))
				continue
			}

			theIssues = append([]APIError{InternalError(err)}, theIssues...)
			return NewMultiError(theIssues)
		}
	}

	if len(theIssues) > 0 {
		return NewMultiError(theIssues)
	}

	// Arguments found OK, now we can modify the system state

	err = application.Create(ctx, cluster, appRef, username)
	if err != nil {
		return InternalError(err)
	}

	desired := DefaultInstances
	if createRequest.Configuration.Instances != nil {
		desired = *createRequest.Configuration.Instances
	}

	err = application.ScalingSet(ctx, cluster, appRef, desired)
	if err != nil {
		return InternalError(err)
	}

	// Save service information.
	err = application.BoundServicesSet(ctx, cluster, appRef,
		createRequest.Configuration.Services, true)
	if err != nil {
		return InternalError(err)
	}

	// Save environment assignments
	err = application.EnvironmentSet(ctx, cluster, appRef,
		createRequest.Configuration.Environment, true)
	if err != nil {
		return InternalError(err)
	}

	err = response.JSON(w, models.ResponseOK)
	if err != nil {
		return InternalError(err)
	}
	return nil
}
