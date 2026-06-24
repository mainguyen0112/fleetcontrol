/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	fleetcontrolv1alpha1 "github.com/mainguyen0112/fleetcontrol/operator/api/v1alpha1"
)

// SatelliteReconciler reconciles a Satellite object
type SatelliteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fleetcontrol.fleetcontrol.io,resources=satellites,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fleetcontrol.fleetcontrol.io,resources=satellites/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fleetcontrol.fleetcontrol.io,resources=satellites/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Satellite object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.24.1/pkg/reconcile
func (r *SatelliteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("Reconciling satellite...", "name", req.Name, "namespace", req.Namespace)

	var satellite fleetcontrolv1alpha1.Satellite
	if err := r.Get(ctx, req.NamespacedName, &satellite); err != nil {
		if apierrors.IsNotFound(err) {
			// Satellite was deleted, nothing to do for now (finalizer logic comes in Phase 6)
			log.Info("Satellite resource not found, likely deleted", "name", req.Name)
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Satellite")
		return ctrl.Result{}, err
	}

	// Phase 1 MVP: simply mark the satellite as Ready once reconciled.
	// Control Plane API integration comes in Phase 7.
	if satellite.Status.Phase != "Ready" {
		satellite.Status.Phase = "Ready"
		satellite.Status.ManagedBy = "operator"

		if err := r.Status().Update(ctx, &satellite); err != nil {
			log.Error(err, "unable to update Satellite status")
			return ctrl.Result{}, err
		}
		log.Info("Satellite marked as Ready", "name", satellite.Name)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SatelliteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fleetcontrolv1alpha1.Satellite{}).
		Named("satellite").
		Complete(r)
}
