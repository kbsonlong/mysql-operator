/*
Copyright 2023.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "github.com/kbsonlong/mysql-operator/api/v1"
	apps "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
)

// MysqlClusterReconciler reconciles a MysqlCluster object
type MysqlClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=batch.alongparty.cn,resources=mysqlclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.alongparty.cn,resources=mysqlclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.alongparty.cn,resources=mysqlclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=pod,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MysqlCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *MysqlClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// TODO(user): 创建 MySQL sts
	cluster := &batchv1.MysqlCluster{}
	statefulset := &apps.StatefulSet{}
	err := r.Get(ctx, req.NamespacedName, cluster)
	if err != nil {
		return ctrl.Result{}, nil
	}
	err = r.Get(ctx, req.NamespacedName, statefulset)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Create StatefulSet", cluster.Name)
			err = r.CreateStatefulSet(ctx, cluster)
			if err != nil {
				r.Recorder.Event(cluster, k8scorev1.EventTypeWarning, "FailedCreateStatefulSet", err.Error())
				return ctrl.Result{}, err
			}
			cluster.Status.Replicas = *cluster.Spec.Replicas
			err = r.Update(ctx, cluster)
			if err != nil {
				r.Recorder.Event(cluster, k8scorev1.EventTypeWarning, "FailedUpdateStatus", err.Error())
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, err
	}
	//binding StatefulSet to MysqlCluster
	if err = ctrl.SetControllerReference(cluster, statefulset, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// TODO(user): 初始化 MySQL 主从集群
	// _ = r.InitMysqlCluster(ctx, cluster)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.MysqlCluster{}).
		Owns(&apps.ReplicaSet{}).
		Complete(r)
}

func (r *MysqlClusterReconciler) CreateStatefulSet(ctx context.Context, cluster *batchv1.MysqlCluster) error {
	statefulset := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      cluster.Name,
		},
		Spec: apps.StatefulSetSpec{
			Replicas: pointer.Int32(*cluster.Spec.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": cluster.Name,
				},
			},

			Template: k8scorev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": cluster.Name,
					},
				},
				Spec: k8scorev1.PodSpec{
					Containers: []k8scorev1.Container{
						{
							Name:            cluster.Name,
							Image:           *&cluster.Spec.Image,
							ImagePullPolicy: "IfNotPresent",
							Env: []k8scorev1.EnvVar{
								{
									Name:  "MYSQL_ROOT_PASSWORD",
									Value: "123456",
								},
							},
							Ports: []k8scorev1.ContainerPort{
								{
									Name:          "mysql",
									Protocol:      k8scorev1.ProtocolSCTP,
									ContainerPort: 3306,
								},
							},
						},
					},
				},
			},
		},
	}
	err := r.Create(ctx, statefulset)
	if err != nil {
		return err
	}
	return nil
}

// func (r *MysqlClusterReconciler) InitMysqlCluster(ctx context.Context, cluster *batchv1.MysqlCluster) error {
// 	pod := &k8scorev1.Pod{}
// 	err = r.Get(ctx, cluster.Namespace, pod)

// }
