/*
Copyright 2026 The KCP Authors.

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

package kcp

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlruntime "sigs.k8s.io/controller-runtime"
	ctrlruntimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	operatorv1alpha1 "github.com/kcp-dev/kcp-operator/sdk/apis/operator/v1alpha1"
	"github.com/kcp-dev/kcp-operator/test/utils"
)

// lowResourceRequirements returns minimal resource requirements for testing.
func lowResourceRequirements() *corev1.ResourceRequirements {
	return &corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("10m"),
			corev1.ResourceMemory: resource.MustParse("32Mi"),
		},
	}
}

// waitForVirtualWorkspacePods waits for VirtualWorkspace pods to become ready.
func waitForVirtualWorkspacePods(t *testing.T, ctx context.Context, client ctrlruntimeclient.Client, namespace, name string) {
	t.Helper()

	opts := []ctrlruntimeclient.ListOption{
		ctrlruntimeclient.InNamespace(namespace),
		ctrlruntimeclient.MatchingLabels{
			"app.kubernetes.io/component": "virtual-workspace",
			"app.kubernetes.io/instance":  name,
		},
	}
	utils.WaitForPods(t, ctx, client, opts...)
}

func TestKcpTestSuite(t *testing.T) {
	testImage := os.Getenv("KCP_E2E_TEST_IMAGE")
	if testImage == "" {
		t.Skip("No $KCP_E2E_TEST_IMAGE defined.")
	}

	ctrlruntime.SetLogger(logr.Discard())

	client := utils.GetKubeClient(t)
	ctx := context.Background()

	// create namspace
	namespace := utils.CreateSelfDestructingNamespace(t, ctx, client, "kcp")

	// deploy the cache server
	cacheServer := utils.DeployCacheServer(ctx, t, client, namespace.Name)

	// Deploy external VirtualWorkspace for root shard
	rootVW := utils.DeployVirtualWorkspace(ctx, t, client, namespace.Name, "root-vw", false, func(vw *operatorv1alpha1.VirtualWorkspace) {
		vw.Spec.Target.RootShardRef = &corev1.LocalObjectReference{
			Name: "r00t",
		}
		vw.Spec.Resources = lowResourceRequirements()
	})

	// externalHostname must match whatever DeployFrontProxy chooses as the name for the FrontProxy
	externalHostname := fmt.Sprintf("front-proxy-front-proxy.%s.svc.cluster.local", namespace.Name)

	// Deploy root shard with external VW
	rootShard := utils.DeployRootShard(ctx, t, client, namespace.Name, externalHostname, func(rs *operatorv1alpha1.RootShard) {
		rs.Spec.KCPVirtualWorkspace = &corev1.LocalObjectReference{
			Name: rootVW.Name,
		}
		rs.Spec.Cache.Reference = &corev1.LocalObjectReference{
			Name: cacheServer.Name,
		}
		rs.Spec.Resources = lowResourceRequirements()
		if rs.Spec.Proxy != nil {
			rs.Spec.Proxy.Resources = lowResourceRequirements()
		}
	})

	// Deploy external VirtualWorkspace for regular shard1
	shard1VW := utils.DeployVirtualWorkspace(ctx, t, client, namespace.Name, "shard1-vw", false, func(vw *operatorv1alpha1.VirtualWorkspace) {
		vw.Spec.Target.ShardRef = &corev1.LocalObjectReference{
			Name: "shard1",
		}
		vw.Spec.Resources = lowResourceRequirements()
	})

	// Deploy regular shard1 with external VW
	utils.DeployShard(ctx, t, client, namespace.Name, "shard1", rootShard.Name, func(s *operatorv1alpha1.Shard) {
		s.Spec.KCPVirtualWorkspace = &corev1.LocalObjectReference{
			Name: shard1VW.Name,
		}
		s.Spec.Resources = lowResourceRequirements()
	})

	// Deploy external VirtualWorkspace for regular shard2
	// shard2VW := utils.DeployVirtualWorkspace(ctx, t, client, namespace.Name, "shard2-vw", false, func(vw *operatorv1alpha1.VirtualWorkspace) {
	// 	vw.Spec.Target.ShardRef = &corev1.LocalObjectReference{
	// 		Name: "shard2",
	// 	}
	// 	vw.Spec.Resources = lowResourceRequirements()
	// })

	// Deploy another regular shard with internal VW
	// utils.DeployShard(ctx, t, client, namespace.Name, "shard2", rootShard.Name, func(s *operatorv1alpha1.Shard) {
	// 	s.Spec.KCPVirtualWorkspace = &corev1.LocalObjectReference{
	// 		Name: shard2VW.Name,
	// 	}
	// 	s.Spec.Resources = lowResourceRequirements()
	// })

	// Deploy front-proxy
	frontProxy := utils.DeployFrontProxy(ctx, t, client, namespace.Name, rootShard.Name, externalHostname, func(fp *operatorv1alpha1.FrontProxy) {
		fp.Spec.Resources = lowResourceRequirements()
	})

	// Wait for both VirtualWorkspace pods to be ready
	t.Log("Waiting for root VirtualWorkspace pods to be ready...")
	waitForVirtualWorkspacePods(t, ctx, client, rootVW.Namespace, rootVW.Name)

	t.Log("Waiting for shard1 VirtualWorkspace pods to be ready...")
	waitForVirtualWorkspacePods(t, ctx, client, shard1VW.Namespace, shard1VW.Name)

	// t.Log("Waiting for shard2 VirtualWorkspace pods to be ready...")
	// waitForVirtualWorkspacePods(t, ctx, client, shard2VW.Namespace, shard2VW.Name)

	///////////////////////////////////////////////////////////////////////////////////

	kubeconfigMap := map[string]string{}

	// create a kubeconfig to access the front proxy
	fpKubeconfig := "e2e-fp-kubeconfig"
	createKubeconfig(t, ctx, client, namespace.Name, fpKubeconfig, operatorv1alpha1.KubeconfigTarget{
		FrontProxyRef: &corev1.LocalObjectReference{
			Name: frontProxy.Name,
		},
	})
	kubeconfigMap["KCP_KUBECONFIG"] = fpKubeconfig

	// create a kubeconfig to access the root shard
	rootKubeconfig := "e2e-root-kubeconfig"
	createKubeconfig(t, ctx, client, namespace.Name, rootKubeconfig, operatorv1alpha1.KubeconfigTarget{
		RootShardRef: &corev1.LocalObjectReference{
			Name: rootShard.Name,
		},
	})
	kubeconfigMap["KCP_SHARD_ROOT_KUBECONFIG"] = rootKubeconfig
	shardList := []string{"root"}

	// create one kubeconfig per shard, each one must have a "shard-base" context in it (as per the
	// kcp e2e test code)
	for _, shard := range []string{"shard1"} { //, "shard2"} {
		shardKubeconfig := fmt.Sprintf("e2e-shard-%s-kubeconfig", shard)
		createKubeconfig(t, ctx, client, namespace.Name, shardKubeconfig, operatorv1alpha1.KubeconfigTarget{
			ShardRef: &corev1.LocalObjectReference{
				Name: shard,
			},
		})

		key := fmt.Sprintf("KCP_SHARD_%s_KUBECONFIG", strings.ToUpper(shard))
		kubeconfigMap[key] = shardKubeconfig
		shardList = append(shardList, shard)
	}

	// deploy kcp e2e test container into the cluster
	volumes := []corev1.Volume{}
	container := corev1.Container{
		Name:            "e2e",
		Image:           testImage,
		ImagePullPolicy: corev1.PullNever,
		Env: []corev1.EnvVar{{
			Name:  "KCP_SHARDS",
			Value: strings.Join(shardList, ","),
		}},
	}

	for key, secretName := range kubeconfigMap {
		fsKey := strings.ReplaceAll(strings.ToLower(key), "_", "-")
		mountPath := fmt.Sprintf("/opt/kubeconfigs/%s", fsKey)
		volumeName := fmt.Sprintf("%s-volume", fsKey)

		volumes = append(volumes, corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: secretName,
				},
			},
		})
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  key,
			Value: fmt.Sprintf("%s/kubeconfig", mountPath),
		})
		container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
			Name:      volumeName,
			ReadOnly:  true,
			MountPath: mountPath,
		})
	}

	testPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    namespace.Name,
			GenerateName: "kcp-e2e-",
			Labels: map[string]string{
				"test": "kcp-e2e",
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers:    []corev1.Container{container},
			Volumes:       volumes,
		},
	}

	t.Log("Creating kcp e2e test pod…")
	if err := client.Create(ctx, testPod); err != nil {
		t.Fatal(err)
	}

	t.Log("Sleeping for 60 minutes…")
	time.Sleep(60 * time.Minute)
}

func createKubeconfig(t *testing.T, ctx context.Context, client ctrlruntimeclient.Client, namespace string, name string, target operatorv1alpha1.KubeconfigTarget) {
	kubeconfig := operatorv1alpha1.Kubeconfig{}
	kubeconfig.Name = name
	kubeconfig.Namespace = namespace

	kubeconfig.Spec = operatorv1alpha1.KubeconfigSpec{
		Target:   target,
		Username: "e2e",
		Validity: metav1.Duration{Duration: 2 * time.Hour},
		SecretRef: corev1.LocalObjectReference{
			Name: name,
		},
		Groups: []string{"system:kcp:admin"},
		// Authorization: &operatorv1alpha1.KubeconfigAuthorization{
		// 	ClusterRoleBindings: operatorv1alpha1.KubeconfigClusterRoleBindings{
		// 		Cluster:      "root",
		// 		ClusterRoles: []string{"cluster-admin"},
		// 	},
		// },
	}

	t.Logf("Creating %s kubeconfig…", kubeconfig.Name)
	if err := client.Create(ctx, &kubeconfig); err != nil {
		t.Fatal(err)
	}
	utils.WaitForObject(t, ctx, client, &corev1.Secret{}, types.NamespacedName{Namespace: kubeconfig.Namespace, Name: kubeconfig.Spec.SecretRef.Name})
}
