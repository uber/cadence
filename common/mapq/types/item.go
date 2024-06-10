// The MIT License (MIT)

// Copyright (c) 2017-2020 Uber Technologies Inc.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package types

import "fmt"

type Item interface {
	// GetAttribute returns the value of the attribute key.
	// Value can be any type. It's up to the client to interpret the value.
	// Attribute keys used as partition keys will be converted to string because they are used as node
	// identifiers in the queue tree.
	GetAttribute(key string) any

	// Offset returns the offset of the item in the queue. e.g. monotonically increasing sequence number or a timestamp
	Offset() int64

	// String returns a human friendly representation of the item for logging purposes
	String() string
}

type ItemToPersist interface {
	Item
	ItemPartitions
}

func NewItemToPersist(item Item, itemPartitions ItemPartitions) ItemToPersist {
	return &defaultItemToPersist{
		item:           item,
		itemPartitions: itemPartitions,
	}
}

type ItemPartitions interface {
	// GetPartitions returns the partition keys ordered by their level in the tree
	GetPartitions() []string

	// GetPartitionValue returns the partition value to determine the target queue
	// e.g.
	//  Below example demonstrates that item is in a catch-all queue for sub-type
	//  	Item.GetAttribute("sub-type") returns "timer"
	//  	ItemPartitions.GetPartitionValue("sub-type") returns "*"
	//
	GetPartitionValue(key string) any
}

func NewItemPartitions(partitions []string, partitionMap map[string]any) ItemPartitions {
	return &defaultItemPartitions{
		partitions:   partitions,
		partitionMap: partitionMap,
	}
}

type defaultItemPartitions struct {
	partitions   []string
	partitionMap map[string]any
}

func (i *defaultItemPartitions) GetPartitions() []string {
	return i.partitions
}

func (i *defaultItemPartitions) GetPartitionValue(key string) any {
	return i.partitionMap[key]
}

func (i *defaultItemPartitions) String() string {
	return fmt.Sprintf("ItemPartitions{partitions:%v, partitionMap:%v}", i.partitions, i.partitionMap)
}

type defaultItemToPersist struct {
	item           Item
	itemPartitions ItemPartitions
}

func (i *defaultItemToPersist) String() string {
	return fmt.Sprintf("ItemToPersist{item:%v, itemPartitions:%v}", i.item, i.itemPartitions)
}

func (i *defaultItemToPersist) Offset() int64 {
	return i.item.Offset()
}

func (i *defaultItemToPersist) GetAttribute(key string) any {
	return i.item.GetAttribute(key)
}

func (i *defaultItemToPersist) GetPartitions() []string {
	return i.itemPartitions.GetPartitions()
}

func (i *defaultItemToPersist) GetPartitionValue(key string) any {
	return i.itemPartitions.GetPartitionValue(key)
}
