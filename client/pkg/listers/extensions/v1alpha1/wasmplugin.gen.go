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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/AtlasCloudTeam/higress/client/pkg/apis/extensions/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// WasmPluginLister helps list WasmPlugins.
// All objects returned here must be treated as read-only.
type WasmPluginLister interface {
	// List lists all WasmPlugins in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.WasmPlugin, err error)
	// WasmPlugins returns an object that can list and get WasmPlugins.
	WasmPlugins(namespace string) WasmPluginNamespaceLister
	WasmPluginListerExpansion
}

// wasmPluginLister implements the WasmPluginLister interface.
type wasmPluginLister struct {
	indexer cache.Indexer
}

// NewWasmPluginLister returns a new WasmPluginLister.
func NewWasmPluginLister(indexer cache.Indexer) WasmPluginLister {
	return &wasmPluginLister{indexer: indexer}
}

// List lists all WasmPlugins in the indexer.
func (s *wasmPluginLister) List(selector labels.Selector) (ret []*v1alpha1.WasmPlugin, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.WasmPlugin))
	})
	return ret, err
}

// WasmPlugins returns an object that can list and get WasmPlugins.
func (s *wasmPluginLister) WasmPlugins(namespace string) WasmPluginNamespaceLister {
	return wasmPluginNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// WasmPluginNamespaceLister helps list and get WasmPlugins.
// All objects returned here must be treated as read-only.
type WasmPluginNamespaceLister interface {
	// List lists all WasmPlugins in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.WasmPlugin, err error)
	// Get retrieves the WasmPlugin from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.WasmPlugin, error)
	WasmPluginNamespaceListerExpansion
}

// wasmPluginNamespaceLister implements the WasmPluginNamespaceLister
// interface.
type wasmPluginNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all WasmPlugins in the indexer for a given namespace.
func (s wasmPluginNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.WasmPlugin, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.WasmPlugin))
	})
	return ret, err
}

// Get retrieves the WasmPlugin from the indexer for a given namespace and name.
func (s wasmPluginNamespaceLister) Get(name string) (*v1alpha1.WasmPlugin, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("wasmplugin"), name)
	}
	return obj.(*v1alpha1.WasmPlugin), nil
}
