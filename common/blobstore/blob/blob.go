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

package blob

/**

In order to add a new compression format do the following:
1. Add a new compression constant
2. Change Compress method to use new compression package
3. Add a case to Decompress to handle new compression format
4. Update unit tests

*/

//const (
//	compressionTag = "compression"
//
//	// following are all compression formats that have ever been used, names should describe the package that was used for compression
//	gzipCompression = "compression/gzip"
//)



type Blob struct {
	Body []byte
	Tags map[string]string
}

func NewBlob(body []byte, tags map[string]string) Blob {
	return Blob{
		Body: body,
		Tags: tags,
	}
}

func DeepCopy(blob Blob) Blob {
	tagsCopy := make(map[string]string, len(blob.Tags))
	for k, v := range blob.Tags {
		tagsCopy[k] = v
	}
	bodyCopy := make([]byte, len(blob.Body), len(blob.Body))
	for i, b := range blob.Body {
		bodyCopy[i] = b
	}
	return Blob{
		Body: bodyCopy,
		Tags: tagsCopy,
	}
}

//
//type (



	// Blob is the entity that blobstore handles
	//Blob interface {
	//	Body() []byte
	//	Tags() map[string]string
	//	Compress() (Blob, error)
	//	Decompress() (Blob, error)
	//	Compressed() bool
	//}
	//
	//blob struct {
	//	body []byte
	//	tags map[string]string
	//}
//)

//// NewBlob constructs blob with body and tags
//func NewBlob(body []byte, tags map[string]string) Blob {
//	return &blob{
//		body: body,
//		tags: tags,
//	}
//}
//
//// Body returns blob's body
//func (b *blob) Body() []byte {
//	return b.body
//}
//
//// Tags returns blob's tags
//func (b *blob) Tags() map[string]string {
//	return b.tags
//}

// Compress compresses blob returning a new blob. Returned Blob does not share any references with this Blob.
//func (b *blob) Compress() (Blob, error) {
//	if b.Compressed() {
//		return nil, fmt.Errorf("blob is already compressed")
//	}
//	var buf bytes.Buffer
//	w := gzip.NewWriter(&buf)
//	if _, err := w.Write(b.body); err != nil {
//		w.Close()
//		return nil, err
//	}
//	// must call close before accessing buf.Bytes()
//	w.Close()
//
//	compressedBody := buf.Bytes()
//	tags := duplicateTags(b.tags)
//	tags[compressionTag] = gzipCompression
//	return &blob{
//		tags: tags,
//		body: compressedBody,
//	}, nil
//}

// Decompress decompresses blob returning a new blob. Returned Blob does not share any references with this Blob.
//func (b *blob) Decompress() (Blob, error) {
//	if !b.Compressed() {
//		return nil, fmt.Errorf("blob already decompressed")
//	}
//	compression := b.tags[compressionTag]
//	tags := duplicateTags(b.tags)
//	delete(tags, compressionTag)
//
//	switch compression {
//	case gzipCompression:
//		dBody, err := gzipDecompress(b.body)
//		if err != nil {
//			return nil, err
//		}
//		return &blob{
//			tags: tags,
//			body: dBody,
//		}, nil
//	default:
//		// this should never happen
//		return nil, fmt.Errorf("blob has unknown compression of %v, cannot decompress", compression)
//	}
//}

// Compressed returns true if the blob is compressed false otherwise
//func (b *blob) Compressed() bool {
//	_, ok := b.tags[compressionTag]
//	return ok
//}
//
//func gzipDecompress(data []byte) ([]byte, error) {
//	r, err := gzip.NewReader(bytes.NewReader(data))
//	if err != nil {
//		return nil, err
//	}
//	defer func() {
//		r.Close()
//	}()
//	return ioutil.ReadAll(r)
//}
//
//func duplicateTags(tags map[string]string) map[string]string {
//	result := make(map[string]string, len(tags))
//	for k, v := range tags {
//		result[k] = v
//	}
//	return result
//}
