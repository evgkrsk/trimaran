/*
Copyright 2021 The Kubernetes Authors.

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

package noderesourcetopology

import (
	"context"
	"reflect"
	"testing"

	"github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/generated/listers/topology/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"

	nrtcache "sigs.k8s.io/scheduler-plugins/pkg/noderesourcetopology/cache"

	topologyv1alpha1 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/apis/topology/v1alpha1"
	faketopologyv1alpha1 "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/generated/clientset/versioned/fake"
	topologyinformers "github.com/k8stopologyawareschedwg/noderesourcetopology-api/pkg/generated/informers/externalversions"
)

type nodeToScoreMap map[string]int64

func initTest(policy topologyv1alpha1.TopologyManagerPolicy) (map[string]*v1.Node, v1alpha1.NodeResourceTopologyLister) {
	nodeTopologies := make([]*topologyv1alpha1.NodeResourceTopology, 3)
	nodesMap := make(map[string]*v1.Node)

	// noderesourcetopology objects
	nodeTopologies[0] = &topologyv1alpha1.NodeResourceTopology{
		ObjectMeta:       metav1.ObjectMeta{Name: "Node1"},
		TopologyPolicies: []string{string(policy)},
		Zones: topologyv1alpha1.ZoneList{
			topologyv1alpha1.Zone{
				Name: "node-0",
				Type: "Node",
				Resources: topologyv1alpha1.ResourceInfoList{
					MakeTopologyResInfo(cpu, "4", "4"),
					MakeTopologyResInfo(memory, "500Mi", "500Mi"),
				},
			},
			topologyv1alpha1.Zone{
				Name: "node-1",
				Type: "Node",
				Resources: topologyv1alpha1.ResourceInfoList{
					MakeTopologyResInfo(cpu, "4", "4"),
					MakeTopologyResInfo(memory, "500Mi", "500Mi"),
				},
			},
		},
	}

	nodeTopologies[1] = &topologyv1alpha1.NodeResourceTopology{
		ObjectMeta:       metav1.ObjectMeta{Name: "Node2"},
		TopologyPolicies: []string{string(policy)},
		Zones: topologyv1alpha1.ZoneList{
			topologyv1alpha1.Zone{
				Name: "node-0",
				Type: "Node",
				Resources: topologyv1alpha1.ResourceInfoList{
					MakeTopologyResInfo(cpu, "2", "2"),
					MakeTopologyResInfo(memory, "50Mi", "50Mi"),
				},
			}, topologyv1alpha1.Zone{
				Name: "node-1",
				Type: "Node",
				Resources: topologyv1alpha1.ResourceInfoList{
					MakeTopologyResInfo(cpu, "2", "2"),
					MakeTopologyResInfo(memory, "50Mi", "50Mi"),
				},
			},
		},
	}
	nodeTopologies[2] = &topologyv1alpha1.NodeResourceTopology{
		ObjectMeta:       metav1.ObjectMeta{Name: "Node3"},
		TopologyPolicies: []string{string(policy)},
		Zones: topologyv1alpha1.ZoneList{
			topologyv1alpha1.Zone{
				Name: "node-0",
				Type: "Node",
				Resources: topologyv1alpha1.ResourceInfoList{
					MakeTopologyResInfo(cpu, "6", "6"),
					MakeTopologyResInfo(memory, "60Mi", "60Mi"),
				},
			}, topologyv1alpha1.Zone{
				Name: "node-1",
				Type: "Node",
				Resources: topologyv1alpha1.ResourceInfoList{
					MakeTopologyResInfo(cpu, "6", "6"),
					MakeTopologyResInfo(memory, "60Mi", "60Mi"),
				},
			},
		},
	}
	// init node objects
	for _, nrt := range nodeTopologies {
		res := makeResourceListFromZones(nrt.Zones)
		nodesMap[nrt.Name] = &v1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: nrt.Name},
			Status: v1.NodeStatus{
				Capacity:    res,
				Allocatable: res,
			},
		}
	}

	// init topology lister
	fakeClient := faketopologyv1alpha1.NewSimpleClientset()
	fakeInformer := topologyinformers.NewSharedInformerFactory(fakeClient, 0).Topology().V1alpha1().NodeResourceTopologies()
	for _, obj := range nodeTopologies {
		fakeInformer.Informer().GetStore().Add(obj)
	}

	return nodesMap, fakeInformer.Lister()
}

func TestNodeResourceScorePlugin(t *testing.T) {

	type podRequests struct {
		pod        *v1.Pod
		name       string
		wantStatus *framework.Status
	}
	pRequests := []podRequests{
		{
			pod: makePodByResourceList(&v1.ResourceList{
				v1.ResourceCPU:    *resource.NewQuantity(2, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(20*1024*1024, resource.DecimalSI)}),
			name:       "Pod1",
			wantStatus: nil,
		},
	}

	// Each testScenario will describe a set pod requests arrived sequentially to the scoring plugin.
	type testScenario struct {
		name      string
		wantedRes nodeToScoreMap
		requests  []podRequests
		strategy  scoreStrategy
	}

	tests := []testScenario{
		{
			// On 0-MaxNodeScore scale
			// Node2 resource fractions:
			// CPU Fraction: 2 / 2 = 100%
			// Memory Fraction: 20M / 50M = 40%
			// Node2 score:(100 + 40) / 2 = 70
			name:      "MostAllocated strategy",
			wantedRes: nodeToScoreMap{"Node2": 70},
			requests:  pRequests,
			strategy:  mostAllocatedScoreStrategy,
		},
		{
			// On 0-MaxNodeScore scale
			// Node3 resource fractions:
			// CPU Fraction: 2 / 6 = 33%
			// Memory Fraction: 20M / 60M = 33%
			// Node3 score: MaxNodeScore - (0.33 - 0.33) = MaxNodeScore
			name:      "BalancedAllocation strategy",
			wantedRes: nodeToScoreMap{"Node3": 100},
			requests:  pRequests,
			strategy:  balancedAllocationScoreStrategy,
		},
		{
			// On 0-MaxNodeScore scale
			// Node1 resource fractions:
			// CPU Fraction: 2 / 4 = 50%
			// Memory Fraction: 20M / 500M = 4%
			// Node1 score: ((100 - 50) + (100 - 4)) / 2 = 73
			name:      "LeastAllocated strategy",
			wantedRes: nodeToScoreMap{"Node1": 73},
			requests:  pRequests,
			strategy:  leastAllocatedScoreStrategy,
		},
	}

	for _, test := range tests {
		nodesMap, lister := initTest(topologyv1alpha1.SingleNUMANodeContainerLevel)
		t.Run(test.name, func(t *testing.T) {
			scoringHandlers := newScoringHandlers(test.strategy, nil)

			tm := &TopologyMatch{
				filterHandlers:  newFilterHandlers(),
				scoringHandlers: scoringHandlers,
				nrtCache:        nrtcache.NewPassthrough(lister),
			}

			for _, req := range test.requests {
				nodeToScore := make(nodeToScoreMap, len(nodesMap))
				for _, node := range nodesMap {
					score, gotStatus := tm.Score(
						context.Background(),
						framework.NewCycleState(),
						req.pod,
						node.ObjectMeta.Name)

					t.Logf("%v; %v; %v; score: %v; status: %v\n",
						test.name,
						req.name,
						node.ObjectMeta.Name,
						score,
						gotStatus)

					if !reflect.DeepEqual(gotStatus, req.wantStatus) {
						t.Errorf("status does not match: %v, want: %v\n", gotStatus, req.wantStatus)
					}
					nodeToScore[node.ObjectMeta.Name] = score
				}
				gotNode := findMaxScoreNode(nodeToScore)
				gotScore := nodeToScore[gotNode]
				t.Logf("%q: got node %q with score %d\n", test.name, gotNode, gotScore)
				for wantNode, wantScore := range test.wantedRes {
					if wantNode != gotNode {
						t.Errorf("failed to select the desired node: wanted: %q, got: %q", wantNode, gotNode)
					}

					if wantScore != gotScore {
						t.Errorf("wrong score for node %q: wanted: %q, got: %d", gotNode, wantScore, gotScore)
					}
				}
			}
		})
	}
}

func TestNodeResourceScorePluginLeastNUMA(t *testing.T) {
	testCases := []struct {
		name        string
		podRequests []v1.ResourceList
		wantedRes   nodeToScoreMap
		policy      topologyv1alpha1.TopologyManagerPolicy
	}{
		{
			name: "container scope, one container case 1",
			podRequests: []v1.ResourceList{
				{
					v1.ResourceCPU:    resource.MustParse("2"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 88,
				"Node2": 88,
				"Node3": 88,
			},
			policy: topologyv1alpha1.BestEffortContainerLevel,
		},
		{
			name: "container scope, one container case 2",
			podRequests: []v1.ResourceList{
				{
					v1.ResourceCPU:    resource.MustParse("4"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 88,
				"Node2": 76,
				"Node3": 88,
			},
			policy: topologyv1alpha1.BestEffortContainerLevel,
		},
		{
			name: "container scope, one container case 3",
			podRequests: []v1.ResourceList{
				{
					v1.ResourceCPU:    resource.MustParse("6"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 76,
				"Node2": 0,
				"Node3": 88,
			},
			policy: topologyv1alpha1.BestEffortContainerLevel,
		},
		{
			name: "container scope, two containers case 1",
			podRequests: []v1.ResourceList{
				{
					v1.ResourceCPU:    resource.MustParse("2"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				{
					v1.ResourceCPU:    resource.MustParse("2"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 88,
				"Node2": 88,
				"Node3": 88,
			},
			policy: topologyv1alpha1.BestEffortContainerLevel,
		},
		{
			name: "container scope, two containers case 2",
			podRequests: []v1.ResourceList{
				{
					v1.ResourceCPU:    resource.MustParse("1"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				{
					v1.ResourceCPU:    resource.MustParse("3"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 88,
				"Node2": 76,
				"Node3": 88,
			},
			policy: topologyv1alpha1.BestEffortContainerLevel,
		},
		{
			name: "container scope, two containers case 3",
			podRequests: []v1.ResourceList{
				{
					v1.ResourceCPU:    resource.MustParse("5"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				{
					v1.ResourceCPU:    resource.MustParse("3"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 76,
				"Node2": 0,
				"Node3": 88,
			},
			policy: topologyv1alpha1.BestEffortContainerLevel,
		},
		{
			name: "container scope, two containers non NUMA resource",
			podRequests: []v1.ResourceList{
				{
					"non-numa-resource": resource.MustParse("1"),
				},
				{
					"non-numa-resource": resource.MustParse("2"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 100,
				"Node2": 100,
				"Node3": 100,
			},
			policy: topologyv1alpha1.BestEffortContainerLevel,
		},
		{
			name: "pod scope, two containers case 1",
			podRequests: []v1.ResourceList{
				{
					v1.ResourceCPU:    resource.MustParse("2"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				{
					v1.ResourceCPU:    resource.MustParse("2"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 88,
				"Node2": 76,
				"Node3": 76,
			},
			policy: topologyv1alpha1.BestEffortPodLevel,
		},
		{
			name: "pod scope, two containers case 2",
			podRequests: []v1.ResourceList{
				{
					v1.ResourceCPU:    resource.MustParse("1"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				{
					v1.ResourceCPU:    resource.MustParse("3"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 88,
				"Node2": 76,
				"Node3": 76,
			},
			policy: topologyv1alpha1.BestEffortPodLevel,
		},
		{
			name: "pod scope, two containers case 3",
			podRequests: []v1.ResourceList{
				{
					v1.ResourceCPU:    resource.MustParse("5"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
				{
					v1.ResourceCPU:    resource.MustParse("3"),
					v1.ResourceMemory: resource.MustParse("50Mi"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 76,
				"Node2": 0,
				"Node3": 76,
			},
			policy: topologyv1alpha1.BestEffortPodLevel,
		},
		{
			name: "pod scope, one containers non NUMA resource",
			podRequests: []v1.ResourceList{
				{
					"non-numa-resource": resource.MustParse("1"),
				},
			},
			wantedRes: nodeToScoreMap{
				"Node1": 100,
				"Node2": 100,
				"Node3": 100,
			},
			policy: topologyv1alpha1.BestEffortPodLevel,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nodesMap, lister := initTest(tc.policy)

			tm := &TopologyMatch{
				filterHandlers:  newFilterHandlers(),
				scoringHandlers: leastNUMAscoreHandlers(),
				nrtCache:        nrtcache.NewPassthrough(lister),
			}
			nodeToScore := make(nodeToScoreMap, len(nodesMap))
			pod := makePodByResourceLists(tc.podRequests...)

			for _, node := range nodesMap {
				score, gotStatus := tm.Score(
					context.Background(),
					framework.NewCycleState(),
					pod,
					node.Name)

				t.Logf("%v; %v; %v; score: %v; status: %v\n",
					tc.name,
					pod.GetName(),
					node.Name,
					score,
					gotStatus)

				nodeToScore[node.Name] = score
			}
			if !reflect.DeepEqual(nodeToScore, tc.wantedRes) {
				t.Errorf("scores for nodes are incorrect wanted: %v, got: %v", tc.wantedRes, nodeToScore)
			}

		})
	}
}

// return the name of the node with the highest score
func findMaxScoreNode(nodeToScore nodeToScoreMap) string {
	max := int64(0)
	electedNode := ""

	for nodeName, score := range nodeToScore {
		if max <= score {
			max = score
			electedNode = nodeName
		}
	}
	return electedNode
}
