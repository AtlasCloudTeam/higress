// Copyright (c) 2022 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"
	json "encoding/json"
	"fmt"

	v1alpha1 "github.com/AtlasCloudTeam/higress/client/pkg/apis/extensions/v1alpha1"
	extensionsv1alpha1 "github.com/AtlasCloudTeam/higress/client/pkg/applyconfiguration/extensions/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeWasmPlugins implements WasmPluginInterface
type FakeWasmPlugins struct {
	Fake *FakeExtensionsV1alpha1
	ns   string
}

var wasmpluginsResource = v1alpha1.SchemeGroupVersion.WithResource("wasmplugins")

var wasmpluginsKind = v1alpha1.SchemeGroupVersion.WithKind("WasmPlugin")

// Get takes name of the wasmPlugin, and returns the corresponding wasmPlugin object, and an error if there is any.
func (c *FakeWasmPlugins) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.WasmPlugin, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(wasmpluginsResource, c.ns, name), &v1alpha1.WasmPlugin{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WasmPlugin), err
}

// List takes label and field selectors, and returns the list of WasmPlugins that match those selectors.
func (c *FakeWasmPlugins) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.WasmPluginList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(wasmpluginsResource, wasmpluginsKind, c.ns, opts), &v1alpha1.WasmPluginList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.WasmPluginList{ListMeta: obj.(*v1alpha1.WasmPluginList).ListMeta}
	for _, item := range obj.(*v1alpha1.WasmPluginList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested wasmPlugins.
func (c *FakeWasmPlugins) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(wasmpluginsResource, c.ns, opts))

}

// Create takes the representation of a wasmPlugin and creates it.  Returns the server's representation of the wasmPlugin, and an error, if there is any.
func (c *FakeWasmPlugins) Create(ctx context.Context, wasmPlugin *v1alpha1.WasmPlugin, opts v1.CreateOptions) (result *v1alpha1.WasmPlugin, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(wasmpluginsResource, c.ns, wasmPlugin), &v1alpha1.WasmPlugin{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WasmPlugin), err
}

// Update takes the representation of a wasmPlugin and updates it. Returns the server's representation of the wasmPlugin, and an error, if there is any.
func (c *FakeWasmPlugins) Update(ctx context.Context, wasmPlugin *v1alpha1.WasmPlugin, opts v1.UpdateOptions) (result *v1alpha1.WasmPlugin, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(wasmpluginsResource, c.ns, wasmPlugin), &v1alpha1.WasmPlugin{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WasmPlugin), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeWasmPlugins) UpdateStatus(ctx context.Context, wasmPlugin *v1alpha1.WasmPlugin, opts v1.UpdateOptions) (*v1alpha1.WasmPlugin, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(wasmpluginsResource, "status", c.ns, wasmPlugin), &v1alpha1.WasmPlugin{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WasmPlugin), err
}

// Delete takes name of the wasmPlugin and deletes it. Returns an error if one occurs.
func (c *FakeWasmPlugins) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(wasmpluginsResource, c.ns, name, opts), &v1alpha1.WasmPlugin{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeWasmPlugins) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(wasmpluginsResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.WasmPluginList{})
	return err
}

// Patch applies the patch and returns the patched wasmPlugin.
func (c *FakeWasmPlugins) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.WasmPlugin, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(wasmpluginsResource, c.ns, name, pt, data, subresources...), &v1alpha1.WasmPlugin{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WasmPlugin), err
}

// Apply takes the given apply declarative configuration, applies it and returns the applied wasmPlugin.
func (c *FakeWasmPlugins) Apply(ctx context.Context, wasmPlugin *extensionsv1alpha1.WasmPluginApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.WasmPlugin, err error) {
	if wasmPlugin == nil {
		return nil, fmt.Errorf("wasmPlugin provided to Apply must not be nil")
	}
	data, err := json.Marshal(wasmPlugin)
	if err != nil {
		return nil, err
	}
	name := wasmPlugin.Name
	if name == nil {
		return nil, fmt.Errorf("wasmPlugin.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(wasmpluginsResource, c.ns, *name, types.ApplyPatchType, data), &v1alpha1.WasmPlugin{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WasmPlugin), err
}

// ApplyStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating ApplyStatus().
func (c *FakeWasmPlugins) ApplyStatus(ctx context.Context, wasmPlugin *extensionsv1alpha1.WasmPluginApplyConfiguration, opts v1.ApplyOptions) (result *v1alpha1.WasmPlugin, err error) {
	if wasmPlugin == nil {
		return nil, fmt.Errorf("wasmPlugin provided to Apply must not be nil")
	}
	data, err := json.Marshal(wasmPlugin)
	if err != nil {
		return nil, err
	}
	name := wasmPlugin.Name
	if name == nil {
		return nil, fmt.Errorf("wasmPlugin.Name must be provided to Apply")
	}
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(wasmpluginsResource, c.ns, *name, types.ApplyPatchType, data, "status"), &v1alpha1.WasmPlugin{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WasmPlugin), err
}
