/*
Copyright 2024.

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
	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"

	scanv1 "github.com/perepeli/clusterscan-operator/api/v1"
)

// ClusterScanReconciler reconciles a ClusterScan object
type ClusterScanReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=scan.example.com,resources=clusterscans,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scan.example.com,resources=clusterscans/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=scan.example.com,resources=clusterscans/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterScan object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *ClusterScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := r.Log.WithValues("clusterscan", req.NamespacedName)

	// Fetch the ClusterScan instance
	var clusterScan scanv1.ClusterScan
	if err := r.Get(ctx, req.NamespacedName, &clusterScan); err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch ClusterScan")
		return ctrl.Result{}, err
	}

	// Check if the scan is scheduled or a one-off
	if clusterScan.Spec.Schedule != "" {
		// Handle CronJob creation
		cronJob := &batchv1.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterScan.Name + "-cron",
				Namespace: clusterScan.Namespace,
			},
			Spec: batchv1.CronJobSpec{
				Schedule: clusterScan.Spec.Schedule,
				JobTemplate: batchv1.JobTemplateSpec{
					Spec: clusterScan.Spec.JobTemplate,
				},
			},
		}
		// Set ClusterScan instance as the owner and controller
		if err := controllerutil.SetControllerReference(&clusterScan, cronJob, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, cronJob); err != nil {
			log.Error(err, "unable to create CronJob for ClusterScan", "cronJob", cronJob)
			return ctrl.Result{}, err
		}
		log.Info("created CronJob for ClusterScan", "cronJob", cronJob)
	} else {
		// Handle one-off Job creation
		job := &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterScan.Name + "-job",
				Namespace: clusterScan.Namespace,
			},
			Spec: clusterScan.Spec.JobTemplate,
		}
		// Set ClusterScan instance as the owner and controller
		if err := controllerutil.SetControllerReference(&clusterScan, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, job); err != nil {
			log.Error(err, "unable to create Job for ClusterScan", "job", job)
			return ctrl.Result{}, err
		}
		log.Info("created Job for ClusterScan", "job", job)
	}

	// Update status
	now := metav1.NewTime(time.Now())
	clusterScan.Status.LastScheduleTime = &now
	if err := r.Status().Update(ctx, &clusterScan); err != nil {
		log.Error(err, "unable to update ClusterScan status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ClusterScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scanv1.ClusterScan{}).
		Owns(&batchv1.Job{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
