// Copyright (c) 2019 Uber Technologies, Inc.
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

package authorization

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/yarpc/api/encoding"
	"go.uber.org/yarpc/api/transport"
	"golang.org/x/net/context"

	"github.com/uber/cadence/common"
	"github.com/uber/cadence/common/config"
)

var pubKeyTest = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAscukltHilaq+o5gIVE4P
GwWl+esvJ2EaEpWw6ogr98Un11YJ4oKkwIkLw4iIo0tveCINA3cZmxaW1RejRWKE
qYFtQ1rYd6BsnFAHXWh2R3A1FtpG6ANUEGkE7OAJe2/L42E/ImJ+GQxRvartInDM
yfiRfB7+L2n3wG+Ni+hBNMtAaX4Wwbj2hup21Jjuo96TuhcGImBFBATGWaYR2wqe
/6by9wJexPHlY/1uDp3SnzF1dCLjp76SGCfyYqOGC/PxhQi7mDxeH9/tIC+lt/Sz
wc1n8gZLtlRlZHinvYa8lhWXqVYw6WD8h4LTgALq9iY+beD1PFQSY1GkQtt0RhRw
eQIDAQAB
-----END PUBLIC KEY-----`

var privKeyTest = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQCxy6SW0eKVqr6j
mAhUTg8bBaX56y8nYRoSlbDqiCv3xSfXVgnigqTAiQvDiIijS294Ig0DdxmbFpbV
F6NFYoSpgW1DWth3oGycUAddaHZHcDUW2kboA1QQaQTs4Al7b8vjYT8iYn4ZDFG9
qu0icMzJ+JF8Hv4vaffAb42L6EE0y0BpfhbBuPaG6nbUmO6j3pO6FwYiYEUEBMZZ
phHbCp7/pvL3Al7E8eVj/W4OndKfMXV0IuOnvpIYJ/Jio4YL8/GFCLuYPF4f3+0g
L6W39LPBzWfyBku2VGVkeKe9hryWFZepVjDpYPyHgtOAAur2Jj5t4PU8VBJjUaRC
23RGFHB5AgMBAAECggEABj1T9Orf0W9nskDQ2QQ7cuVdZEJjpMrbTK1Aw1L8/Qc9
TSkINDEayaV9mn1RXe61APcBSdP4ER7nXfTZiQ21LhLcWWg9T3cbh1b70oRqyI9z
Pi6HSBeWz4kfUBX9izMQFBZKzjYn6qaJp1b8bGXKRWkcvPRZqLhmsRPmeH3xrOHe
qsIDhYXMjRoOgEUxLbk8iPLP6nx0icPJl/tHK2l76R+1Ko6TBE69Md2krUIuh0u4
nm9n+Az+0GuvkFsLw5KMGhSBeqB+ez5qtFa8T8CUCn98IjiUDOwgZdFrNldFLcZf
putw7O2qCA9LT+mFBQ6CVsVu/9tKeXQ9sJ7p3lxhwQKBgQDjt7HNIabLncdXPMu0
ByRyNVme0+Y1vbj9Q7iodk77hvlzWpD1p5Oyvq7cN+Cb4c1iO/ZQXMyUw+9hLgmf
LNquH2d4hK1Jerzc/ciwu6dUBsCW8+0VJd4M2UNN15rJMPvbZGmqMq9Np1iCTCjE
dvHo7xjPcJhsbhMbHq+PaUU7OQKBgQDH4KuaHBFTGUPkRaQGAZNRB8dDvSExV6ID
Pblzr80g9kKHUnQCQfIDLjHVgDbTaSCdRw7+EXRyRmLy5mfPWEbUFfIemEpEcEcb
3geWeVDx4Z/FwprWFuVifRopRSQ/FAbMXLIui7OHXWLEtzBvLkR/uS2VIVPm10PV
pbh2EXifQQKBgQDbcOLbjelBYLt/euvGgfeCQ50orIS1Fy5UidVCKjh0tR5gJk95
G1L+tjilqQc+0LtuReBYkwTm+2YMXSQSi1P05fh9MEYZgDjOMZYbkcpu887V6Rx3
+7Te5uOv+OyFozmhs0MMK6m5iGGHtsK2iPUYBoj/Jj8MhorM4KZH6ic4KQKBgQCl
3zIpg09xSc9Iue5juZz6qtzXvzWzkAj4bZnggq1VxGfzix6Q3Q8tSoG6r1tQWLbj
Lpwnhm6/guAMud6+eIDW8ptqfnFrmE26t6hOXMEq6lXANT5vmrKj6DP0uddZrZHy
uJ55+B91n68elvPP4HKiGBfW4cCSGmTGAXAyM0+JwQKBgQCz2cNiFrr+oEnlHDLg
EqsiEufppT4FSZPy9/MtuWuMgEOBu34cckYaai+nahQLQvH62KskTK0EUjE1ywub
NPORuXcugxIBMHWyseOS7lrtrlSBxU9gntS7jHdM3IMrrUy9YZBvPvFGP0wLdpKM
nvt3vT46hs3n28XZpb18uRkSDw==
-----END PRIVATE KEY-----`

type Mocks struct {
	cfg   config.OAuthAuthorizer
	token string
	ctx   context.Context
	att   Attributes
}

func getMocksBase(t *testing.T) Mocks {
	cfg := config.OAuthAuthorizer{
		Enable: true,
		JwtCredentials: config.JwtCredentials{
			Algorithm:  "RS256",
			PublicKey:  pubKeyTest,
			PrivateKey: privKeyTest,
		},
		JwtTTL: 100000,
	}
	token := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6Ikp" +
		"vaG4gRG9lIiwicGVybWlzc2lvbiI6InJlYWQiLCJkb21haW4iOiJ0ZXN0LWRvbWFpbiIsImlhdCI6MTkx" +
		"NjIzOTAyMX0.m9W50TgDsfDgquFAmKQh5olJqbyueWpPWGV3bSeH3uVVgjER-og5AyeEj4v--6WyF6R48" +
		"etuK40YEI3jlf5oAf_o0hhvCfYcSQS59RmGvPtfpNLTk1WCEjZPxnQ7BKbGZ7Mc99Ey5VLceFqwB4kEgH" +
		"UJhrRzf3ClMYG2V-45r4qmxOsdQyQSxBSJ-G8LTDJR_BDUvXFA6mdXGwsTRGkj-dYD_EgbECRNqysEG78" +
		"4YAMxJB7RHbaTZs6KpjDw04k4wKDFCQZUpshB3O3F3DFakYtaV4G0GeHxydEyRMIE1OtlCbnIoqnwAsln" +
		"zADSdEQo4kjsVseDbUq3f8S8EDIARw"

	ctx := context.Background()
	ctx, call := encoding.NewInboundCall(ctx)
	err := call.ReadFromRequest(&transport.Request{
		Headers: transport.NewHeaders().With(common.AuthorizationTokenHeaderName, token),
	})
	assert.NoError(t, err)

	att := Attributes{
		Actor:      "admin",
		APIName:    "",
		DomainName: "test",
		TaskList:   nil,
	}

	return Mocks{
		cfg:   cfg,
		ctx:   ctx,
		token: token,
		att:   att,
	}
}

func TestCorrectPayload(t *testing.T) {
	mocks := getMocksBase(t)
	authorizer := NewOAuthAuthorizer(mocks.cfg)
	result, err := authorizer.Authorize(mocks.ctx, &mocks.att)
	assert.NoError(t, err)
	assert.Equal(t, result.Decision, DecisionAllow)
}
