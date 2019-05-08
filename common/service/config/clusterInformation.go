// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package config

import "fmt"

// ToClusterInformation convert deprecated ClustersInfo to ClustersInformation
// Deprecated: pleasee use ClustersInformation
func (c ClustersInfo) ToClusterInformation() ClustersInformation {
	clustersInfo := ClustersInformation{}
	clustersInfo.EnableGlobalDomain = c.EnableGlobalDomain
	clustersInfo.VersionIncrement = c.FailoverVersionIncrement
	clustersInfo.MasterClusterName = c.MasterClusterName
	clustersInfo.CurrentClusterName = c.CurrentClusterName
	clustersInfo.ClusterInformation = map[string]ClusterInformation{}
	for k, v := range c.ClusterInitialFailoverVersions {
		address, ok := c.ClusterAddress[k]
		if !ok {
			panic(fmt.Sprintf("unable to find address for cluster %v", k))
		}

		clusterInfo := ClusterInformation{}
		clusterInfo.Enabled = true
		clusterInfo.InitialVersion = v
		clusterInfo.RPCName = address.RPCName
		clusterInfo.RPCAddress = address.RPCAddress
		clustersInfo.ClusterInformation[k] = clusterInfo
	}

	return clustersInfo
}
