// PromHouse
// Copyright (C) 2017 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/common/model"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/Percona-Lab/PromHouse/prompb"
	"github.com/Percona-Lab/PromHouse/storages/blackhole"
)

func getWriteRequest() *prompb.WriteRequest {
	start := model.Now().Add(-6 * time.Second)

	return &prompb.WriteRequest{
		TimeSeries: []*prompb.TimeSeries{
			{
				Labels: []*prompb.Label{
					{Name: "__name__", Value: "http_requests_total"},
					{Name: "code", Value: "200"},
					{Name: "handler", Value: "query"},
				},
				Samples: []*prompb.Sample{
					{Value: 13, TimestampMs: int64(start)},
					{Value: 14, TimestampMs: int64(start.Add(1 * time.Second))},
					{Value: 14, TimestampMs: int64(start.Add(2 * time.Second))},
					{Value: 14, TimestampMs: int64(start.Add(3 * time.Second))},
					{Value: 15, TimestampMs: int64(start.Add(4 * time.Second))},
				},
			},
			{
				Labels: []*prompb.Label{
					{Name: "__name__", Value: "http_requests_total"},
					{Name: "code", Value: "400"},
					{Name: "handler", Value: "query_range"},
				},
				Samples: []*prompb.Sample{
					{Value: 9, TimestampMs: int64(start)},
					{Value: 9, TimestampMs: int64(start.Add(1 * time.Second))},
					{Value: 9, TimestampMs: int64(start.Add(2 * time.Second))},
					{Value: 11, TimestampMs: int64(start.Add(3 * time.Second))},
					{Value: 11, TimestampMs: int64(start.Add(4 * time.Second))},
				},
			},
			{
				Labels: []*prompb.Label{
					{Name: "__name__", Value: "http_requests_total"},
					{Name: "code", Value: "200"},
					{Name: "handler", Value: "prometheus"},
				},
				Samples: []*prompb.Sample{
					{Value: 591, TimestampMs: int64(start)},
					{Value: 592, TimestampMs: int64(start.Add(1 * time.Second))},
					{Value: 593, TimestampMs: int64(start.Add(2 * time.Second))},
					{Value: 594, TimestampMs: int64(start.Add(3 * time.Second))},
					{Value: 595, TimestampMs: int64(start.Add(4 * time.Second))},
				},
			},
		},
	}
}

func TestWrite(t *testing.T) {
	h := PromAPI{
		Storage: blackhole.New(),
		Logger: logrus.NewEntry(&logrus.Logger{
			Level: logrus.FatalLevel,
		}),
	}

	data, err := proto.Marshal(getWriteRequest())
	require.NoError(t, err)
	r := bytes.NewReader(snappy.Encode(nil, data))
	req, err := http.NewRequest("", "", r)
	require.NoError(t, err)
	require.NoError(t, h.Write(context.Background(), nil, req))
}

func BenchmarkWrite(b *testing.B) {
	h := PromAPI{
		Storage: blackhole.New(),
		Logger: logrus.NewEntry(&logrus.Logger{
			Level: logrus.FatalLevel,
		}),
	}

	data, err := proto.Marshal(getWriteRequest())
	require.NoError(b, err)
	r := bytes.NewReader(snappy.Encode(nil, data))
	req, err := http.NewRequest("", "", r)
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Seek(0, io.SeekStart)
		err = h.Write(context.Background(), nil, req)
	}
	b.StopTimer()

	require.NoError(b, err)
}