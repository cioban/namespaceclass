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
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	//rbacv1 "k8s.io/api/rbac/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//appsv1 "k8s.io/api/apps/v1"
	//batchv1 "k8s.io/api/batch/v1"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	loglib "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

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
	log := loglib.FromContext(ctx)

	// reconcile namespace obj
	ns := &corev1.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, ns); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "error retrieving the namespace")
		return ctrl.Result{}, err
	}
	if ns != nil {
		return r.NamespaceReconcile(ctx, req, ns, &log)
	}

	// reconcile namespaceclass obj
	nsclass := &namespaceclassv1.NamespaceClass{}
	if err := r.Get(ctx, req.NamespacedName, nsclass); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "error retrieving the namespaceclass")
		return ctrl.Result{}, err
	}
	if ns != nil {
		log.Info("NSCLASS")
		return ctrl.Result{}, nil
		//	return r.NamespaceReconcile(ctx, req, ns, &log)
	}

	return ctrl.Result{}, nil
}

func (r *NamespaceClassReconciler) NamespaceClassReconcile(ctx context.Context, req ctrl.Request, namespaceclass *namespaceclassv1.NamespaceClass, log *logr.Logger) (ctrl.Result, error) {
	log.Info(" - Namespaceclass reconciliation")

	return ctrl.Result{}, nil
}

func (r *NamespaceClassReconciler) NamespaceReconcile(ctx context.Context, req ctrl.Request, namespace *corev1.Namespace, log *logr.Logger) (ctrl.Result, error) {

	// Namespace is being deleted
	if namespace.DeletionTimestamp != nil {
		log.Info("Namespace is deleted. Skipping...")
		return ctrl.Result{}, nil
	}

	if _, label := namespace.Labels[ExpectedLabel]; !label {
		return ctrl.Result{}, nil
	}

	log.Info(" - Namespace reconciliation")
	// Get the NSCLASS for de NS
	nsclass := &namespaceclassv1.NamespaceClass{}
	if err := r.Get(ctx, client.ObjectKey{Name: namespace.Labels[ExpectedLabel]}, nsclass); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "error retrieving the namespaceclass")
		return ctrl.Result{}, err
	}

	// Manage the NS resources set by the NSCLASS
	log.Info(" - NS resources")
	for _, resource := range nsclass.Spec.NamespaceResources {
		u := &unstructured.Unstructured{}
		if err := json.Unmarshal(resource.Raw, &u.Object); err != nil {
			log.Error(err, "Error loading resouce object")
			return ctrl.Result{}, err
		}
		u.SetNamespace(namespace.Name)
		if err := r.CreateOrUpdateResource(ctx, nsclass, u, log); err != nil {
			log.Error(err, "Error Creating/Updating resource")
		}

	}

	return ctrl.Result{}, nil
}

func (r *NamespaceClassReconciler) CreateOrUpdateResource(ctx context.Context, nsclass *namespaceclassv1.NamespaceClass, resource *unstructured.Unstructured, log *logr.Logger) error {
	liveResource := resource.DeepCopy()
	errR := r.Get(ctx, types.NamespacedName{
		Name:      resource.GetName(),
		Namespace: resource.GetNamespace(),
	}, liveResource)
	if errR != nil && !apierrors.IsNotFound(errR) {
		log.Error(errR, "Failed to get current resource", "resource", resource)
		return errR
	}

	if err := controllerutil.SetControllerReference(nsclass, resource, r.Scheme); err != nil {
		return err
	}

	if apierrors.IsNotFound(errR) {
		log.Info(" >> Creating NS resource", "resource", resource)
		return r.Create(ctx, resource)
	}

	if resource != liveResource {
		liveResource = resource

		log.Info(" >> Updating NS resource", "resource", resource)
		return r.Update(ctx, liveResource)
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceClassReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&namespaceclassv1.NamespaceClass{}).
		WithEventFilter(predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return true
			},
			CreateFunc: func(e event.CreateEvent) bool {
				return true
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		}).
		Watches(&corev1.Namespace{}, &handler.EnqueueRequestForObject{}, builder.WithPredicates(
			predicate.Or(predicate.GenerationChangedPredicate{}, predicate.LabelChangedPredicate{}))).
		Named("namespaceclass").
		// Watching resources owned by the controller
		// For this example, the controller only supports reconciliation in Networkpolicy and serviceaccounts
		Owns(&networkingv1.NetworkPolicy{}).
		Owns(&corev1.ServiceAccount{}).
		Complete(r)
}
