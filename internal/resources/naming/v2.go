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

type version2 struct{}

func NewVersion2() Scheme {
	return &version2{}
}

func (v *version2) getResourceLabels(instance, component string) map[string]string {
	return map[string]string{
		appManagedByLabel: "kcp-operator",
		appNameLabel:      "kcp",
		appInstanceLabel:  instance,
		appComponentLabel: component,
	}
}

// RootShard naming

func (v *version2) RootShardDeploymentName(r *operatorv1alpha1.RootShard) string {
	return fmt.Sprintf("%s-root-kcp", r.Name)
}

func (v *version2) RootShardProxyDeploymentName(r *operatorv1alpha1.RootShard) string {
	return fmt.Sprintf("%s-root-proxy", r.Name)
}

func (v *version2) RootShardServiceName(r *operatorv1alpha1.RootShard) string {
	return fmt.Sprintf("%s-root-kcp", r.Name)
}

func (v *version2) RootShardResourceLabels(r *operatorv1alpha1.RootShard) map[string]string {
	return v.getResourceLabels(r.Name, "rootshard")
}

func (v *version2) RootShardProxyResourceLabels(r *operatorv1alpha1.RootShard) map[string]string {
	return v.getResourceLabels(r.Name, "rootshard-proxy")
}

func (v *version2) RootShardBaseHost(r *operatorv1alpha1.RootShard) string {
	return fqService(v.RootShardServiceName(r), r.Namespace, r.Spec.ClusterDomain)
}

func (v *version2) RootShardBaseURL(r *operatorv1alpha1.RootShard) string {
	if r.Spec.ShardBaseURL != "" {
		return r.Spec.ShardBaseURL
	}
	return fmt.Sprintf("https://%s:6443", v.RootShardBaseHost(r))
}

func (v *version2) RootShardCertificateName(r *operatorv1alpha1.RootShard, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-root-%s-cert", r.Name, cert)
}

func (v *version2) RootShardProxyCertificateName(r *operatorv1alpha1.RootShard, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-root-proxy-%s-cert", r.Name, cert)
}

func (v *version2) RootShardCAName(r *operatorv1alpha1.RootShard, ca operatorv1alpha1.CA) string {
	if ca == operatorv1alpha1.RootCA {
		return fmt.Sprintf("%s-root-ca", r.Name)
	}
	return fmt.Sprintf("%s-root-%s-ca", r.Name, ca)
}

func (v *version2) RootShardProxyDynamicKubeconfigName(r *operatorv1alpha1.RootShard) string {
	return fmt.Sprintf("%s-root-proxy-dynamic", r.Name)
}

func (v *version2) RootShardProxyConfigName(r *operatorv1alpha1.RootShard) string {
	return fmt.Sprintf("%s-root-proxy-config", r.Name)
}

func (v *version2) RootShardProxyServiceName(r *operatorv1alpha1.RootShard) string {
	return fmt.Sprintf("%s-root-proxy", r.Name)
}

func (v *version2) RootShardProxyBaseHost(r *operatorv1alpha1.RootShard) string {
	return fqService(v.RootShardProxyServiceName(r), r.Namespace, r.Spec.ClusterDomain)
}

func (v *version2) RootShardKubeconfigSecret(r *operatorv1alpha1.RootShard, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-root-%s", r.Name, cert)
}

// Shard naming

func (v *version2) ShardDeploymentName(s *operatorv1alpha1.Shard) string {
	return fmt.Sprintf("%s-shard-kcp", s.Name)
}

func (v *version2) ShardServiceName(s *operatorv1alpha1.Shard) string {
	return fmt.Sprintf("%s-shard-kcp", s.Name)
}

func (v *version2) ShardResourceLabels(s *operatorv1alpha1.Shard) map[string]string {
	return v.getResourceLabels(s.Name, "shard")
}

func (v *version2) ShardBaseHost(s *operatorv1alpha1.Shard) string {
	return fqService(v.ShardServiceName(s), s.Namespace, s.Spec.ClusterDomain)
}

func (v *version2) ShardBaseURL(s *operatorv1alpha1.Shard) string {
	if s.Spec.ShardBaseURL != "" {
		return s.Spec.ShardBaseURL
	}
	return fmt.Sprintf("https://%s:6443", v.ShardBaseHost(s))
}

func (v *version2) ShardCertificateName(s *operatorv1alpha1.Shard, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-shard-%s-cert", s.Name, cert)
}

func (v *version2) ShardKubeconfigSecret(s *operatorv1alpha1.Shard, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-shard-%s", s.Name, cert)
}

// CacheServer naming

func (v *version2) CacheServerDeploymentName(c *operatorv1alpha1.CacheServer) string {
	return fmt.Sprintf("%s-cache-server", c.Name)
}

func (v *version2) CacheServerServiceName(c *operatorv1alpha1.CacheServer) string {
	return fmt.Sprintf("%s-cache-server", c.Name)
}

