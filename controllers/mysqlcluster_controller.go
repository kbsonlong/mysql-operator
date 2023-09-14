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
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "github.com/kbsonlong/mysql-operator/api/v1"
	"github.com/kbsonlong/mysql-operator/pkg/k8s"
	"github.com/kbsonlong/mysql-operator/pkg/mysql"
	apps "k8s.io/api/apps/v1"
	k8scorev1 "k8s.io/api/core/v1"
)

// MysqlClusterReconciler reconciles a MysqlCluster object
type MysqlClusterReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

const (
	NameLabel         = "app.kubernetes.io/name"
	InstanceLabel     = "app.kubernetes.io/instance"
	ManagedByLabel    = "app.kubernetes.io/managed-by"
	PartOfLabel       = "app.kubernetes.io/part-of"
	ComponentLabel    = "app.kubernetes.io/component"
	MySQLPrimaryLabel = "mysql.alongparty.cn/primary"
)

type MasterInfo struct {
	master_log_file string
	master_log_pos  int
	node            NodeInfo
	binlog          BinLogInfo
}

type NodeInfo struct {
	user     string
	password string
	host     string
	port     int
}
type BinLogInfo struct {
	File              string
	Position          int
	Binlog_Do_DB      string
	Binlog_Ignore_DB  string
	Executed_Gtid_Set string
}

//+kubebuilder:rbac:groups=batch.alongparty.cn,resources=mysqlclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.alongparty.cn,resources=mysqlclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.alongparty.cn,resources=mysqlclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

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
	err = k8s.CreateConfig(r.Client, ctx, req, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = k8s.CretaeOrUpdateSecret(r.Client, ctx, req, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}
	err = r.Get(ctx, req.NamespacedName, statefulset)
	r.Recorder.Event(cluster, k8scorev1.EventTypeNormal, "CreateStatefulSet", fmt.Sprintf("Create StatefulSet %s/%s Now", req.Namespace, cluster.Name))
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Create StatefulSet")
			err = r.doReconcileStatefulSet(ctx, cluster)
			if err != nil {
				r.Recorder.Event(cluster, k8scorev1.EventTypeWarning, "FailedCreateStatefulSet", err.Error())
				return ctrl.Result{}, err
			}
			r.Recorder.Event(cluster, k8scorev1.EventTypeNormal, "Created", fmt.Sprintf("Created StatefulSet %s/%s", req.Namespace, cluster.Name))
			// return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// 创建svc
	err = r.doReconcileService(ctx, cluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	// 检查容器ready数量与副本数是否一致
	Pods, err := r.getPods(ctx, cluster)
	if err != nil && !errors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	fmt.Println(len(Pods.Items))
	var i int
	for _, pod := range Pods.Items {
		if pod.Status.ContainerStatuses[0].Ready {
			fmt.Println(pod.ObjectMeta.Name)
			i++
		}
	}
	if i != int(*cluster.Spec.Replicas) {
		fmt.Println("Pod Ready less than replica")
		return ctrl.Result{}, nil
	}

	// TODO(user): 初始化 MySQL 主从集群
	fmt.Println("###################################")
	fmt.Println("Starting Initializing Mysql Cluster")
	if !cluster.Status.Initialized {
		err = r.InitMysqlCluster(ctx, cluster)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	// 更新crd状态
	defer func() {
		err = r.updateStatus(ctx, cluster, statefulset)
		if err != nil {
			log.Error(err, "failed to update cluster status", "replset")
		}
	}()

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.MysqlCluster{}).
		Owns(&apps.ReplicaSet{}).
		Complete(r)
}

func (r *MysqlClusterReconciler) updateStatus(ctx context.Context, cluster *batchv1.MysqlCluster, sts *apps.StatefulSet) error {
	log := log.FromContext(ctx)
	log.Info("update mysqlcluster state")

	cluster.Status.Replica = *cluster.Spec.Replicas
	time_now := time.Now().Nanosecond()
	fmt.Println(int32(time_now))
	cluster.Status.LastScheduleTime = int32(time_now)
	cluster.Status.Initialized = true
	// 必须使用 r.Status().Update() 更新，否则不会展示 Status 字段
	err := r.Status().Update(ctx, cluster)

	if err != nil {
		r.Recorder.Event(cluster, k8scorev1.EventTypeWarning, "FailedUpdateStatus", err.Error())
		return err
	}
	return err
}

func (r *MysqlClusterReconciler) doReconcileStatefulSet(ctx context.Context, cluster *batchv1.MysqlCluster) error {

	log := log.FromContext(ctx)

	fileMode := int32(0644)
	volumes := []k8scorev1.Volume{
		ensureVolume("config", k8scorev1.VolumeSource{
			ConfigMap: &k8scorev1.ConfigMapVolumeSource{
				LocalObjectReference: k8scorev1.LocalObjectReference{
					Name: cluster.Name,
				},
				DefaultMode: &fileMode,
			},
		}),
		ensureVolume("dynamic", k8scorev1.VolumeSource{
			EmptyDir: &k8scorev1.EmptyDirVolumeSource{},
		}),
	}

	containers := []k8scorev1.Container{
		{
			Name:            cluster.Name,
			Image:           *&cluster.Spec.Image,
			ImagePullPolicy: "IfNotPresent",
			EnvFrom: []k8scorev1.EnvFromSource{
				{
					SecretRef: &k8scorev1.SecretEnvSource{
						LocalObjectReference: k8scorev1.LocalObjectReference{
							Name: k8s.GetSecretName(cluster.Name),
						},
					},
				},
			},
			// Env: []k8scorev1.EnvVar{
			// 	{
			// 		Name:  "TEST_ENV",
			// 		Value: "123456",
			// 	},
			// 	{
			// 		Name:  "MYSQL_ROOT_PASSWORD",
			// 		Value: "$(TEST_ENV)",
			// 	},
			// },
			Ports: []k8scorev1.ContainerPort{
				{
					Name:          "mysql",
					Protocol:      k8scorev1.ProtocolSCTP,
					ContainerPort: 3306,
				},
			},
			VolumeMounts: []k8scorev1.VolumeMount{
				{
					Name:      "config",
					MountPath: "/etc/mysql/conf.d/",
				},
				{
					Name:      "dynamic",
					MountPath: "/etc/mysql/mysql.conf.d/",
				},
				// {
				// 	Name:      "initdb",
				// 	MountPath: "/docker-entrypoint-initdb.d",
				// },
			},
		},
	}

	// var init_cmd string
	// init_cmd = fmt.Sprintf("echo -e \"[mysqld]\nserver-id=1$(echo $HOSTNAME | awk -F '-' '{print $NF}')\">/etc/mysql/mysql.conf.d/dynamic.cnf")
	initContainers := []k8scorev1.Container{
		{
			Name:            fmt.Sprintf("%s-init", cluster.Name),
			Image:           *&cluster.Spec.InitImage,
			ImagePullPolicy: "IfNotPresent",
			// Command:         []string{"/bin/sh", "-c", init_cmd},
			VolumeMounts: []k8scorev1.VolumeMount{
				{
					Name:      "dynamic",
					MountPath: "/etc/mysql/mysql.conf.d/",
				},
			},
		},
	}

	statefulset := &apps.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      cluster.Name,
		},
		Spec: apps.StatefulSetSpec{
			Replicas: pointer.Int32(*cluster.Spec.Replicas),
			Selector: &metav1.LabelSelector{
				MatchLabels: r.Labels(cluster),
			},

			Template: k8scorev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: r.Labels(cluster),
				},
				Spec: k8scorev1.PodSpec{
					InitContainers: initContainers,
					Containers:     containers,
					Volumes:        volumes,
				},
			},
		},
	}

	// statefulset 与 crd 资源建立关联,
	// 建立关联后，删除 crd 资源时就会将 statefulset 也删除掉
	log.Info("set sts reference")
	if err := controllerutil.SetControllerReference(cluster, statefulset, r.Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}
	err := r.Create(ctx, statefulset)
	if err != nil {
		return err
	}

	return nil
}

