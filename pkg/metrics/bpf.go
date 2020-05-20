// Copyright 2018-2020 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"context"
	"encoding/json"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type bpfCollector struct {
	bpfMemory *prometheus.Desc
}

func newbpfCollector() *bpfCollector {
	return &bpfCollector{
		bpfMemory: prometheus.NewDesc(
			prometheus.BuildFQName(Namespace, "", "bpf_maps_memory_bytes"),
			"BPF kernel memory usage size in bytes.",
			nil, nil,
		),
	}
}

func (s *bpfCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.bpfMemory
}

type memoryEntry struct {
	BytesMemlock uint64 `json:"bytes_memlock"`
}

func (s *bpfCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bpftool", "-j", "map", "show")
	out, err := cmd.Output()
	if err != nil {
		log.WithError(err).Error("Error while getting bpftool output")
		return
	}

	var memoryEntries []memoryEntry
	err = json.Unmarshal(out, &memoryEntries)
	if err != nil {
		log.WithError(err).Error("Error while unmarshalling bpftool output")
		return
	}
	var totalMem uint64
	for _, entry := range memoryEntries {
		totalMem += entry.BytesMemlock
	}

	ch <- prometheus.MustNewConstMetric(
		s.bpfMemory,
		prometheus.GaugeValue,
		float64(totalMem),
	)

}
