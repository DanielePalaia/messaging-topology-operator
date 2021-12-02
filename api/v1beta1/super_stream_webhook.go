package v1beta1

import (
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (s *SuperStream) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(s).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-rabbitmq-com-v1beta1-superstream,mutating=false,failurePolicy=fail,groups=rabbitmq.com,resources=superstreams,versions=v1beta1,name=vsuperstream.kb.io,sideEffects=none,admissionReviewVersions=v1

var _ webhook.Validator = &SuperStream{}

// no validation on create
func (s *SuperStream) ValidateCreate() error {
	return nil
}

// returns error type 'forbidden' for updates on superstream name and rabbitmqClusterReference
func (s *SuperStream) ValidateUpdate(old runtime.Object) error {
	oldSuperStream, ok := old.(*SuperStream)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a superstream but got a %T", old))
	}

	detailMsg := "updates on name, vhost and rabbitmqClusterReference are all forbidden"
	if s.Spec.Name != oldSuperStream.Spec.Name {
		return apierrors.NewForbidden(s.GroupResource(), s.Name,
			field.Forbidden(field.NewPath("spec", "name"), detailMsg))
	}
	if s.Spec.Vhost != oldSuperStream.Spec.Vhost {
		return apierrors.NewForbidden(s.GroupResource(), s.Name,
			field.Forbidden(field.NewPath("spec", "vhost"), detailMsg))
	}

	if s.Spec.RabbitmqClusterReference != oldSuperStream.Spec.RabbitmqClusterReference {
		return apierrors.NewForbidden(s.GroupResource(), s.Name,
			field.Forbidden(field.NewPath("spec", "rabbitmqClusterReference"), detailMsg))
	}

	if !routingKeyUpdatePermitted(oldSuperStream.Spec.RoutingKeys, s.Spec.RoutingKeys) {
		return apierrors.NewForbidden(s.GroupResource(), s.Name,
			field.Forbidden(field.NewPath("spec", "routingKeys"), "updates may only add to the existing list of routing keys"))
	}

	if s.Spec.Partitions < oldSuperStream.Spec.Partitions {
		return apierrors.NewForbidden(s.GroupResource(), s.Name,
			field.Forbidden(field.NewPath("spec", "partitions"), "updates may only increase the partition count, and may not decrease it"))
	}

	return nil
}

// ValidateDelete no validation on delete
func (s *SuperStream) ValidateDelete() error {
	return nil
}

// routingKeyUpdatePermitted allows updates only if adding additional keys at the end of the list of keys
func routingKeyUpdatePermitted(old, new []string) bool {
	if len(old) == 0 && len(new) != 0 {
		return false
	}
	for i := 0; i < len(old); i++ {
		if old[i] != new[i] {
			return false
		}
	}
	return true
}