func ensureVolume(name string, source k8scorev1.VolumeSource) k8scorev1.Volume {
	return k8scorev1.Volume{
		Name:         name,
		VolumeSource: source,
	}
}

func (r *MysqlClusterReconciler) doReconcileService(ctx context.Context, cluster *batchv1.MysqlCluster) error {

	log := log.FromContext(ctx)
	// log := r.Log.WithValues("func", "doReconcileService")
	selector := r.Labels(cluster)
	svc := &k8scorev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      cluster.Name,
			Labels:    selector,
		},
		Spec: k8scorev1.ServiceSpec{
			Ports: []k8scorev1.ServicePort{
				{
					Name:     "mysql",
					Port:     3306,
					Protocol: k8scorev1.ProtocolTCP,
				},
			},
			Selector: selector,
			Type:     k8scorev1.ServiceTypeClusterIP,
		},
	}

	// service 与 crd 资源建立关联,
	// 建立关联后，删除 crd 资源时就会将 service 也删除掉
	log.Info("set svc reference")
	if err := controllerutil.SetControllerReference(cluster, svc, r.Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}
	// 创建service
	log.Info("start create service")
	if err := r.Create(ctx, svc); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, svc); err != nil {
				log.Error(err, "create service error")
				return err
			}
		}

		// return err
	}

	// 创建 headless service
	headlessService := &k8scorev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cluster.Namespace,
			Name:      fmt.Sprintf("%s-headless", cluster.Name),
			Labels:    selector,
		},
		Spec: k8scorev1.ServiceSpec{
			Ports: []k8scorev1.ServicePort{
				{
					Name:     "mysql",
					Port:     3306,
					Protocol: k8scorev1.ProtocolTCP,
				},
			},
			Selector:  selector,
			Type:      k8scorev1.ServiceTypeClusterIP,
			ClusterIP: "None",
		},
	}

	// service 与 crd 资源建立关联,
	// 建立关联后，删除 crd 资源时就会将 service 也删除掉
	log.Info("set reference")
	if err := controllerutil.SetControllerReference(cluster, headlessService, r.Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}

	log.Info("start create headlessService")
	if err := r.Create(ctx, headlessService); err != nil {
		if errors.IsAlreadyExists(err) {
			if err := r.Update(ctx, headlessService); err != nil {
				log.Error(err, "create headlessService error")
				return nil
			}
		}

		// return err
	}

	for i := 0; i < int(*cluster.Spec.Replicas); i++ {
		fmt.Println(i)
		pod_name := fmt.Sprintf("%s-%d", cluster.Name, i)
		selector = map[string]string{
			"statefulset.kubernetes.io/pod-name": pod_name,
			NameLabel:                            "mysql-server",
			InstanceLabel:                        cluster.Name,
			ManagedByLabel:                       "mysql-server-operator",
			PartOfLabel:                          "mysql-server",
		}
		pod_svc := &k8scorev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: cluster.Namespace,
				Name:      pod_name,
				Labels:    selector,
			},
			Spec: k8scorev1.ServiceSpec{
				Ports: []k8scorev1.ServicePort{
					{
						Name:     "mysql",
						Port:     3306,
						Protocol: k8scorev1.ProtocolTCP,
					},
				},
				Selector: selector,
				Type:     k8scorev1.ServiceTypeClusterIP,
			},
		}

		if err := controllerutil.SetControllerReference(cluster, pod_svc, r.Scheme); err != nil {
			log.Error(err, "SetControllerReference error")
			return err
		}

		log.Info("start create Pod Service")
		if err := r.Create(ctx, pod_svc); err != nil {
			if errors.IsAlreadyExists(err) {
				if err := r.Update(ctx, pod_svc); err != nil {
					log.Error(err, "create or update pod_svc error")
					return nil
				}
			}
		}

	}

	log.Info("create service success")

	return nil
}

