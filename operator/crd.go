// Copyright 2020 Authors of Cilium
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

package main

import (
	"context"
	"time"

	"github.com/cilium/cilium/pkg/controller"

	"github.com/pkg/errors"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getCRD(client clientset.Interface, name string) error {
	log.Debugf("Getting CRD %s from Kubernetes apiserver", name)
	if _, err := client.ApiextensionsV1beta1().CustomResourceDefinitions().Get(context.TODO(), name, metav1.GetOptions{}); err != nil {
		log.WithError(err).Debugf("Could not get CRD %s from Kubernetes apiserver, retrying", name)
		return err
	}
	log.Debugf("Check of CRD %s successful", name)
	return nil
}

// waitForCRD waits for the given CRD to be available until the timeout defined
// by CRDRetry. Returns an error when timeout exceeded.
func waitForCRD(client clientset.Interface, name string) error {
	ctrlMgr := controller.NewManager()
	ctrlName := "wait-for-crd-" + name
	ctrlMgr.UpdateController(ctrlName, controller.ControllerParams{
		RunInterval: 15 * time.Second,
		DoFunc: func(ctx context.Context) error {
			return getCRD(client, name)
		},
	})
	if err := ctrlMgr.RemoveControllerAndWait(ctrlName); err != nil {
		return errors.Wrapf(err, "Could not wait for CRD %s status", name)
	}
	return nil
}
