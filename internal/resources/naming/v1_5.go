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

package naming

import (
	"fmt"

	operatorv1alpha1 "github.com/kcp-dev/kcp-operator/sdk/apis/operator/v1alpha1"
)

// version1_5 is a minimal change from v1 that only fixes actual naming conflicts
// while keeping RootShard names completely unchanged for easier migration.
//
// Changes from v1:
// - Shard certificates get "-shard-" infix to avoid conflict with RootShard
// - Shard kubeconfigs get "-shard-" infix to avoid conflict with RootShard
// - CacheServer certificates/CAs get "-cache-server-" infix to avoid conflicts
// - VirtualWorkspace certificates get "-virtual-workspace-" infix to avoid conflicts
// - FrontProxy certificates drop RootShard dependency (breaking conflict potential)
//
// UNCHANGED (zero migration effort):
// - All RootShard resources keep exact v1 names
// - All Shard deployments/services keep exact v1 names
// - All CacheServer/VirtualWorkspace/FrontProxy deployments/services keep exact v1 names
// - Bundle names keep exact v1 names
//
// NOTE: This scheme still has naming conflicts! See MIGRATION.md for details.
// A RootShard named "X-shard" will conflict with a Shard named "X".

type version1_5 struct {
	version1
}

func NewVersion1_5() Scheme {
	return &version1_5{
		version1: version1{},
	}
}

var _ Scheme = &version1_5{}

// Shard naming - deployments/services unchanged, secrets/certs get -shard- infix

func (v *version1_5) ShardDeploymentName(s *operatorv1alpha1.Shard) string {
	return fmt.Sprintf("%s-shard-kcp", s.Name)
}

func (v *version1_5) ShardServiceName(s *operatorv1alpha1.Shard) string {
	return fmt.Sprintf("%s-shard-kcp", s.Name)
}

func (v *version1_5) ShardResourceLabels(s *operatorv1alpha1.Shard) map[string]string {
	return v.getResourceLabels(s.Name, "shard")
}

func (v *version1_5) ShardBaseHost(s *operatorv1alpha1.Shard) string {
	return fqService(v.ShardServiceName(s), s.Namespace, s.Spec.ClusterDomain)
}

func (v *version1_5) ShardBaseURL(s *operatorv1alpha1.Shard) string {
	if s.Spec.ShardBaseURL != "" {
		return s.Spec.ShardBaseURL
	}
	return fmt.Sprintf("https://%s:6443", v.ShardBaseHost(s))
}

// CHANGED: Add -shard- infix to avoid conflict with RootShard certificates
func (v *version1_5) ShardCertificateName(s *operatorv1alpha1.Shard, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-shard-%s", s.Name, cert)
}

// CHANGED: Add -shard- infix to avoid conflict with RootShard kubeconfigs
func (v *version1_5) ShardKubeconfigSecret(s *operatorv1alpha1.Shard, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-shard-%s-kubeconfig", s.Name, cert)
}

// CacheServer naming - deployments/services unchanged, secrets/certs/CAs get -cache-server- infix

func (v *version1_5) CacheServerDeploymentName(c *operatorv1alpha1.CacheServer) string {
	return fmt.Sprintf("%s-cache-server", c.Name)
}

func (v *version1_5) CacheServerServiceName(c *operatorv1alpha1.CacheServer) string {
	return fmt.Sprintf("%s-cache-server", c.Name)
}

func (v *version1_5) CacheServerResourceLabels(c *operatorv1alpha1.CacheServer) map[string]string {
	return v.getResourceLabels(c.Name, "cache-server")
}

func (v *version1_5) CacheServerBaseHost(c *operatorv1alpha1.CacheServer) string {
	return fqService(v.CacheServerServiceName(c), c.Namespace, c.Spec.ClusterDomain)
}

func (v *version1_5) CacheServerBaseURL(c *operatorv1alpha1.CacheServer) string {
	return fmt.Sprintf("https://%s:6443", v.CacheServerBaseHost(c))
}

// CHANGED: Add -cache-server- infix to avoid conflict with RootShard certificates
func (v *version1_5) CacheServerCertificateName(c *operatorv1alpha1.CacheServer, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-cache-server-%s", c.Name, cert)
}

// CHANGED: Add -cache-server- infix to avoid conflict with RootShard CAs
func (v *version1_5) CacheServerCAName(cacheServerName string, ca operatorv1alpha1.CA) string {
	if ca == operatorv1alpha1.RootCA {
		return fmt.Sprintf("%s-cache-server-ca", cacheServerName)
	}
	return fmt.Sprintf("%s-cache-server-%s-ca", cacheServerName, ca)
}

