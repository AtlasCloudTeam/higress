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

// Code generated by informer-gen. DO NOT EDIT.

package v1

import (
	"context"
	time "time"

	networkingv1 "github.com/AtlasCloudTeam/higress/client/pkg/apis/networking/v1"
	versioned "github.com/AtlasCloudTeam/higress/client/pkg/clientset/versioned"
	internalinterfaces "github.com/AtlasCloudTeam/higress/client/pkg/informers/externalversions/internalinterfaces"
	v1 "github.com/AtlasCloudTeam/higress/client/pkg/listers/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// McpBridgeInformer provides access to a shared informer and lister for
// McpBridges.
type McpBridgeInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1.McpBridgeLister
}

type mcpBridgeInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewMcpBridgeInformer constructs a new informer for McpBridge type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewMcpBridgeInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredMcpBridgeInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredMcpBridgeInformer constructs a new informer for McpBridge type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredMcpBridgeInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkingV1().McpBridges(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.NetworkingV1().McpBridges(namespace).Watch(context.TODO(), options)
			},
		},
		&networkingv1.McpBridge{},
		resyncPeriod,
		indexers,
	)
}

func (f *mcpBridgeInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredMcpBridgeInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *mcpBridgeInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&networkingv1.McpBridge{}, f.defaultInformer)
}

func (f *mcpBridgeInformer) Lister() v1.McpBridgeLister {
	return v1.NewMcpBridgeLister(f.Informer().GetIndexer())
}
