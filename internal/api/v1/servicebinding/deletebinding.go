package servicebinding

import (
	"context"

	"github.com/epinio/epinio/helpers/kubernetes"
	"github.com/epinio/epinio/internal/application"
	"github.com/epinio/epinio/internal/services"

	. "github.com/epinio/epinio/pkg/api/core/v1/errors"
)

func DeleteBinding(ctx context.Context, cluster *kubernetes.Cluster, org, appName, serviceName, username string) APIErrors {

	app, err := application.Lookup(ctx, cluster, org, appName)
	if err != nil {
		return InternalError(err)
	}
	if app == nil {
		return AppIsNotKnown(appName)
	}

	_, err = services.Lookup(ctx, cluster, org, serviceName)
	if err != nil && err.Error() == "service not found" {
		return ServiceIsNotKnown(serviceName)
	}
	if err != nil {
		return InternalError(err)
	}

	// Take old state
	oldBound, err := application.BoundServiceNameSet(ctx, cluster, app.Meta)
	if err != nil {
		return InternalError(err)
	}

	err = application.BoundServicesUnset(ctx, cluster, app.Meta, serviceName)
	if err != nil {
		return InternalError(err)
	}

	if app.Workload != nil {
		// For this read the new set of bound services back,
		// as full service structures
		newBound, err := application.BoundServices(ctx, cluster, app.Meta)
		if err != nil {
			return InternalError(err)
		}

		err = application.NewWorkload(cluster, app.Meta).BoundServicesChange(ctx, username, oldBound, newBound)
		if err != nil {
			return InternalError(err)
		}
	}

	return nil
}