func (r *MysqlClusterReconciler) InitMysqlCluster(ctx context.Context, cluster *batchv1.MysqlCluster) error {

	var info BinLogInfo
	var node NodeInfo
	var master_host, change_sql, slave_host string
	master_host = fmt.Sprintf("%s-0.%s", cluster.Name, cluster.Namespace)
	fmt.Println(master_host)
	dsn := mysql.GetDsn(map[string]interface{}{"user_name": "root", "password": "123456", "host": master_host})
	db := mysql.DbConnect(dsn, false)

	// 创建同步用户
	db.QueryRow("SELECT user,host FROM mysql.user WHERE user='slave'").Scan(&node.user, &node.host)

	fmt.Println(node.user, node.host)
	if node.user != "slave" {
		_, err := db.Exec("CREATE USER 'slave'@'%' IDENTIFIED WITH 'mysql_native_password' BY '123456' ")
		if err != nil {
			return err
		}
		_, err = db.Exec("GRANT SELECT, RELOAD, SUPER, REPLICATION SLAVE, REPLICATION CLIENT, SHOW VIEW ON *.* TO 'slave'@'%';")
		if err != nil {
			return err
		}
	}

	//stpt3：查询数据库
	if err := db.QueryRow("show master status").Scan(&info.File, &info.Position, &info.Binlog_Do_DB, &info.Binlog_Ignore_DB, &info.Executed_Gtid_Set); err != nil {
		fmt.Println("获取 Master 状态失败。。")
		return err
	}
	change_sql = fmt.Sprintf(`change master to master_host='%s',
		master_user='%s',
		master_password='%s',
		master_port=%d,
		master_log_file='%s',
		master_log_pos=%d,
		master_connect_retry=30;
		`, master_host, "slave", "123456", 3306, info.File, info.Position)

	fmt.Println(change_sql)
	slave_host = fmt.Sprintf("%s-1.%s", cluster.Name, cluster.Namespace)
	slave_dsn := mysql.GetDsn(map[string]interface{}{"username": "root", "password": "123456", "host": slave_host})
	slavedb := mysql.DbConnect(slave_dsn, false)

	slavedb.Exec(change_sql)
	_, err := slavedb.Exec(change_sql)
	if err != nil {
		return err
	}

	_, err = slavedb.Exec("start slave")
	if err != nil {
		return err
	}

	return nil

}

func (r *MysqlClusterReconciler) Labels(cluster *batchv1.MysqlCluster) map[string]string {
	return map[string]string{
		NameLabel:      "mysql-server",
		InstanceLabel:  cluster.Name,
		ManagedByLabel: "mysql-server-operator",
		PartOfLabel:    "mysql-server",
	}
}

func (r *MysqlClusterReconciler) getPods(ctx context.Context, cluster *batchv1.MysqlCluster) (k8scorev1.PodList, error) {
	Pods := k8scorev1.PodList{}
	err := r.Client.List(ctx,
		&Pods,
		&client.ListOptions{
			Namespace:     cluster.Namespace,
			LabelSelector: labels.SelectorFromSet(r.Labels(cluster)),
		},
	)
	return Pods, err
}
