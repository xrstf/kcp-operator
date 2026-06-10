/*
Copyright 2026 The kcp Authors.

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

package migration

import (
	"context"
	"fmt"

	"github.com/kcp-dev/kcp-operator/internal/reconciling"
	k8sreconciling "k8c.io/reconciler/pkg/reconciling"

	certmanagerv1 "github.com/cert-manager/cert-manager/pkg/apis/certmanager/v1"
	corev1 "k8s.io/api/core/v1"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type CertPatchFunc func(*certmanagerv1.Certificate) (*certmanagerv1.Certificate, error)
type CertSecretPatchFunc func(*corev1.Secret, *certmanagerv1.Certificate) (*corev1.Secret, error)

// MirrorCertificate can duplicate an existing cert-manager Certificate under
// a new name, assuming that the Cert's spec is not meant to change (i.e. common
// name, orgs etc. are the same). The goal is to migrate an old version1 certificate
// to a new version2 certificate without having cert-manager re-issue it. This
// is required to keep other certs, especially those embedded into kubeconfigs, work.
//
// Since this function requires the cert to be issued already, it will return an
// error it the cert's Secret does not exist yet.
func MirrorCertificate(
	ctx context.Context,
	client ctrlruntimeclient.Client,
	namespace string, oldCertName string,
	certPatcher CertPatchFunc,
	secretPatcher CertSecretPatchFunc,
) error {
	// get the original cert
	oldCert := &certmanagerv1.Certificate{}
	if err := client.Get(ctx, ctrlruntimeclient.ObjectKey{Namespace: namespace, Name: oldCertName}, oldCert); err != nil {
		return fmt.Errorf("error getting original cert: %w", err)
	}

	// get the cert's secret
	oldSecret := &corev1.Secret{}
	if err := client.Get(ctx, ctrlruntimeclient.ObjectKey{Namespace: namespace, Name: oldCert.Spec.SecretName}, oldSecret); err != nil {
		return fmt.Errorf("error getting original cert secret: %w", err)
	}

	// apply any custom patches to the cert (these must not change anything that's embedded into
	// the cert, but may change the SecretName)
	patchedCert, err := certPatcher(oldCert)
	if err != nil {
		return fmt.Errorf("error patching cert: %w", err)
	}

	// Now patch the secret (this might make use of the changed SecretName or cert name, hence it
	// happens after patching the cert and makes the cert available to the patcher func).
	patchedSecret, err := secretPatcher(oldSecret, patchedCert)
	if err != nil {
		return fmt.Errorf("error patching secret: %w", err)
	}

	// First create the secret, so there is no race condition between us and cert-manager and we
	// avoid one (potentially expensive) cert issuance.
	if err := k8sreconciling.ReconcileSecrets(ctx, []k8sreconciling.NamedSecretReconcilerFactory{
		func() (string, k8sreconciling.SecretReconciler) {
			return patchedSecret.Name, func(s *corev1.Secret) (*corev1.Secret, error) {
				s.Labels = patchedSecret.Labels
				s.Annotations = patchedSecret.Annotations
				s.Data = patchedSecret.Data
				return s, nil
			}
		},
	}, namespace, client); err != nil {
		return fmt.Errorf("error reconciling new cert secret: %w", err)
	}

	if err := reconciling.ReconcileCertificates(ctx, []reconciling.NamedCertificateReconcilerFactory{
		func() (string, reconciling.CertificateReconciler) {
			return patchedCert.Name, func(c *certmanagerv1.Certificate) (*certmanagerv1.Certificate, error) {
				c.Labels = patchedCert.Labels
				c.Annotations = patchedCert.Annotations
				c.Spec = patchedCert.Spec
				return c, nil
			}
		},
	}, namespace, client); err != nil {
		return fmt.Errorf("error reconciling new cert: %w", err)
	}

	return nil
}
