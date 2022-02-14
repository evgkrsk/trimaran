/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1alpha1 "sigs.k8s.io/scheduler-plugins/apis/scheduling/v1alpha1"
)

// FakePodGroups implements PodGroupInterface
type FakePodGroups struct {
	Fake *FakeSchedulingV1alpha1
	ns   string
}

var podgroupsResource = schema.GroupVersionResource{Group: "scheduling.sigs.k8s.io", Version: "v1alpha1", Resource: "podgroups"}

var podgroupsKind = schema.GroupVersionKind{Group: "scheduling.sigs.k8s.io", Version: "v1alpha1", Kind: "PodGroup"}

// Get takes name of the podGroup, and returns the corresponding podGroup object, and an error if there is any.
func (c *FakePodGroups) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PodGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(podgroupsResource, c.ns, name), &v1alpha1.PodGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodGroup), err
}

// List takes label and field selectors, and returns the list of PodGroups that match those selectors.
func (c *FakePodGroups) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PodGroupList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(podgroupsResource, podgroupsKind, c.ns, opts), &v1alpha1.PodGroupList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PodGroupList{ListMeta: obj.(*v1alpha1.PodGroupList).ListMeta}
	for _, item := range obj.(*v1alpha1.PodGroupList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested podGroups.
func (c *FakePodGroups) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(podgroupsResource, c.ns, opts))

}

// Create takes the representation of a podGroup and creates it.  Returns the server's representation of the podGroup, and an error, if there is any.
func (c *FakePodGroups) Create(ctx context.Context, podGroup *v1alpha1.PodGroup, opts v1.CreateOptions) (result *v1alpha1.PodGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(podgroupsResource, c.ns, podGroup), &v1alpha1.PodGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodGroup), err
}

// Update takes the representation of a podGroup and updates it. Returns the server's representation of the podGroup, and an error, if there is any.
func (c *FakePodGroups) Update(ctx context.Context, podGroup *v1alpha1.PodGroup, opts v1.UpdateOptions) (result *v1alpha1.PodGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(podgroupsResource, c.ns, podGroup), &v1alpha1.PodGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodGroup), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePodGroups) UpdateStatus(ctx context.Context, podGroup *v1alpha1.PodGroup, opts v1.UpdateOptions) (*v1alpha1.PodGroup, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(podgroupsResource, "status", c.ns, podGroup), &v1alpha1.PodGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodGroup), err
}

// Delete takes name of the podGroup and deletes it. Returns an error if one occurs.
func (c *FakePodGroups) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(podgroupsResource, c.ns, name, opts), &v1alpha1.PodGroup{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePodGroups) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(podgroupsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.PodGroupList{})
	return err
}

// Patch applies the patch and returns the patched podGroup.
func (c *FakePodGroups) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PodGroup, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podgroupsResource, c.ns, name, pt, data, subresources...), &v1alpha1.PodGroup{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodGroup), err
}