func (v *version2) CacheServerResourceLabels(c *operatorv1alpha1.CacheServer) map[string]string {
	return v.getResourceLabels(c.Name, "cache-server")
}

func (v *version2) CacheServerBaseHost(c *operatorv1alpha1.CacheServer) string {
	return fqService(v.CacheServerServiceName(c), c.Namespace, c.Spec.ClusterDomain)
}

func (v *version2) CacheServerBaseURL(c *operatorv1alpha1.CacheServer) string {
	return fmt.Sprintf("https://%s:6443", v.CacheServerBaseHost(c))
}

func (v *version2) CacheServerCertificateName(c *operatorv1alpha1.CacheServer, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-cache-server-%s-cert", c.Name, cert)
}

func (v *version2) CacheServerCAName(cacheServerName string, ca operatorv1alpha1.CA) string {
	if ca == operatorv1alpha1.RootCA {
		return fmt.Sprintf("%s-cache-server-ca", cacheServerName)
	}
	return fmt.Sprintf("%s-cache-server-%s-ca", cacheServerName, ca)
}

func (v *version2) CacheServerClientCertificateName(cacheServerName string) string {
	cs := &operatorv1alpha1.CacheServer{}
	cs.Name = cacheServerName

	return v.CacheServerCertificateName(cs, operatorv1alpha1.ClientCertificate)
}

func (v *version2) CacheServerKubeconfigName(cacheServerName string) string {
	return fmt.Sprintf("%s-cache-server", cacheServerName)
}

// VirtualWorkspace naming

func (v *version2) VirtualWorkspaceDeploymentName(vw *operatorv1alpha1.VirtualWorkspace) string {
	return fmt.Sprintf("%s-virtual-workspace", vw.Name)
}

func (v *version2) VirtualWorkspaceServiceName(vw *operatorv1alpha1.VirtualWorkspace) string {
	return fmt.Sprintf("%s-virtual-workspace", vw.Name)
}

func (v *version2) VirtualWorkspaceResourceLabels(vw *operatorv1alpha1.VirtualWorkspace) map[string]string {
	return v.getResourceLabels(vw.Name, "virtual-workspace")
}

func (v *version2) VirtualWorkspaceBaseHost(vw *operatorv1alpha1.VirtualWorkspace) string {
	return fqService(v.VirtualWorkspaceServiceName(vw), vw.Namespace, vw.Spec.ClusterDomain)
}

func (v *version2) VirtualWorkspaceBaseURL(vw *operatorv1alpha1.VirtualWorkspace) string {
	return fmt.Sprintf("https://%s:6443", v.VirtualWorkspaceBaseHost(vw))
}

func (v *version2) VirtualWorkspaceCertificateName(vw *operatorv1alpha1.VirtualWorkspace, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-virtual-workspace-%s-cert", vw.Name, cert)
}

// FrontProxy naming

func (v *version2) FrontProxyResourceLabels(fp *operatorv1alpha1.FrontProxy) map[string]string {
	return v.getResourceLabels(fp.Name, "front-proxy")
}

func (v *version2) FrontProxyDeploymentName(fp *operatorv1alpha1.FrontProxy) string {
	return fmt.Sprintf("%s-front-proxy", fp.Name)
}

func (v *version2) FrontProxyCertificateName(_ *operatorv1alpha1.RootShard, fp *operatorv1alpha1.FrontProxy, cert operatorv1alpha1.Certificate) string {
	return fmt.Sprintf("%s-front-proxy-%s-cert", fp.Name, cert)
}

func (v *version2) FrontProxyDynamicKubeconfigName(_ *operatorv1alpha1.RootShard, fp *operatorv1alpha1.FrontProxy) string {
	return fmt.Sprintf("%s-front-proxy-dynamic", fp.Name)
}

func (v *version2) FrontProxyConfigName(fp *operatorv1alpha1.FrontProxy) string {
	return fmt.Sprintf("%s-front-proxy-config", fp.Name)
}

func (v *version2) FrontProxyServiceName(fp *operatorv1alpha1.FrontProxy) string {
	return fmt.Sprintf("%s-front-proxy", fp.Name)
}

func (v *version2) FrontProxyBaseHost(fp *operatorv1alpha1.FrontProxy, r *operatorv1alpha1.RootShard) string {
	return fqService(v.FrontProxyServiceName(fp), fp.Namespace, r.Spec.ClusterDomain)
}

// Bundle naming

func (v *version2) BundleName(ownerName string) string {
	return fmt.Sprintf("%s-bundle", ownerName)
}

func (v *version2) MergedCABundleName(ownerName string) string {
	return fmt.Sprintf("%s-merged-ca-bundle", ownerName)
}

func (v *version2) MergedClientCAName(ownerName string) string {
	return fmt.Sprintf("%s-merged-client-ca", ownerName)
}
