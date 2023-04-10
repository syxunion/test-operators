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
	"runtime/debug"

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	matev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	uccpsv1 "uccps_docs.domain/api/v1"
)

// DocumentReconciler reconciles a Document object
type DocumentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=uccps.uccps.document.domain,resources=documents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=uccps.uccps.document.domain,resources=documents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=uccps.uccps.document.domain,resources=documents/finalizers,verbs=update
//+kubebuilder:rbac:groups=uccps.uccps.document.domain,resources=documents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployment,verbs=get;update;patch;watch;list;delete;create
//+kubebuilder:rbac:groups=apps,resources=services,verbs=get;update;patch;watch;list;delete;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Document object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *DocumentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	fmt.Println(fmt.Sprintf("1,%v", req))
	fmt.Println(fmt.Sprintf("2,%s", debug.Stack()))
	// 实例化数据结构
	instance := &uccpsv1.Document{}

	// 通过客户端工具查询，查询条件是
	err := r.Get(ctx, req.NamespacedName, instance)

	if err != nil {

		// 如果没有实例，就返回空结果，这样外部就不再立即调用Reconcile方法了
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}

		// 返回错误信息给外部
		return ctrl.Result{}, err
	}

	// 查找deployment
	deployment := &appsv1.Deployment{}

	// 用客户端工具查询
	err = r.Get(ctx, req.NamespacedName, deployment)

	// 查找时发生异常，以及查出来没有结果的处理逻辑
	if err != nil {
		// 如果没有实例就要创建了
		if errors.IsNotFound(err) {

			// 先要创建service
			if err = createService(ctx, instance, r, req); err != nil {
				// 返回错误信息给外部
				return ctrl.Result{}, err
			}

			//创建 路由
			if err = createRoute(ctx, instance, r, req); err != nil {
				// 返回错误信息给外部
				return ctrl.Result{}, err
			}

			// 立即创建deployment
			if err = createDocumentDeployment(ctx, instance, r); err != nil {
				// 返回错误信息给外部
				return ctrl.Result{}, err
			}

			// 创建成功就可以返回了
			return ctrl.Result{}, nil
		} else {
			// 返回错误信息给外部
			return ctrl.Result{}, err
		}
	}

	// 通过客户端更新deployment
	//if err = r.Update(ctx, deployment); err != nil {
	//	// 返回错误信息给外部
	//	return ctrl.Result{}, err
	//}
	//判断字段是否更新

	if err = updateDocumentDeployment(ctx, instance, r); err != nil {
		// 返回错误信息给外部
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DocumentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&uccpsv1.Document{}).
		Complete(r)
}

func createDocumentDeployment(ctx context.Context, document *uccpsv1.Document, r *DocumentReconciler) error {
	defOne := int32(1)
	name := document.Spec.Name
	dep1 := &appsv1.Deployment{
		TypeMeta: matev1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deloyment"},
		ObjectMeta: matev1.ObjectMeta{
			Name:      document.Name,
			Namespace: document.Namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: &defOne,
			Selector: &matev1.LabelSelector{
				MatchLabels: map[string]string{"app": document.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: matev1.ObjectMeta{
					Labels: map[string]string{"app": document.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "document",
							Env: []corev1.EnvVar{
								{Name: "AUTH_NAME", Value: name},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080, Name: "https-document"},
							},
							Image:           document.Spec.Image,
							ImagePullPolicy: "Always",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
									corev1.ResourceMemory: *resource.NewMilliQuantity(500, resource.BinarySI),
								},
							},
						},
					},
				},
			},
		},
	}
	// 这一步非常关键！
	// 建立关联后，删除资源时就会将deployment也删除掉
	if err := controllerutil.SetControllerReference(document, dep1, r.Scheme); err != nil {
		return err
	}

	// 创建deployment
	if err := r.Create(ctx, dep1); err != nil {
		return err
	}
	return nil
}

