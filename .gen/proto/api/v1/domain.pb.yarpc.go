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

// Code generated by protoc-gen-yarpc-go. DO NOT EDIT.
// source: uber/cadence/api/v1/domain.proto

package apiv1

var yarpcFileDescriptorClosure824795d6ae7d8e2f = [][]byte{
	// uber/cadence/api/v1/domain.proto
	[]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x56, 0x5f, 0x6f, 0xdb, 0xb6,
		0x17, 0xfd, 0xc9, 0x4e, 0xf2, 0x73, 0xae, 0x1c, 0xd7, 0x65, 0x9a, 0x46, 0xf1, 0x86, 0x45, 0x4d,
		0x51, 0xc0, 0xeb, 0x83, 0xbc, 0x78, 0xc3, 0xd6, 0x6e, 0xd8, 0x83, 0x63, 0xa9, 0x9d, 0x87, 0x2c,
		0x0b, 0x64, 0x37, 0x0f, 0xdb, 0x83, 0x40, 0x4b, 0xb4, 0x4d, 0x54, 0x16, 0x05, 0x8a, 0x76, 0x9a,
		0xb7, 0x61, 0x1f, 0x62, 0x1f, 0x66, 0x8f, 0xfb, 0x64, 0x83, 0x28, 0x4a, 0xf1, 0x1f, 0x21, 0xd9,
		0x1b, 0x79, 0xef, 0x3d, 0x87, 0x47, 0x47, 0xf7, 0x52, 0x02, 0x73, 0x31, 0x26, 0xbc, 0xe3, 0xe3,
		0x80, 0x44, 0x3e, 0xe9, 0xe0, 0x98, 0x76, 0x96, 0xe7, 0x9d, 0x80, 0xcd, 0x31, 0x8d, 0xac, 0x98,
		0x33, 0xc1, 0xd0, 0x61, 0x5a, 0x61, 0xa9, 0x0a, 0x0b, 0xc7, 0xd4, 0x5a, 0x9e, 0xb7, 0xbe, 0x98,
		0x32, 0x36, 0x0d, 0x49, 0x47, 0x96, 0x8c, 0x17, 0x93, 0x4e, 0xb0, 0xe0, 0x58, 0x50, 0xa6, 0x40,
		0xad, 0xd3, 0xcd, 0xbc, 0xa0, 0x73, 0x92, 0x08, 0x3c, 0x8f, 0xb3, 0x82, 0xb3, 0xbf, 0x6a, 0xb0,
		0x67, 0xcb, 0x63, 0x50, 0x03, 0x2a, 0x34, 0x30, 0x34, 0x53, 0x6b, 0xef, 0xbb, 0x15, 0x1a, 0x20,
		0x04, 0x3b, 0x11, 0x9e, 0x13, 0xa3, 0x22, 0x23, 0x72, 0x8d, 0xde, 0xc2, 0x5e, 0x22, 0xb0, 0x58,
		0x24, 0x46, 0xd5, 0xd4, 0xda, 0x8d, 0xee, 0x0b, 0xab, 0x44, 0x95, 0x95, 0x11, 0x0e, 0x65, 0xa1,
		0xab, 0x00, 0xc8, 0x04, 0x3d, 0x20, 0x89, 0xcf, 0x69, 0x9c, 0xea, 0x33, 0x76, 0x24, 0xeb, 0x6a,
		0x08, 0x9d, 0x82, 0xce, 0x6e, 0x23, 0xc2, 0x3d, 0x32, 0xc7, 0x34, 0x34, 0x76, 0x65, 0x05, 0xc8,
		0x90, 0x93, 0x46, 0xd0, 0x5b, 0xd8, 0x09, 0xb0, 0xc0, 0xc6, 0x9e, 0x59, 0x6d, 0xeb, 0xdd, 0x57,
		0x0f, 0x9c, 0x6d, 0xd9, 0x58, 0x60, 0x27, 0x12, 0xfc, 0xce, 0x95, 0x10, 0x34, 0x83, 0x97, 0xb7,
		0x8c, 0x7f, 0x9c, 0x84, 0xec, 0xd6, 0x23, 0x9f, 0x88, 0xbf, 0x48, 0x4f, 0xf4, 0x38, 0x11, 0x24,
		0x92, 0xab, 0x98, 0x70, 0xca, 0x02, 0xe3, 0xff, 0xa6, 0xd6, 0xd6, 0xbb, 0x27, 0x56, 0x66, 0x9b,
		0x95, 0xdb, 0x66, 0xd9, 0xca, 0x56, 0xd7, 0xcc, 0x59, 0x9c, 0x9c, 0xc4, 0xcd, 0x39, 0xae, 0x25,
		0x05, 0xea, 0x43, 0x7d, 0x8c, 0x03, 0x6f, 0x4c, 0x23, 0xcc, 0x29, 0x49, 0x8c, 0x9a, 0xa4, 0x34,
		0x4b, 0xc5, 0x5e, 0xe0, 0xe0, 0x42, 0xd5, 0xb9, 0xfa, 0xf8, 0x7e, 0x83, 0x7e, 0x87, 0xe3, 0x19,
		0x4d, 0x04, 0xe3, 0x77, 0x1e, 0xe6, 0xfe, 0x8c, 0x2e, 0x71, 0xe8, 0x29, 0xe3, 0xf7, 0xa5, 0xf1,
		0x2f, 0x4b, 0xf9, 0x7a, 0xaa, 0x56, 0x59, 0x7f, 0xa4, 0x38, 0xd6, 0xc3, 0xe8, 0x2b, 0x78, 0xb6,
		0x45, 0xbe, 0xe0, 0xd4, 0x00, 0x69, 0x38, 0xda, 0x00, 0x7d, 0xe0, 0x14, 0x61, 0x68, 0x2d, 0x69,
		0x42, 0xc7, 0x34, 0xa4, 0x62, 0x5b, 0x91, 0xfe, 0xdf, 0x15, 0x19, 0xf7, 0x34, 0x1b, 0xa2, 0xbe,
		0x85, 0xe3, 0xb2, 0x23, 0x52, 0x5d, 0x75, 0xa9, 0xeb, 0x68, 0x1b, 0x9a, 0x4a, 0xb3, 0xe0, 0x10,
		0xfb, 0x82, 0x2e, 0x89, 0xe7, 0x87, 0x8b, 0x44, 0x10, 0xee, 0xc9, 0xa6, 0x3d, 0x90, 0x98, 0xa7,
		0x59, 0xaa, 0x9f, 0x65, 0xae, 0xd2, 0x0e, 0xbe, 0x86, 0x9a, 0x2a, 0x4c, 0x8c, 0x86, 0xec, 0xa3,
		0x6f, 0x4a, 0x85, 0x2b, 0x8c, 0x4b, 0xe2, 0x90, 0xfa, 0xf2, 0xdd, 0xf7, 0x59, 0x34, 0xa1, 0xd3,
		0xbc, 0x11, 0x0a, 0x16, 0xf4, 0x25, 0x34, 0x27, 0x98, 0x86, 0x6c, 0x49, 0xb8, 0xb7, 0x24, 0x3c,
		0x49, 0xbb, 0xfb, 0x89, 0xa9, 0xb5, 0xab, 0xee, 0x93, 0x3c, 0x7e, 0x93, 0x85, 0x51, 0x1b, 0x9a,
		0x34, 0xf1, 0xa6, 0x21, 0x1b, 0xe3, 0xd0, 0xcb, 0xa6, 0xdb, 0x68, 0x9a, 0x5a, 0xbb, 0xe6, 0x36,
		0x68, 0xf2, 0x5e, 0x86, 0xd5, 0x30, 0xbe, 0x83, 0x83, 0x82, 0x94, 0x46, 0x13, 0x66, 0x3c, 0x95,
		0x6d, 0x54, 0x3e, 0x6f, 0xef, 0x54, 0xe5, 0x20, 0x9a, 0x30, 0xb7, 0x3e, 0x59, 0xd9, 0xb5, 0xbe,
		0x83, 0xfd, 0x62, 0x14, 0x50, 0x13, 0xaa, 0x1f, 0xc9, 0x9d, 0x1a, 0xf1, 0x74, 0x89, 0x9e, 0xc1,
		0xee, 0x12, 0x87, 0x8b, 0x7c, 0xc8, 0xb3, 0xcd, 0xf7, 0x95, 0x37, 0xda, 0x99, 0x0d, 0xa7, 0x8f,
		0x58, 0x80, 0x5e, 0x40, 0x7d, 0xcd, 0xf3, 0x8c, 0x57, 0xf7, 0xef, 0xdd, 0x3e, 0xfb, 0x5b, 0x03,
		0x7d, 0xa5, 0xc9, 0xd1, 0xcf, 0x50, 0x2b, 0x06, 0x43, 0x93, 0xee, 0x5b, 0x8f, 0x0d, 0x86, 0x95,
		0x2f, 0xb2, 0x71, 0x2e, 0xf0, 0x2d, 0x0f, 0x0e, 0xd6, 0x52, 0x25, 0x8f, 0xf7, 0x66, 0xf5, 0xf1,
		0xf4, 0xee, 0xd9, 0x83, 0x67, 0xdd, 0x49, 0xfb, 0x56, 0x2c, 0xf8, 0x53, 0x83, 0x83, 0xb5, 0x24,
		0x7a, 0x0e, 0x7b, 0x9c, 0xe0, 0x84, 0x45, 0xea, 0x10, 0xb5, 0x43, 0x2d, 0xa8, 0xb1, 0x98, 0x70,
		0x2c, 0x18, 0x57, 0x4e, 0x16, 0x7b, 0xf4, 0x23, 0xd4, 0x7d, 0x4e, 0xb0, 0x20, 0x81, 0x97, 0x5e,
		0xbe, 0xf2, 0xe2, 0xd4, 0xbb, 0xad, 0xad, 0x2b, 0x66, 0x94, 0xdf, 0xcc, 0xae, 0xae, 0xea, 0xd3,
		0xc8, 0xd9, 0x3f, 0x15, 0xa8, 0xaf, 0xbe, 0xdf, 0xd2, 0x76, 0xd3, 0xca, 0xdb, 0x6d, 0x04, 0x46,
		0x51, 0x9a, 0x08, 0xcc, 0x85, 0x57, 0x5c, 0xff, 0xca, 0x91, 0x87, 0x64, 0x3c, 0xcf, 0xb1, 0xc3,
		0x14, 0x5a, 0xc4, 0xd1, 0x0d, 0x9c, 0x14, 0xac, 0xe4, 0x53, 0x4c, 0x39, 0x59, 0xa1, 0x7d, 0xfc,
		0xe9, 0x8e, 0x73, 0xb0, 0x23, 0xb1, 0xf7, 0xbc, 0x5d, 0x38, 0xf2, 0xd9, 0x3c, 0x0e, 0x49, 0x6a,
		0x55, 0x32, 0xc3, 0x3c, 0xf0, 0x7c, 0xb6, 0x88, 0x84, 0xfc, 0x54, 0xec, 0xba, 0x87, 0x45, 0x72,
		0x98, 0xe6, 0xfa, 0x69, 0x0a, 0xbd, 0x82, 0x46, 0x4c, 0xa2, 0x80, 0x46, 0xd3, 0x0c, 0x91, 0x18,
		0xbb, 0x66, 0xb5, 0xbd, 0xeb, 0x1e, 0xa8, 0xa8, 0x2c, 0x4d, 0x5e, 0xff, 0xa1, 0x41, 0x7d, 0xf5,
		0xa3, 0x84, 0x4e, 0xe0, 0xc8, 0xfe, 0xf5, 0x97, 0xde, 0xe0, 0xca, 0x1b, 0x8e, 0x7a, 0xa3, 0x0f,
		0x43, 0x6f, 0x70, 0x75, 0xd3, 0xbb, 0x1c, 0xd8, 0xcd, 0xff, 0xa1, 0xcf, 0xc1, 0x58, 0x4f, 0xb9,
		0xce, 0xfb, 0xc1, 0x70, 0xe4, 0xb8, 0x8e, 0xdd, 0xd4, 0xb6, 0xb3, 0xb6, 0x73, 0xed, 0x3a, 0xfd,
		0xde, 0xc8, 0xb1, 0x9b, 0x95, 0x6d, 0x5a, 0xdb, 0xb9, 0x74, 0xd2, 0x54, 0xf5, 0xf5, 0x0c, 0x1a,
		0x1b, 0x37, 0xde, 0x67, 0x70, 0xdc, 0x73, 0xfb, 0x3f, 0x0d, 0x6e, 0x7a, 0x97, 0xa5, 0x2a, 0x36,
		0x93, 0xf6, 0x60, 0xd8, 0xbb, 0xb8, 0x94, 0x2a, 0x4a, 0xa0, 0xce, 0x55, 0x96, 0xac, 0x5c, 0xdc,
		0xc0, 0xb1, 0xcf, 0xe6, 0x65, 0xad, 0x7e, 0x51, 0xeb, 0xc5, 0xf4, 0x3a, 0x7d, 0x25, 0xd7, 0xda,
		0x6f, 0x9d, 0x29, 0x15, 0xb3, 0xc5, 0xd8, 0xf2, 0xd9, 0xbc, 0xb3, 0xf6, 0xf3, 0x61, 0x4d, 0x49,
		0x94, 0xfd, 0x30, 0xa8, 0xff, 0x90, 0x1f, 0x70, 0x4c, 0x97, 0xe7, 0xe3, 0x3d, 0x19, 0xfb, 0xfa,
		0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x6b, 0x36, 0xf7, 0xb7, 0xab, 0x08, 0x00, 0x00,
	},
	// google/protobuf/duration.proto
	[]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4b, 0xcf, 0xcf, 0x4f,
		0xcf, 0x49, 0xd5, 0x2f, 0x28, 0xca, 0x2f, 0xc9, 0x4f, 0x2a, 0x4d, 0xd3, 0x4f, 0x29, 0x2d, 0x4a,
		0x2c, 0xc9, 0xcc, 0xcf, 0xd3, 0x03, 0x8b, 0x08, 0xf1, 0x43, 0xe4, 0xf5, 0x60, 0xf2, 0x4a, 0x56,
		0x5c, 0x1c, 0x2e, 0x50, 0x25, 0x42, 0x12, 0x5c, 0xec, 0xc5, 0xa9, 0xc9, 0xf9, 0x79, 0x29, 0xc5,
		0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0xcc, 0x41, 0x30, 0xae, 0x90, 0x08, 0x17, 0x6b, 0x5e, 0x62, 0x5e,
		0x7e, 0xb1, 0x04, 0x93, 0x02, 0xa3, 0x06, 0x6b, 0x10, 0x84, 0xe3, 0xd4, 0xcc, 0xc8, 0x25, 0x9c,
		0x9c, 0x9f, 0xab, 0x87, 0x66, 0xa6, 0x13, 0x2f, 0xcc, 0xc4, 0x00, 0x90, 0x48, 0x00, 0x63, 0x94,
		0x21, 0x54, 0x45, 0x7a, 0x7e, 0x4e, 0x62, 0x5e, 0xba, 0x5e, 0x7e, 0x51, 0x3a, 0xc2, 0x81, 0x25,
		0x95, 0x05, 0xa9, 0xc5, 0xfa, 0xd9, 0x79, 0xf9, 0xe5, 0x79, 0x70, 0xc7, 0x16, 0x24, 0xfd, 0x60,
		0x64, 0x5c, 0xc4, 0xc4, 0xec, 0x1e, 0xe0, 0xb4, 0x8a, 0x49, 0xce, 0x1d, 0xa2, 0x39, 0x00, 0xaa,
		0x43, 0x2f, 0x3c, 0x35, 0x27, 0xc7, 0x1b, 0xa4, 0x3e, 0x04, 0xa4, 0x35, 0x89, 0x0d, 0x6c, 0x94,
		0x31, 0x20, 0x00, 0x00, 0xff, 0xff, 0xef, 0x8a, 0xb4, 0xc3, 0xfb, 0x00, 0x00, 0x00,
	},
	// google/protobuf/timestamp.proto
	[]byte{
		0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0x4f, 0xcf, 0xcf, 0x4f,
		0xcf, 0x49, 0xd5, 0x2f, 0x28, 0xca, 0x2f, 0xc9, 0x4f, 0x2a, 0x4d, 0xd3, 0x2f, 0xc9, 0xcc, 0x4d,
		0x2d, 0x2e, 0x49, 0xcc, 0x2d, 0xd0, 0x03, 0x0b, 0x09, 0xf1, 0x43, 0x14, 0xe8, 0xc1, 0x14, 0x28,
		0x59, 0x73, 0x71, 0x86, 0xc0, 0xd4, 0x08, 0x49, 0x70, 0xb1, 0x17, 0xa7, 0x26, 0xe7, 0xe7, 0xa5,
		0x14, 0x4b, 0x30, 0x2a, 0x30, 0x6a, 0x30, 0x07, 0xc1, 0xb8, 0x42, 0x22, 0x5c, 0xac, 0x79, 0x89,
		0x79, 0xf9, 0xc5, 0x12, 0x4c, 0x0a, 0x8c, 0x1a, 0xac, 0x41, 0x10, 0x8e, 0x53, 0x2b, 0x23, 0x97,
		0x70, 0x72, 0x7e, 0xae, 0x1e, 0x9a, 0xa1, 0x4e, 0x7c, 0x70, 0x23, 0x03, 0x40, 0x42, 0x01, 0x8c,
		0x51, 0x46, 0x50, 0x25, 0xe9, 0xf9, 0x39, 0x89, 0x79, 0xe9, 0x7a, 0xf9, 0x45, 0xe9, 0x48, 0x6e,
		0xac, 0x2c, 0x48, 0x2d, 0xd6, 0xcf, 0xce, 0xcb, 0x2f, 0xcf, 0x43, 0xb8, 0xb7, 0x20, 0xe9, 0x07,
		0x23, 0xe3, 0x22, 0x26, 0x66, 0xf7, 0x00, 0xa7, 0x55, 0x4c, 0x72, 0xee, 0x10, 0xdd, 0x01, 0x50,
		0x2d, 0x7a, 0xe1, 0xa9, 0x39, 0x39, 0xde, 0x20, 0x0d, 0x21, 0x20, 0xbd, 0x49, 0x6c, 0x60, 0xb3,
		0x8c, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0xae, 0x65, 0xce, 0x7d, 0xff, 0x00, 0x00, 0x00,
	},
}
