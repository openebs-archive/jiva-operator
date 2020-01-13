/*
Copyright 2019 The OpenEBS Authors
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

package jivavolume

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/go-logr/logr"
	jv "github.com/openebs/jiva-operator/pkg/apis/openebs/v1alpha1"
	"github.com/openebs/jiva-operator/pkg/jiva"
	"github.com/openebs/jiva-operator/pkg/kubernetes/container"
	pts "github.com/openebs/jiva-operator/pkg/kubernetes/podtemplatespec"
	"github.com/openebs/jiva-operator/pkg/volume"
	operr "github.com/pkg/errors"

	deploy "github.com/openebs/jiva-operator/pkg/kubernetes/deployment"
	pvc "github.com/openebs/jiva-operator/pkg/kubernetes/pvc"
	svc "github.com/openebs/jiva-operator/pkg/kubernetes/service"
	sts "github.com/openebs/jiva-operator/pkg/kubernetes/statefulset"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	log           = logf.Log.WithName("controller_jivavolume")
	svcNameFormat = "%s-jiva-ctrl-svc.%s.svc.cluster.local"
)

const (
	pdbAPIVersion = "policyv1beta1"
)

var (
	installFuncs = []func(r *ReconcileJivaVolume, cr *jv.JivaVolume,
		reqLog logr.Logger) error{
		createControllerService,
		createControllerDeployment,
		createReplicaStatefulSet,
		createReplicaPodDisruptionBudget,
	}

	updateErrMsg = "Failed to update JivaVolume with service info"
)

// Add creates a new JivaVolume Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileJivaVolume{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("jivavolume-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource JivaVolume
	err = c.Watch(&source.Kind{Type: &jv.JivaVolume{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner JivaVolume
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &jv.JivaVolume{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &jv.JivaVolume{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &jv.JivaVolume{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileJivaVolume implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileJivaVolume{}

// ReconcileJivaVolume reconciles a JivaVolume object
type ReconcileJivaVolume struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a JivaVolume object and makes changes based on the state read
// and what is in the JivaVolume.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileJivaVolume) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the JivaVolume instance
	instance := &jv.JivaVolume{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// initially Phase will be "", so it will skip switch case
	// Once it has started boostrapping it will set the Phase to Pending/Failed
	// depends upon the error. If bootstrap is successful it will set the Phase
	// to syncing which will be changed to Ready later when volume becomes RW
	switch instance.Status.Phase {
	case jv.JivaVolumePhaseReady, jv.JivaVolumePhaseSyncing:
		return reconcile.Result{}, r.getAndUpdateVolumeStatus(instance, reqLogger)
	case jv.JivaVolumePhaseDeleting:
		reqLogger.Info("start tearing down jiva components", "JivaVolume: ", instance)
		return reconcile.Result{}, nil
	case jv.JivaVolumePhasePending, jv.JivaVolumePhaseFailed:
		reqLogger.Info("start bootstraping jiva components", "JivaVolume: ", instance)
		return reconcile.Result{}, r.bootstrapJiva(instance, reqLogger)
	}

	reqLogger.Info("start bootstraping jiva components")
	return reconcile.Result{}, r.bootstrapJiva(instance, reqLogger)
}

func (r *ReconcileJivaVolume) finally(err error, cr *jv.JivaVolume, reqLog logr.Logger) {
	if err != nil {
		cr.Status.Phase = jv.JivaVolumePhaseFailed
	} else {
		cr.Status.Phase = jv.JivaVolumePhaseSyncing
	}

	if err := r.updateJivaVolume(cr); err != nil {
		reqLog.Error(err, "failed to update JivaVolume phase")
	}
}

// 1. Create controller svc
// 2. Create controller deploy
// 3. Create replica statefulset
func (r *ReconcileJivaVolume) bootstrapJiva(cr *jv.JivaVolume, reqLog logr.Logger) (err error) {
	defer r.finally(err, cr, reqLog)

	for _, f := range installFuncs {
		if err = f(r, cr, reqLog); err != nil {
			return err
		}
	}

	return nil
}

// TODO: add logic to create disruption budget for replicas
func createReplicaPodDisruptionBudget(r *ReconcileJivaVolume, cr *jv.JivaVolume, reqLog logr.Logger) error {
	min := cr.Spec.Policy.Spec.Target.ReplicationFactor
	pdbObj := &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: pdbAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pdb",
			Namespace: cr.Namespace,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: defaultReplicaLabels(cr.Spec.PV),
			},
			MinAvailable: &intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: int32(min/2 + 1),
			},
		},
	}

	instance := &policyv1beta1.PodDisruptionBudget{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: pdbObj.Name, Namespace: pdbObj.Namespace}, instance)
	if err != nil && errors.IsNotFound(err) {
		// Set JivaVolume instance as the owner and controller
		if err := controllerutil.SetControllerReference(cr, pdbObj, r.scheme); err != nil {
			return err
		}

		reqLog.V(2).Info("Creating a new pod disruption budget", "Pdb.Namespace", pdbObj.Namespace, "Pdb.Name", pdbObj.Name)
		err = r.client.Create(context.TODO(), pdbObj)
		if err != nil {
			return err
		}
		// pdb created successfully - don't requeue
		return nil
	} else if err != nil {
		return operr.Wrapf(err, "failed to get the pod disruption budget details: %v", pdbObj.Name)
	}

	return nil
}

func createControllerDeployment(r *ReconcileJivaVolume, cr *jv.JivaVolume,
	reqLog logr.Logger) error {
	reps := int32(1)

	dep, err := deploy.NewBuilder().WithName(cr.Name + "-jiva-ctrl").
		WithNamespace(cr.Namespace).
		WithLabels(defaultControllerLabels(cr.Spec.PV)).
		WithReplicas(&reps).
		WithStrategyType(appsv1.RecreateDeploymentStrategyType).
		WithSelectorMatchLabelsNew(defaultControllerLabels(cr.Spec.PV)).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabels(defaultControllerLabels(cr.Spec.PV)).
				WithAnnotations(defaultAnnotations()).
				WithContainerBuilders(
					container.NewBuilder().
						WithName("jiva-controller").
						WithImage(getImage("OPENEBS_IO_JIVA_CONTROLLER_IMAGE",
							"jiva-controller")).
						WithPortsNew(defaultControllerPorts()).
						WithCommandNew([]string{
							"launch",
						}).
						WithArgumentsNew([]string{
							"controller",
							"--frontend",
							"gotgt",
							"--clusterIP",
							cr.Spec.ISCSISpec.TargetIP,
							cr.Name,
						}).
						WithEnvsNew([]corev1.EnvVar{
							{
								Name:  "REPLICATION_FACTOR",
								Value: strconv.Itoa(cr.Spec.Policy.Spec.Target.ReplicationFactor),
							},
						}).
						WithResources(cr.Spec.Policy.Spec.Target.Resources).
						WithImagePullPolicy(corev1.PullIfNotPresent),
					container.NewBuilder().
						WithImage(getImage("OPENEBS_IO_MAYA_EXPORTER_IMAGE",
							"exporter")).
						WithName("maya-volume-exporter").
						WithCommandNew([]string{"maya-exporter"}).
						WithPortsNew([]corev1.ContainerPort{
							{
								ContainerPort: 9500,
								Protocol:      "TCP",
							},
						},
						),
				),
		).Build()

	if err != nil {
		return fmt.Errorf("failed to build deployment object, err: %v", err)
	}

	instance := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, instance)
	if err != nil && errors.IsNotFound(err) {
		// Set JivaVolume instance as the owner and controller
		if err := controllerutil.SetControllerReference(cr, dep, r.scheme); err != nil {
			return err
		}

		reqLog.V(2).Info("Creating a new deployment", "Deploy.Namespace", dep.Namespace, "Deploy.Name", dep.Name)
		err = r.client.Create(context.TODO(), dep)
		if err != nil {
			return err
		}
		// deployment created successfully - don't requeue
		return nil
	} else if err != nil {
		return operr.Wrapf(err, "failed to get the deployment details: %v", dep.Name)
	}

	return nil
}

func getImage(key, component string) string {
	image, present := os.LookupEnv(key)
	if !present {
		switch component {
		case "jiva-controller", "jiva-replica":
			image = "quay.io/openebs/jiva:ci"
		case "exporter":
			image = "quay.io/openebs/m-exporter:ci"
		}
	}
	return image
}

func defaultReplicaLabels(pv string) map[string]string {
	return map[string]string{
		"openebs.io/cas-type":          "jiva",
		"openebs.io/component":         "jiva-replica",
		"openebs.io/persistent-volume": pv,
	}
}

func defaultControllerLabels(pv string) map[string]string {
	return map[string]string{
		"openebs.io/cas-type":          "jiva",
		"openebs.io/component":         "jiva-controller",
		"openebs.io/persistent-volume": pv,
	}
}

func defaultAnnotations() map[string]string {
	return map[string]string{"prometheus.io/path": "/metrics",
		"prometheus.io/port":  "9500",
		"prometheus.io/scrap": "true",
	}
}

func defaultControllerPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			ContainerPort: 3260,
			Protocol:      "TCP",
		},
		{
			ContainerPort: 9501,
			Protocol:      "TCP",
		},
	}
}

func defaultControllerSVCPorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name:       "iscsi",
			Port:       3260,
			Protocol:   "TCP",
			TargetPort: intstr.IntOrString{IntVal: 3260},
		},
		{
			Name:       "api",
			Port:       9501,
			Protocol:   "TCP",
			TargetPort: intstr.IntOrString{IntVal: 9501},
		},
		{
			Name:       "m-exporter",
			Port:       9500,
			Protocol:   "TCP",
			TargetPort: intstr.IntOrString{IntVal: 9500},
		},
	}
}

func defaultReplicaPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			ContainerPort: 9502,
			Protocol:      "TCP",
		},
		{
			ContainerPort: 9503,
			Protocol:      "TCP",
		},
		{
			ContainerPort: 9504,
			Protocol:      "TCP",
		},
	}
}

func defaultServiceLabels(pv string) map[string]string {
	return map[string]string{
		"openebs.io/cas-type":          "jiva",
		"openebs.io/component":         "jiva-controller-service",
		"openebs.io/persistent-volume": pv,
	}
}

// TODO: Add code to configure resource limits, nodeAffinity etc.
func createReplicaStatefulSet(r *ReconcileJivaVolume, cr *jv.JivaVolume,
	reqLog logr.Logger) error {

	var (
		err                            error
		replicaCount                   int32
		stsObj                         *appsv1.StatefulSet
		blockOwnerDeletion, controller = true, true
	)
	rc := cr.Spec.Policy.Spec.Target.ReplicationFactor
	replicaCount = int32(rc)
	prev := true

	size := strings.Split(cr.Spec.Capacity, "i")[0]
	capacity, err := units.RAMInBytes(size)
	if err != nil {
		return fmt.Errorf("failed to convert human readable size: %v into int64, err: %v", cr.Spec.Capacity, err)
	}

	stsObj, err = sts.NewBuilder().
		WithName(cr.Name + "-jiva-rep").
		WithLabelsNew(defaultReplicaLabels(cr.Spec.PV)).
		WithNamespace(cr.Namespace).
		WithServiceName("jiva-replica-svc").
		WithPodManagementPolicy(appsv1.ParallelPodManagement).
		WithStrategyType(appsv1.RollingUpdateStatefulSetStrategyType).
		WithReplicas(&replicaCount).
		WithSelectorMatchLabels(defaultReplicaLabels(cr.Spec.PV)).
		WithPodTemplateSpecBuilder(
			pts.NewBuilder().
				WithLabels(defaultReplicaLabels(cr.Spec.PV)).
				WithAffinity(&corev1.Affinity{
					PodAntiAffinity: &corev1.PodAntiAffinity{
						RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
							{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: defaultReplicaLabels(cr.Spec.PV),
								},
								TopologyKey: "kubernetes.io/hostname",
							},
						},
					},
				}).
				WithContainerBuilders(
					container.NewBuilder().
						WithName("jiva-replica").
						WithImage(getImage("OPENEBS_IO_JIVA_REPLICA_IMAGE",
							"jiva-replica")).
						WithPortsNew(defaultReplicaPorts()).
						WithCommandNew([]string{
							"launch",
						}).
						WithArgumentsNew([]string{
							"replica",
							"--frontendIP",
							fmt.Sprintf(svcNameFormat, cr.Name, cr.Namespace),
							"--size",
							fmt.Sprint(capacity),
							"openebs",
						}).
						WithImagePullPolicy(corev1.PullIfNotPresent).
						WithPrivilegedSecurityContext(&prev).
						WithResources(cr.Spec.Policy.Spec.Replica.Resources).
						WithVolumeMountsNew([]corev1.VolumeMount{
							{
								Name:      "openebs",
								MountPath: "/openebs",
							},
						}),
				),
		).
		WithPVC(
			pvc.NewBuilder().
				WithName("openebs").
				WithNamespace(cr.Namespace).
				WithOwnerReferenceNew([]metav1.OwnerReference{{
					APIVersion:         cr.APIVersion,
					BlockOwnerDeletion: &blockOwnerDeletion,
					Controller:         &controller,
					Kind:               cr.Kind,
					Name:               cr.Name,
					UID:                cr.UID,
				},
				}).
				WithStorageClass(cr.Spec.Policy.Spec.ReplicaSC).
				WithAccessModes([]corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}).
				WithCapacity(cr.Spec.Capacity),
		).Build()

	if err != nil {
		return fmt.Errorf("failed to build statefulset object, err: %v", err)
	}

	instance := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: stsObj.Name, Namespace: stsObj.Namespace}, instance)
	if err != nil && errors.IsNotFound(err) {
		// Set JivaVolume instance as the owner and controller
		if err := controllerutil.SetControllerReference(cr, stsObj, r.scheme); err != nil {
			return err
		}

		reqLog.V(2).Info("Creating a new Statefulset", "Statefulset.Namespace", stsObj.Namespace, "Sts.Name", stsObj.Name)
		err = r.client.Create(context.TODO(), stsObj)
		if err != nil {
			return err
		}
		// Statefulset created successfully - don't requeue
		return nil
	} else if err != nil {
		return operr.Wrapf(err, "failed to get the statefulset details: %v", stsObj.Name)
	}

	return nil
}

func updateJivaVolumeWithServiceInfo(r *ReconcileJivaVolume, cr *jv.JivaVolume, reqLog logr.Logger) error {
	ctrlSVC := &v1.Service{}
	if err := r.client.Get(context.TODO(),
		types.NamespacedName{
			Name:      cr.Name + "-jiva-ctrl-svc",
			Namespace: cr.Namespace,
		}, ctrlSVC); err != nil {
		return fmt.Errorf("%s, err: %v", updateErrMsg, err)
	}
	cr.Spec.ISCSISpec.TargetIP = ctrlSVC.Spec.ClusterIP
	var found bool
	for _, port := range ctrlSVC.Spec.Ports {
		if port.Name == "iscsi" {
			found = true
			cr.Spec.ISCSISpec.TargetPort = port.Port
			cr.Spec.ISCSISpec.Iqn = "iqn.2016-09.com.openebs.jiva" + ":" + cr.Spec.PV
		}
	}

	if !found {
		return fmt.Errorf("%s, err: can't find targetPort in target service spec: {%+v}", updateErrMsg, ctrlSVC)
	}

	reqLog.V(2).Info("Updating JivaVolume with iscsi spec", "ISCSISpec", cr.Spec.ISCSISpec)
	cr.Status.Phase = jv.JivaVolumePhasePending
	if err := r.client.Update(context.TODO(), cr); err != nil {
		return fmt.Errorf("%s, err: %v, JivaVolume CR: {%+v}", updateErrMsg, err, cr)
	}

	// Update cr with the updated fields so that we don't get
	// resourceVersion changed error in next steps
	if err := r.getJivaVolume(cr); err != nil {
		return fmt.Errorf("%s, err: %v, JivaVolume CR: {%+v}", updateErrMsg, err, cr)
	}

	return nil
}

func createControllerService(r *ReconcileJivaVolume, cr *jv.JivaVolume,
	reqLog logr.Logger) error {

	// By default type is clusterIP
	svcObj, err := svc.NewBuilder().
		WithName(cr.Name + "-jiva-ctrl-svc").
		WithLabelsNew(defaultServiceLabels(cr.Spec.PV)).
		WithNamespace(cr.Namespace).
		WithSelectorsNew(map[string]string{
			"openebs.io/cas-type":          "jiva",
			"openebs.io/persistent-volume": cr.Spec.PV,
		}).
		WithPorts(defaultControllerSVCPorts()).
		Build()

	if err != nil {
		return fmt.Errorf("failed to build service object, err: %v", err)
	}

	instance := &v1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: svcObj.Name, Namespace: svcObj.Namespace}, instance)
	if err != nil && errors.IsNotFound(err) {
		// Set JivaVolume instance as the owner and controller
		if err := controllerutil.SetControllerReference(cr, svcObj, r.scheme); err != nil {
			return err
		}

		reqLog.V(2).Info("Creating a new service", "Service.Namespace", svcObj.Namespace, "Service.Name", svcObj.Name)
		err = r.client.Create(context.TODO(), svcObj)
		if err != nil {
			return err
		}
		// Wait for service to get created
		time.Sleep(1 * time.Second)
		return updateJivaVolumeWithServiceInfo(r, cr, reqLog)
	} else if err != nil {
		return operr.Wrapf(err, "failed to get the service details: %v", svcObj.Name)

	}

	return updateJivaVolumeWithServiceInfo(r, cr, reqLog)

}

func deleteResource(name, ns string, r *ReconcileJivaVolume, obj runtime.Object) error {
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, obj)
	if err != nil && errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}

	err = r.client.Delete(context.TODO(), obj)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileJivaVolume) updateJivaVolume(cr *jv.JivaVolume) error {
	if err := r.client.Update(context.TODO(), cr); err != nil {
		return fmt.Errorf("failed to update JivaVolume CR: {%+v}, err: %v", cr, err)
	}
	return nil
}

func (r *ReconcileJivaVolume) getJivaVolume(cr *jv.JivaVolume) error {
	instance := &jv.JivaVolume{}
	if err := r.client.Get(context.TODO(),
		types.NamespacedName{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		}, instance); err != nil {
		return err
	}

	// update cr with the latest change
	cr = instance.DeepCopy()
	return nil
}

// setdefaults set the default value
func setdefaults(cr *jv.JivaVolume) {
	cr.Status = jv.JivaVolumeStatus{
		Status: "Unknown",
		Phase:  jv.JivaVolumePhaseSyncing,
	}
}

func (r *ReconcileJivaVolume) updateStatus(err error, cr *jv.JivaVolume, reqLog logr.Logger) {
	if err != nil {
		setdefaults(cr)
	}
	if err := r.updateJivaVolume(cr); err != nil {
		reqLog.Error(err, "failed to update status")
	}
}

func (r *ReconcileJivaVolume) getAndUpdateVolumeStatus(cr *jv.JivaVolume, reqLog logr.Logger) (err error) {
	var (
		cli *jiva.ControllerClient
	)

	if err = r.getJivaVolume(cr); err != nil {
		return fmt.Errorf("Failed to getAndUpdateVolumeStatus, err: %v", err)
	}

	defer r.updateStatus(err, cr, reqLog)
	addr := cr.Spec.ISCSISpec.TargetIP + ":9501"
	if len(addr) == 0 {
		return fmt.Errorf("Failed to get volume stats: target address is empty")
	}
	cli = jiva.NewControllerClient(addr)
	stats := &volume.Stats{}
	err = cli.Get("/stats", stats)
	if err != nil {
		// log err only, as controller must be in container creating state
		// don't return err as it will dump stack trace unneccesary
		reqLog.V(2).Info("Failed to get volume stats", "err", err)
	}

	reqLog.V(2).Info("Update status", "Stats", stats)

	cr.Status.Status = stats.TargetStatus
	cr.Status.ReplicaCount = len(stats.Replicas)
	cr.Status.ReplicaStatuses = make([]jv.ReplicaStatus, len(stats.Replicas))

	for i, rep := range stats.Replicas {
		cr.Status.ReplicaStatuses[i].Address = rep.Address
		cr.Status.ReplicaStatuses[i].Mode = rep.Mode
	}

	if stats.TargetStatus == "RW" {
		cr.Status.Phase = jv.JivaVolumePhaseReady
	} else if stats.TargetStatus == "RO" {
		cr.Status.Phase = jv.JivaVolumePhaseSyncing
	}

	return nil
}
