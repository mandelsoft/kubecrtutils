package main

import (
	"fmt"
	"os"

	"github.com/mandelsoft/kubecrtutils/cluster/cluster"
	"github.com/mandelsoft/kubecrtutils/cluster/config"
	"github.com/mandelsoft/kubecrtutils/merge"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/client-go/tools/clientcmd"
)

func Error(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err)
		os.Exit(1)
	}
}

func main() {
	// 1. Load Kubernetes configuration
	// Uses the KUBECONFIG environment variable or default location
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath == "" {
		kubeconfigPath = clientcmd.RecommendedHomeFile
	}

	rcfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	Error(err)

	cfg := &config.Config{
		RestConfig: rcfg,
	}
	cl, err := cluster.NewCluster("test", cfg)
	Error(err)
	conv := cl.GetTypeConverter()

	live := `
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    annotations:
      cross-cluster.io/owner-id: dataplane#1fhyywfw31te2v9c/coredns.mandelsoft.org/HostedZone/test/test
      deployment.kubernetes.io/revision: "1"
      hashes.config.yaml: ae16fe3aa7d738ce829ff343e98ec436d2008b3ac0727b012927c1835f97b0a7
    creationTimestamp: "2026-05-04T16:24:26Z"
    generation: 2
    labels:
      app: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
      target: runtime
    managedFields:
    - apiVersion: apps/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            .: {}
            f:cross-cluster.io/owner-id: {}
            f:hashes.config.yaml: {}
          f:labels:
            .: {}
            f:app: {}
            f:target: {}
        f:spec:
          f:progressDeadlineSeconds: {}
          f:replicas: {}
          f:revisionHistoryLimit: {}
          f:selector: {}
          f:strategy:
            f:rollingUpdate:
              .: {}
              f:maxSurge: {}
              f:maxUnavailable: {}
            f:type: {}
          f:template:
            f:metadata:
              f:labels:
                .: {}
                f:app: {}
            f:spec:
              f:automountServiceAccountToken: {}
              f:containers:
                k:{"name":"dns-container"}:
                  .: {}
                  f:args: {}
                  f:image: {}
                  f:imagePullPolicy: {}
                  f:name: {}
                  f:ports:
                    .: {}
                    k:{"containerPort":53,"protocol":"UDP"}:
                      .: {}
                      f:containerPort: {}
                      f:name: {}
                      f:protocol: {}
                  f:resources: {}
                  f:terminationMessagePath: {}
                  f:terminationMessagePolicy: {}
                  f:volumeMounts:
                    .: {}
                    k:{"mountPath":"/etc/coredns/config"}:
                      .: {}
                      f:mountPath: {}
                      f:name: {}
                      f:readOnly: {}
              f:dnsPolicy: {}
              f:restartPolicy: {}
              f:schedulerName: {}
              f:securityContext: {}
              f:terminationGracePeriodSeconds: {}
              f:volumes:
                .: {}
                k:{"name":"config-volume"}:
                  .: {}
                  f:configMap:
                    .: {}
                    f:defaultMode: {}
                    f:items: {}
                    f:name: {}
                  f:name: {}
      manager: coredns.mandelsoft.org/hostedzone
      operation: Update
      time: "2026-05-04T18:15:31Z"
    - apiVersion: apps/v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            f:deployment.kubernetes.io/revision: {}
        f:status:
          f:availableReplicas: {}
          f:conditions:
            .: {}
            k:{"type":"Available"}:
              .: {}
              f:lastTransitionTime: {}
              f:lastUpdateTime: {}
              f:message: {}
              f:reason: {}
              f:status: {}
              f:type: {}
            k:{"type":"Progressing"}:
              .: {}
              f:lastTransitionTime: {}
              f:lastUpdateTime: {}
              f:message: {}
              f:reason: {}
              f:status: {}
              f:type: {}
          f:observedGeneration: {}
          f:readyReplicas: {}
          f:replicas: {}
          f:updatedReplicas: {}
      manager: kube-controller-manager
      operation: Update
      subresource: status
      time: "2026-05-04T18:15:31Z"
    name: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
    namespace: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
    resourceVersion: "1329146"
    uid: 1224c085-1e15-4fa8-a33b-368d1c1f415a
  spec:
    progressDeadlineSeconds: 600
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
    strategy:
      rollingUpdate:
        maxSurge: 25%
        maxUnavailable: 25%
      type: RollingUpdate
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
      spec:
        automountServiceAccountToken: false
        containers:
        - args:
          - -conf
          - /etc/coredns/config/Corefile
          image: mandelsoft/restdyndns-coredns:latest
          imagePullPolicy: Always
          name: dns-container
          ports:
          - containerPort: 53
            name: dns-udp
            protocol: UDP
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /etc/coredns/config
            name: config-volume
            readOnly: true
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            items:
            - key: Corefile
              path: Corefile
            name: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
          name: config-volume
`
	desired := `
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    annotations:
      cross-cluster.io/owner-id: dataplane#1fhyywfw31te2v9c/coredns.mandelsoft.org/HostedZone/test/test
      hashes.config.yaml: ae16fe3aa7d738ce829ff343e98ec436d2008b3ac0727b012927c1835f97b0a7
    labels:
      app: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
      target: runtime
    name: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
    namespace: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
  spec:
    progressDeadlineSeconds: 600
    replicas: 1
    revisionHistoryLimit: 10
    selector:
      matchLabels:
        app: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
    strategy:
      rollingUpdate:
        maxSurge: 25%
        maxUnavailable: 25%
      type: RollingUpdate
    template:
      metadata:
        creationTimestamp: null
        labels:
          app: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
        annotations:
          hash: hash
      spec:
        automountServiceAccountToken: false
        containers:
        - args:
          - -conf
          - /etc/coredns/config/Corefile
          image: mandelsoft/restdyndns-coredns:latestx
          imagePullPolicy: Always
          name: dns-container
          ports:
          - containerPort: 53
            name: dns-udp
            protocol: UDP
          resources: {}
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /etc/coredns/config
            name: config-volume
            readOnly: true
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        terminationGracePeriodSeconds: 30
        volumes:
        - configMap:
            defaultMode: 420
            items:
            - key: Corefile
              path: Corefile
            name: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
          name: config-volume

`

	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	liveObj := &unstructured.Unstructured{}
	_, _, err = dec.Decode([]byte(live), nil, liveObj)
	Error(err)
	desiredObj := &unstructured.Unstructured{}
	_, _, err = dec.Decode([]byte(desired), nil, desiredObj)
	Error(err)

	manager := "kube-controller-manager/hostedzone"

	merger, err := merge.NewObjectMerger(conv, cl.GetScheme(), manager)
	Error(err)
	p, err := merger.ComputeSSAPatch(liveObj, desiredObj)
	Error(err)
	fmt.Printf("PATCH: %s\n", string(p))
}