// CHANGED: Now uses CacheServerCertificateName for consistency
func (v *version1_5) CacheServerClientCertificateName(cacheServerName string) string {
	cs := &operatorv1alpha1.CacheServer{}
	cs.Name = cacheServerName

	return v.CacheServerCertificateName(cs, operatorv1alpha1.ClientCertificate)
}

func (v *version1_5) CacheServerKubeconfigName(cacheServerName string) string {
	return fmt.Sprintf("%s-kubeconfig", cacheServerName)
}

// VirtualWorkspace naming - deployments/services unchanged, certs get -virtual-workspace- infix

func (v *version1_5) VirtualWorkspaceDeploymentName(vw *operatorv1alpha1.VirtualWorkspace) string {
	return fmt.Sprintf("%s-virtual-workspace", vw.Name)
}

func (v *version1_5) VirtualWorkspaceServiceName(vw *operatorv1alpha1.VirtualWorkspace) string {
	return fmt.Sprintf("%s-virtual-workspace", vw.Name)
}

func (v *version1_5) VirtualWorkspaceResourceLabels(vw *operatorv1alpha1.VirtualWorkspace) map[string]string {
	return v.getResourceLabels(vw.Name, "virtual-workspace")
}

func (v *version1_5) VirtualWorkspaceBaseHost(vw *operatorv1alpha1.VirtualWorkspace) string {
	return fqService(v.VirtualWorkspaceServiceName(vw), vw.Namespace, vw.Spec.ClusterDomain)
}

func (v *version1_5) VirtualWorkspaceBaseURL(vw *operatorv1alpha1.VirtualWorkspace) string {
	return fmt.Sprintf("https://%s:6443", v.VirtualWorkspaceBaseHost(vw))
}

// CHANGED: Add -virtual-workspace- infix to avoid potential conflicts
func (v *version1_5) VirtualWorkspaceCertificateName(vw *operatorv1alpha1.VirtualWorkspace, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-virtual-workspace-%s", vw.Name, cert)
}

// FrontProxy naming - deployments/services unchanged, certs drop RootShard dependency

func (v *version1_5) FrontProxyResourceLabels(fp *operatorv1alpha1.FrontProxy) map[string]string {
	return v.getResourceLabels(fp.Name, "front-proxy")
}

func (v *version1_5) FrontProxyDeploymentName(fp *operatorv1alpha1.FrontProxy) string {
	return fmt.Sprintf("%s-front-proxy", fp.Name)
}

// CHANGED: Remove RootShard name prefix to avoid conflicts like "root-fp-cert" vs RootShard "root-fp"
func (v *version1_5) FrontProxyCertificateName(_ *operatorv1alpha1.RootShard, fp *operatorv1alpha1.FrontProxy, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-front-proxy-%s", fp.Name, cert)
}

// CHANGED: Remove RootShard name prefix
func (v *version1_5) FrontProxyDynamicKubeconfigName(_ *operatorv1alpha1.RootShard, fp *operatorv1alpha1.FrontProxy) string {
	return fmt.Sprintf("%s-front-proxy-dynamic-kubeconfig", fp.Name)
}

func (v *version1_5) FrontProxyConfigName(fp *operatorv1alpha1.FrontProxy) string {
	return fmt.Sprintf("%s-config", fp.Name)
}

func (v *version1_5) FrontProxyServiceName(fp *operatorv1alpha1.FrontProxy) string {
	return fmt.Sprintf("%s-front-proxy", fp.Name)
}

func (v *version1_5) FrontProxyBaseHost(fp *operatorv1alpha1.FrontProxy, r *operatorv1alpha1.RootShard) string {
	return fqService(v.FrontProxyServiceName(fp), fp.Namespace, r.Spec.ClusterDomain)
}

// Bundle naming - ALL UNCHANGED FROM V1

func (v *version1_5) BundleName(ownerName string) string {
	return fmt.Sprintf("%s-bundle", ownerName)
}

func (v *version1_5) MergedCABundleName(ownerName string) string {
	return fmt.Sprintf("%s-merged-ca-bundle", ownerName)
}

func (v *version1_5) MergedClientCAName(ownerName string) string {
	return fmt.Sprintf("%s-merged-client-ca", ownerName)
}
