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

package v1

import (
	"context"
	"fmt"
	tensorboard "github.com/kubeflow/kubeflow/components/tensorboard-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// nolint:unused
// log is for logging in this package.
var deploymentlog = logf.Log.WithName("deployment-resource")

// SetupDeploymentWebhookWithManager registers the webhook for Deployment in the manager.
func SetupDeploymentWebhookWithManager(mgr ctrl.Manager) error {
	defaulter := &DeploymentCustomDefaulter{
		Client: mgr.GetClient(),
	}
	return ctrl.NewWebhookManagedBy(mgr).For(&appsv1.Deployment{}).
		WithDefaulter(defaulter).Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-apps-v1-deployment,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps,resources=deployments,verbs=create;update,versions=v1,name=mdeployment-v1.kb.io,admissionReviewVersions=v1

// DeploymentCustomDefaulter struct is responsible for setting default values on the custom resource of the
// Kind Deployment when those are created or updated.
//
// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as it is used only for temporary operations and does not need to be deeply copied.
type DeploymentCustomDefaulter struct {
	Client client.Client
}

var _ webhook.CustomDefaulter = &DeploymentCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the Kind Deployment.
func (d *DeploymentCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	deployment, ok := obj.(*appsv1.Deployment)

	if !ok {
		return fmt.Errorf("expected an Deployment object but got %T", obj)
	}
	deploymentlog.Info("Defaulting for Deployment", "name", deployment.GetName())

	// TODO(user): fill in your defaulting logic.
	//tb, err := d.getTensorboardFromDeployment(ctx, d.Client, deployment)
	//if err != nil {
	//	deploymentlog.Error(err, "fail to get tensorborad")
	//}
	//
	//fmt.Println("===================================>", tb.Annotations)

	toleration := corev1.Toleration{
		Key:      "system/node-group",
		Operator: corev1.TolerationOpEqual,
		Value:    "taint-e625069",
		Effect:   corev1.TaintEffectNoSchedule,
	}

	if deployment.Spec.Template.Spec.Tolerations == nil {
		deployment.Spec.Template.Spec.Tolerations = []corev1.Toleration{}
	}

	deployment.Spec.Template.Spec.Tolerations = append(deployment.Spec.Template.Spec.Tolerations, toleration)
	return nil
}

// getTensorboardFromDeployment
func (d *DeploymentCustomDefaulter) getTensorboardFromDeployment(ctx context.Context, k8sClient client.Client, deployment *appsv1.Deployment) (*tensorboard.Tensorboard, error) {
	for _, ownerRef := range deployment.OwnerReferences {
		if ownerRef.Kind == "Tensorboard" && ownerRef.APIVersion == "kubeflow.org/v1" {
			tensorboard := &tensorboard.Tensorboard{
				TypeMeta:   metav1.TypeMeta{},
				ObjectMeta: metav1.ObjectMeta{},
				Spec:       tensorboard.TensorboardSpec{},
				Status:     tensorboard.TensorboardStatus{},
			}
			err := k8sClient.Get(ctx, client.ObjectKey{
				Namespace: deployment.Namespace,
				Name:      ownerRef.Name,
			}, tensorboard)
			if err != nil {
				return nil, fmt.Errorf("failed to get Tensorboard object: %w", err)
			}
			return tensorboard, nil
		}
	}
	return nil, fmt.Errorf("no Tensorboard owner found for Deployment %s", deployment.Name)
}
