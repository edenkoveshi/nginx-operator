/*
Copyright 2021.

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

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"reflect"
	"time"

	nginxv1alpha1 "eden.koveshi/nginx/nginx-webserver/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NGINXWebServerReconciler reconciles a NGINXWebServer object
type NGINXWebServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=nginx.eden.koveshi,resources=nginxwebservers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=nginx.eden.koveshi,resources=nginxwebservers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=nginx.eden.koveshi,resources=nginxwebservers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NGINXWebServer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *NGINXWebServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// your logic here

	nginx := &nginxv1alpha1.NGINXWebServer{}
	err := r.Get(ctx, req.NamespacedName, nginx)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Nginx resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Nginx")
		return ctrl.Result{}, err
	}

	found := &appsv1.Deployment{}
	err = r.Get(ctx, types.NamespacedName{Name: nginx.Name, Namespace: nginx.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.newDeployment(nginx)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}

	current_replicas := nginx.Spec.Replicas
	containers := found.Spec.Template.Spec.Containers
	images := make([]string, len(containers))
	for i := 0; i < len(containers); i++ {
		images = append(images, containers[i].Image)
	}
	//log.Info("", "Listing", "images:", images)
	//log.Info("", "Listing", "containers:", containers)
	if *found.Spec.Replicas != current_replicas {
		found.Spec.Replicas = &current_replicas
		err = r.Update(ctx, found)
		if err != nil {
			log.Error(err, "Failed to update Deployment", "Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
			return ctrl.Result{}, err
		}
		// Ask to requeue after 1 minute in order to give enough time for the
		// pods be created on the cluster side and the operand be able
		// to do the next update step accurately.
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Update the nginx status with the pod names
	// List the pods for this nginx's deployment
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(nginx.Namespace),
		client.MatchingLabels(r.setLabels(nginx.Name)),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		log.Error(err, "Failed to list pods", "nginx.Namespace", nginx.Namespace, "nginx.Name", nginx.Name)
		return ctrl.Result{}, err
	}
	podNames := getPodNames(podList.Items)

	// Update status.Replicas if needed
	if !reflect.DeepEqual(podNames, nginx.Status.Replicas) {
		nginx.Status.Replicas = podNames
		err := r.Status().Update(ctx, nginx)
		if err != nil {
			log.Error(err, "Failed to update nginx status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{RequeueAfter: time.Minute}, nil
}

func (r *NGINXWebServerReconciler) newDeployment(m *nginxv1alpha1.NGINXWebServer) *appsv1.Deployment {
	ls := r.setLabels(m.Name)
	replicas := m.Spec.Replicas
	image := m.Spec.Image

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:   image,
						Name:    "nginx-web-server",
						Command: []string{"nginx", "-g", "daemon off;"},
						Ports: []corev1.ContainerPort{{
							ContainerPort: 80,
							Name:          "nginx",
						}},
					}},
				},
			},
		},
	}
	// Set Nginx instance as the owner and controller
	ctrl.SetControllerReference(m, dep, r.Scheme)
	return dep
}

func (r *NGINXWebServerReconciler) setLabels(name string) map[string]string {
	return map[string]string{"app": "nginx", "name": name}
}

// getPodNames returns the pod names of the array of pods passed in
func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

// SetupWithManager sets up the controller with the Manager.
func (r *NGINXWebServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nginxv1alpha1.NGINXWebServer{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