func updateDocumentDeployment(ctx context.Context, document *uccpsv1.Document, r *DocumentReconciler) error {
	defOne := int32(1)
	name := document.Spec.Name
	dep1 := &appsv1.Deployment{
		TypeMeta: matev1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(), Kind: "Deloyment"},
		ObjectMeta: matev1.ObjectMeta{
			Name:      document.Name,
			Namespace: document.Namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: &defOne,
			Selector: &matev1.LabelSelector{
				MatchLabels: map[string]string{"app": document.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: matev1.ObjectMeta{
					Labels: map[string]string{"app": document.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "document",
							Env: []corev1.EnvVar{
								{Name: "AUTH_NAME", Value: name},
							},
							Ports: []corev1.ContainerPort{
								{ContainerPort: 8080, Name: "http-document"},
							},
							Image:           "harbor.chinauos.com/syx-test/doc-test:1.3",
							ImagePullPolicy: "Always",
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    *resource.NewMilliQuantity(100, resource.DecimalSI),
									corev1.ResourceMemory: *resource.NewMilliQuantity(500, resource.BinarySI),
								},
							},
						},
					},
				},
			},
		},
	}
	// 这一步非常关键！
	// 建立关联后，删除资源时就会将deployment也删除掉
	if err := controllerutil.SetControllerReference(document, dep1, r.Scheme); err != nil {
		return err
	}

	// 创建deployment
	if err := r.Update(ctx, dep1); err != nil {
		return err
	}
	return nil
}

func createService(ctx context.Context, document *uccpsv1.Document, r *DocumentReconciler, req ctrl.Request) error {

	service := &corev1.Service{}

	err := r.Get(ctx, req.NamespacedName, service)

	// 如果查询结果没有错误，证明service正常，就不做任何操作
	if err == nil {
		return nil
	}

	// 如果错误不是NotFound，就返回错误
	if !errors.IsNotFound(err) {
		return err
	}

	svc := &corev1.Service{
		TypeMeta: matev1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String(), Kind: "Service"},
		ObjectMeta: matev1.ObjectMeta{Name: document.Name,
			Namespace: document.Namespace,
		},
		Spec: corev1.ServiceSpec{Ports: []corev1.ServicePort{
			{
				Name:       "https-document",
				Port:       8080,
				Protocol:   "TCP",
				TargetPort: intstr.FromString("https-document"),
			},
		},
			Selector: map[string]string{"app": document.Name},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	// 这一步非常关键！
	// 建立关联后，删除elasticweb资源时就会将deployment也删除掉
	if err := controllerutil.SetControllerReference(document, svc, r.Scheme); err != nil {
		return err
	}

	// 创建service
	if err := r.Create(ctx, svc); err != nil {
		return err
	}
	return nil
}

//创建路由
func createRoute(ctx context.Context, document *uccpsv1.Document, r *DocumentReconciler, req ctrl.Request) error {

	route := &routev1.Route{}

	err := r.Get(ctx, req.NamespacedName, route)

	// 如果查询结果没有错误，证明service正常，就不做任何操作
	if err == nil {
		return nil
	}

	// 如果错误不是NotFound，就返回错误
	if !errors.IsNotFound(err) {
		return err
	}

	var weight int32

	weight = 100
	svcr := &routev1.Route{
		ObjectMeta: matev1.ObjectMeta{
			Namespace: document.Namespace,
		},
		Spec: routev1.RouteSpec{
			Host: "uccps-document",
			Port: &routev1.RoutePort{
				TargetPort: intstr.IntOrString{StrVal: "8080"},
			},
			TLS: &routev1.TLSConfig{
				Termination:                   "edge",
				InsecureEdgeTerminationPolicy: "Redirect",
			},
			WildcardPolicy: routev1.WildcardPolicyNone,

			To: routev1.RouteTargetReference{
				Kind:   "Service",
				Name:   "https-document",
				Weight: &weight,
			},
		},
	}
	// 这一步非常关键！
	// 建立关联后，删除elasticweb资源时就会将deployment也删除掉
	if err := controllerutil.SetControllerReference(document, svcr, r.Scheme); err != nil {
		return err
	}

	// 创建路由
	if err := r.Create(ctx, svcr); err != nil {
		return err
	}
	return nil
}
