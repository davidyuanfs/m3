// Copyright (c) 2018 Uber Technologies, Inc.
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

package filter

import (
	"testing"

	"github.com/m3db/m3/src/query/storage"
	"github.com/m3db/m3/src/query/storage/mock"

	"github.com/m3db/m3/src/query/models"
	"github.com/stretchr/testify/assert"
)

var (
	local  mock.Storage
	remote mock.Storage
	multi  mock.Storage

	q = &storage.FetchQuery{}
)

func init() {
	local = mock.NewMockStorage()
	local.SetTypeResult(storage.TypeLocalDC)
	remote = mock.NewMockStorage()
	remote.SetTypeResult(storage.TypeRemoteDC)
	multi = mock.NewMockStorage()
	multi.SetTypeResult(storage.TypeMultiDC)
}

func TestLocalOnly(t *testing.T) {
	assert.True(t, LocalOnly(q, local))
	assert.False(t, LocalOnly(q, remote))
	assert.False(t, LocalOnly(q, multi))
}

func TestAllowAll(t *testing.T) {
	assert.True(t, AllowAll(q, local))
	assert.True(t, AllowAll(q, remote))
	assert.True(t, AllowAll(q, multi))
}

func TestAllowNone(t *testing.T) {
	assert.False(t, AllowNone(q, local))
	assert.False(t, AllowNone(q, remote))
	assert.False(t, AllowNone(q, multi))
}

func TestReadOptimizedFilterWithEqual(t *testing.T) {
	fetchQueryToOregionDev := storage.FetchQuery{
		TagMatchers: models.Matchers{
			models.Matcher{
				Type:  models.MatchEqual,
				Name:  []byte(storageNameLabelKey),
				Value: []byte("oregon-dev"),
			},
			models.Matcher{
				Type:  models.MatchEqual,
				Name:  []byte("some-other-label"),
				Value: []byte("something-else"),
			},
		},
	}
	{
		store := mock.NewMockStorageWithName(localStorageName)
		assert.True(t, ReadOptimizedFilter(&fetchQueryToOregionDev, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"dev-aws-oregon-dev")
		assert.True(t, ReadOptimizedFilter(&fetchQueryToOregionDev, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"dev-aws-us-east-1")
		assert.False(t, ReadOptimizedFilter(&fetchQueryToOregionDev, store))
	}
	fetchQueryToOregionDev = storage.FetchQuery{
		TagMatchers: models.Matchers{
			models.Matcher{
				Type:  models.MatchNotEqual,
				Name:  []byte(storageNameLabelKey),
				Value: []byte("oregon-dev"),
			},
			models.Matcher{
				Type:  models.MatchEqual,
				Name:  []byte("some-other-label"),
				Value: []byte("something-else"),
			},
		},
	}
	{
		store := mock.NewMockStorageWithName(localStorageName)
		assert.True(t, ReadOptimizedFilter(&fetchQueryToOregionDev, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"dev-aws-oregon-dev")
		assert.False(t, ReadOptimizedFilter(&fetchQueryToOregionDev, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"dev-aws-us-east-1")
		assert.True(t, ReadOptimizedFilter(&fetchQueryToOregionDev, store))
	}
}

func TestReadOptimizedFilterWithRegex(t *testing.T) {
	fetchQuery := storage.FetchQuery{
		TagMatchers: models.Matchers{
			models.Matcher{
				Type:  models.MatchRegexp,
				Name:  []byte(storageNameLabelKey),
				Value: []byte("shard-a|shard-b-bb-bbb|shard-c-ccc-d"),
			},
			models.Matcher{
				Type:  models.MatchRegexp,
				Name:  []byte("some-other-label"),
				Value: []byte("suffix-d"),
			},
		},
	}
	{
		store := mock.NewMockStorageWithName(localStorageName)
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-a")
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-b-bb-bbb")
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-c-ccc-d")
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}

	fetchQuery = storage.FetchQuery{
		TagMatchers: models.Matchers{
			models.Matcher{
				Type:  models.MatchNotRegexp,
				Name:  []byte(storageNameLabelKey),
				Value: []byte("shard-a|shard-b-bb-bbb|shard-c-ccc-d"),
			},
			models.Matcher{
				Type:  models.MatchRegexp,
				Name:  []byte("some-other-label"),
				Value: []byte("suffix-d"),
			},
		},
	}
	{
		store := mock.NewMockStorageWithName(localStorageName)
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-a")
		assert.False(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-b-bb-bbb")
		assert.False(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-c-ccc-d")
		assert.False(t, ReadOptimizedFilter(&fetchQuery, store))
	}

	fetchQuery = storage.FetchQuery{
		TagMatchers: models.Matchers{
			models.Matcher{
				Type:  models.MatchRegexp,
				Name:  []byte(storageNameLabelKey),
				Value: []byte("shard-a|shard-a-.*"),
			},
			models.Matcher{
				Type:  models.MatchRegexp,
				Name:  []byte("some-other-label"),
				Value: []byte("suffix-d"),
			},
		},
	}
	{
		store := mock.NewMockStorageWithName(localStorageName)
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-a")
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-aa")
		assert.False(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-a-")
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-a-aa")
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}

	fetchQuery = storage.FetchQuery{
		TagMatchers: models.Matchers{
			models.Matcher{
				Type:  models.MatchRegexp,
				Name:  []byte(storageNameLabelKey),
				Value: []byte(".*"),
			},
			models.Matcher{
				Type:  models.MatchRegexp,
				Name:  []byte("some-other-label"),
				Value: []byte("suffix-d"),
			},
		},
	}
	{
		store := mock.NewMockStorageWithName(localStorageName)
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-a")
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-aa")
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-a-")
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
	{
		store := mock.NewMockStorageWithName(remoteStorePrefix+"prod-aws-shard-a-aa")
		assert.True(t, ReadOptimizedFilter(&fetchQuery, store))
	}
}