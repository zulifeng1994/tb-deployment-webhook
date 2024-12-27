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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const tolerationKey = "system/node-group"

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

	tb, err := d.getTensorboardFromDeployment(ctx, d.Client, deployment)
	if err != nil {
		deploymentlog.Error(err, "fail to get tensorboard")
	}
	if tb == nil {
		return nil
	}

	var toleration corev1.Toleration
	nodeGroup, ok := tb.Annotations[tolerationKey]
	if ok {
		toleration = corev1.Toleration{
			Key:      tolerationKey,
			Operator: corev1.TolerationOpEqual,
			Value:    nodeGroup,
			Effect:   corev1.TaintEffectNoSchedule,
		}
		selector := map[string]string{
			tolerationKey: nodeGroup,
		}
		deployment.Spec.Template.Spec.NodeSelector = selector
	} else {
		toleration = corev1.Toleration{
			Key:      tolerationKey,
			Operator: corev1.TolerationOpExists,
		}
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
		if ownerRef.Kind == "Tensorboard" && ownerRef.APIVersion == "tensorboard.kubeflow.org/v1alpha1" {
			tb := &tensorboard.Tensorboard{}
			err := k8sClient.Get(ctx, client.ObjectKey{
				Namespace: deployment.Namespace,
				Name:      ownerRef.Name,
			}, tb)
			if err != nil {
				return nil, fmt.Errorf("failed to get Tensorboard object: %w", err)
			}
			return tb, nil
		}
	}
	return nil, nil
}
