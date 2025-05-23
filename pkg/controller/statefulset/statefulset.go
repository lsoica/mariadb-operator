package statefulset

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type StatefulSetReconciler struct {
	client.Client
}

func NewStatefulSetReconciler(client client.Client) *StatefulSetReconciler {
	return &StatefulSetReconciler{
		Client: client,
	}
}

func (r *StatefulSetReconciler) Reconcile(ctx context.Context, desiredSts *appsv1.StatefulSet) error {
	return r.ReconcileWithUpdates(ctx, desiredSts, true)
}

func (r *StatefulSetReconciler) ReconcileWithUpdates(ctx context.Context, desiredSts *appsv1.StatefulSet,
	shouldUpdate bool) error {
	key := client.ObjectKeyFromObject(desiredSts)
	var existingSts appsv1.StatefulSet
	if err := r.Get(ctx, key, &existingSts); err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("error getting StatefulSet: %v", err)
		}
		if err := r.Create(ctx, desiredSts); err != nil {
			return fmt.Errorf("error creating StatefulSet: %v", err)
		}
		return nil
	}

	// If StatefulSet is being deleted, don't update it
	if !existingSts.DeletionTimestamp.IsZero() {
		log.FromContext(ctx).V(1).Info("StatefulSet is being deleted. Skipping...")
		return nil
	}

	if shouldUpdate {
		patch := client.MergeFrom(existingSts.DeepCopy())
		existingSts.Spec.Template = desiredSts.Spec.Template
		existingSts.Spec.UpdateStrategy = desiredSts.Spec.UpdateStrategy
		existingSts.Spec.Replicas = desiredSts.Spec.Replicas
		return r.Patch(ctx, &existingSts, patch)
	}
	return nil
}
