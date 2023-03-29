package query

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

type Builder struct {
	query                 Query         // query
	from                  int           // from
	size                  int           // size
	sorters               []Sorter      // sort
	searchAfterSortValues []interface{} // search_after

}

func NewBuilder() *Builder {
	return &Builder{
		from: -1,
		size: -1,
	}
}

func (s *Builder) Query(query Query) *Builder {
	s.query = query
	return s
}

func (s *Builder) From(from int) *Builder {
	s.from = from
	return s
}

func (s *Builder) Sortby(sorters ...Sorter) *Builder {
	s.sorters = sorters
	return s
}

func (s *Builder) Size(size int) *Builder {
	s.size = size
	return s
}

func (s *Builder) SearchAfter(v ...interface{}) *Builder {
	s.searchAfterSortValues = v
	return s
}

// Source returns the serializable JSON for the source builder.
func (s *Builder) Source() (interface{}, error) {
	source := make(map[string]interface{})

	if s.from != -1 {
		source["from"] = s.from
	}
	if s.size != -1 {
		source["size"] = s.size
	}

	if s.query != nil {
		src, err := s.query.Source()
		if err != nil {
			return nil, err
		}
		source["query"] = src
	}
	if len(s.sorters) > 0 {
		var sortarr []interface{}
		for _, sorter := range s.sorters {
			src, err := sorter.Source()
			if err != nil {
				return nil, err
			}
			sortarr = append(sortarr, src)
		}
		source["sort"] = sortarr
	}

	if len(s.searchAfterSortValues) > 0 {
		source["search_after"] = s.searchAfterSortValues
	}

	return source, nil
}
