/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package camel

import (
	"github.com/Masterminds/semver"
	"github.com/apache/camel-k/pkg/apis/camel/v1alpha1"
)

// CatalogVersion --
type CatalogVersion struct {
	Version        *semver.Version
	RuntimeVersion *semver.Version
	Catalog        *v1alpha1.CamelCatalog
}

// CatalogVersionCollection --
type CatalogVersionCollection []CatalogVersion

// Len returns the length of a collection. The number of CatalogVersion instances
// on the slice.
func (c CatalogVersionCollection) Len() int {
	return len(c)
}

// Less is needed for the sort interface to compare two CatalogVersion objects on the
// slice. If checks if one is less than the other.
func (c CatalogVersionCollection) Less(i, j int) bool {
	if c[i].Version.LessThan(c[j].Version) {
		return true
	}
	if c[i].Version.GreaterThan(c[j].Version) {
		return false
	}

	return c[i].RuntimeVersion.LessThan(c[j].RuntimeVersion)
}

// Swap is needed for the sort interface to replace the CatalogVersion objects
// at two different positions in the slice.
func (c CatalogVersionCollection) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
