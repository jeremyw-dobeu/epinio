package env

import (
	"github.com/epinio/epinio/helpers/kubernetes"
	"github.com/epinio/epinio/internal/api/v1/deploy"
	"github.com/epinio/epinio/internal/api/v1/response"
	"github.com/epinio/epinio/internal/application"
	"github.com/epinio/epinio/internal/cli/server/requestctx"
	apierror "github.com/epinio/epinio/pkg/api/core/v1/errors"
	"github.com/epinio/epinio/pkg/api/core/v1/models"
	"github.com/gin-gonic/gin"
)

// Set handles the API endpoint /namespaces/:namespace/applications/:app/environment (POST)
// It receives the namespace, application name, var name and value,
// and add/modifies the variable in the  application's environment.
func (hc Controller) Set(c *gin.Context) apierror.APIErrors {
	ctx := c.Request.Context()
	log := requestctx.Logger(ctx)
	username := requestctx.User(ctx).Username

	namespaceName := c.Param("namespace")
	appName := c.Param("app")

	log.Info("processing environment variable assignment",
		"namespace", namespaceName, "app", appName)

	cluster, err := kubernetes.GetCluster(ctx)
	if err != nil {
		return apierror.InternalError(err)
	}

	app, err := application.Lookup(ctx, cluster, namespaceName, appName)
	if err != nil {
		return apierror.InternalError(err)
	}
	if app == nil {
		return apierror.AppIsNotKnown(appName)
	}

	var setRequest models.EnvVariableMap
	err = c.BindJSON(&setRequest)
	if err != nil {
		return apierror.BadRequest(err)
	}

	err = application.EnvironmentSet(ctx, cluster, app.Meta, setRequest, false)
	if err != nil {
		return apierror.InternalError(err)
	}

	if app.Workload != nil {
		_, apierr := deploy.DeployApp(ctx, cluster, app.Meta, username, "", nil, nil)
		if apierr != nil {
			return apierr
		}
	}

	response.OK(c)
	return nil
}
