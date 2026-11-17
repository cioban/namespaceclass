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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	namespaceclassv1 "github.com/cioban/namespaceclass/api/v1"
)

const (
	ExpectedLabel = "namespaceclass.cioban.cc/name"
)

// NamespaceClassReconciler reconciles a NamespaceClass object
type NamespaceClassReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=namespaceclass.cioban.cc,resources=namespaceclasses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=namespaceclass.cioban.cc,resources=namespaceclasses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=namespaceclass.cioban.cc,resources=namespaceclasses/finalizers,verbs=update
// +kubebuilder:rbac:groups=*,resources=*,verbs=update;create;delete;patch;list;watch;get

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *NamespaceClassReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("Reconciliation loop")

	// reconcile namespace obj
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, ns); err != nil {
		log.Error(err, "error retrieving the namespace")
		return ctrl.Result{}, err
	} else {
		return r.NamespaceReconcile(ctx, req, ns)
	}

	/*
		nsClass := ns.Labels[ExpectedLabel]
		log.Info("NSCLASS")
		log.Info(nsClass)

		return ctrl.Result{}, nil
	*/
}

func (r *NamespaceClassReconciler) NamespaceReconcile(ctx context.Context, req ctrl.Request, namespace *corev1.Namespace) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info(" - Namespace reconciliation")

	log.Info(namespace.Name)
	log.Info(namespace.Labels[ExpectedLabel])

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&namespaceclassv1.NamespaceClass{}).
		Watches(&corev1.Namespace{}, &handler.EnqueueRequestForObject{}).
		Named("namespaceclass").
		Complete(r)
}
