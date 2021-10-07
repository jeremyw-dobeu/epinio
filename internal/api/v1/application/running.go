package application

import (
	"net/http"

	"github.com/epinio/epinio/helpers/kubernetes"
	"github.com/epinio/epinio/internal/api/v1/response"
	"github.com/epinio/epinio/internal/application"
	"github.com/epinio/epinio/internal/duration"
	"github.com/epinio/epinio/internal/organizations"
	"github.com/epinio/epinio/pkg/api/core/v1/models"

	"github.com/julienschmidt/httprouter"

	. "github.com/epinio/epinio/pkg/api/core/v1/errors"
)

// Running handles the API endpoint GET /namespaces/:org/applications/:app/running
// It waits for the specified application to be running (i.e. its
// deployment to be complete), before it returns. An exception is if
// the application does not become running without
// `duration.ToAppBuilt()` (default: 10 minutes). In that case it
// returns with an error after that time.
func (hc Controller) Running(w http.ResponseWriter, r *http.Request) APIErrors {
	ctx := r.Context()
	params := httprouter.ParamsFromContext(ctx)
	org := params.ByName("org")
	appName := params.ByName("app")

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

	app, err := application.Lookup(ctx, cluster, org, appName)
	if err != nil {
		return InternalError(err)
	}

	if app == nil {
		return AppIsNotKnown(appName)
	}

	if app.Workload == nil {
		// While the app exists it has no workload, and therefore no status
		return NewAPIError("No status available for application without workload",
			"", http.StatusBadRequest)
	}

	err = cluster.WaitForDeploymentCompleted(
		ctx, nil, org, appName, duration.ToAppBuilt())
	if err != nil {
		return InternalError(err)
	}

	err = response.JSON(w, models.ResponseOK)
	if err != nil {
		return InternalError(err)
	}
	return nil
}
