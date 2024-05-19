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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	scanv1 "github.com/perepeli/clusterscan-operator/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

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

func (r *ClusterScanReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("clusterscan", req.NamespacedName)

	// Fetch the ClusterScan instance
	var clusterScan scanv1.ClusterScan
	if err := r.Get(ctx, req.NamespacedName, &clusterScan); err != nil {
		if errors.IsNotFound(err) {
			// Resource not found, possibly was deleted - nothing to do
			return ctrl.Result{}, nil
		}
		// Error reading the object - log and requeue the request
		log.Error(err, "unable to fetch ClusterScan")
		return ctrl.Result{}, err
	}

	var unifiedStatus, message string

	// Check if the scan is scheduled or a one-off
	if clusterScan.Spec.Schedule != "" {
		// Handle CronJob creation or update
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

		// Check if CronJob already exists
		existingCronJob := &batchv1.CronJob{}
		err := r.Get(ctx, client.ObjectKey{Name: cronJob.Name, Namespace: cronJob.Namespace}, existingCronJob)
		if err != nil {
			if errors.IsNotFound(err) {
				// Create the CronJob
				if err := r.Create(ctx, cronJob); err != nil {
					log.Error(err, "unable to create CronJob for ClusterScan", "cronJob", cronJob)
					return ctrl.Result{}, err
				}
				log.Info("created CronJob for ClusterScan", "cronJob", cronJob)
				unifiedStatus = "Scheduled"
				message = "CronJob has been scheduled."
			} else {
				log.Error(err, "unable to fetch existing CronJob for ClusterScan", "cronJob", cronJob)
				return ctrl.Result{}, err
			}
		} else {
			// Reflect the status of the existing CronJob
			if existingCronJob.Spec.Suspend != nil && *existingCronJob.Spec.Suspend {
				unifiedStatus = "Suspended"
				message = "CronJob is suspended."
			} else if len(existingCronJob.Status.Active) > 0 {
				unifiedStatus = "Running"
				message = "CronJob is currently running."
			} else if existingCronJob.Status.LastScheduleTime != nil {
				unifiedStatus = "Scheduled"
				message = "CronJob is scheduled and awaiting next run."
			} else {
				unifiedStatus = "Inactive"
				message = "CronJob is inactive."
			}
		}
	} else {
		// Handle one-off Job creation or update
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

		// Check if Job already exists
		existingJob := &batchv1.Job{}
		err := r.Get(ctx, client.ObjectKey{Name: job.Name, Namespace: job.Namespace}, existingJob)
		if err != nil {
			if errors.IsNotFound(err) {
				// Create the Job
				if err := r.Create(ctx, job); err != nil {
					log.Error(err, "unable to create Job for ClusterScan", "job", job)
					return ctrl.Result{}, err
				}
				log.Info("created Job for ClusterScan", "job", job)
				unifiedStatus = "Active"
				message = "Job has been created and is active."
			} else {
				log.Error(err, "unable to fetch existing Job for ClusterScan", "job", job)
				return ctrl.Result{}, err
			}
		} else {
			// Reflect the status of the existing Job
			if existingJob.Status.Active > 0 {
				unifiedStatus = "Running"
				message = "Job is currently running."
			} else if existingJob.Status.Succeeded > 0 {
				unifiedStatus = "Succeeded"
				message = "Job has completed successfully."
			} else if existingJob.Status.Failed > 0 {
				unifiedStatus = "Failed"
				message = "Job has failed."
			} else {
				unifiedStatus = "Pending"
				message = "Job is pending."
			}
		}
	}

	// Update status
	now := metav1.NewTime(time.Now())
	clusterScan.Status.LastScheduleTime = &now
	clusterScan.Status.UnifiedStatus = unifiedStatus
	clusterScan.Status.Message = message
	if err := r.Status().Update(ctx, &clusterScan); err != nil {
		log.Error(err, "unable to update ClusterScan status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func getCronJobMessage(cronJob *batchv1.CronJob) string {
	if len(cronJob.Status.Active) > 0 {
		return fmt.Sprintf("CronJob is active with %d active jobs", len(cronJob.Status.Active))
	}
	return "CronJob is inactive"
}

func (r *ClusterScanReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scanv1.ClusterScan{}). // Watch the primary resource: ClusterScan
		Owns(&batchv1.Job{}).       // Watch the owned resource: Job
		Owns(&batchv1.CronJob{}).   // Watch the owned resource: CronJob
		Complete(r)                 // Register the reconciler with the controller
}
