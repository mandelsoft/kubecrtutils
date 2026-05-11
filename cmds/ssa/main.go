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

var serviceLive = `
  apiVersion: v1
  kind: Service
  metadata:
    annotations:
      cross-cluster.io/owner-id: dataplane#1fhyywfw31te2v9c/coredns.mandelsoft.org/HostedZone/test/test
      dns.gardener.cloud/class: garden
      dns.gardener.cloud/dnsnames: test.test.dnsservice.dns-runtime.mandelsoft.shoot.canary.k8s-hana.ondemand.com
      dns.gardener.cloud/ttl: "60"
      service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: instance
      service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing
      service.beta.kubernetes.io/aws-load-balancer-type: external
    creationTimestamp: "2026-05-11T10:08:08Z"
    finalizers:
    - garden.dns.gardener.cloud/service-dns
    - service.k8s.aws/resources
    labels:
      app: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
      target: runtime
    managedFields:
    - apiVersion: v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:finalizers:
            v:"service.k8s.aws/resources": {}
      manager: controller
      operation: Update
      time: "2026-05-11T10:08:08Z"
    - apiVersion: v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:annotations:
            .: {}
            f:cross-cluster.io/owner-id: {}
            f:dns.gardener.cloud/class: {}
            f:dns.gardener.cloud/dnsnames: {}
            f:dns.gardener.cloud/ttl: {}
            f:service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: {}
            f:service.beta.kubernetes.io/aws-load-balancer-scheme: {}
            f:service.beta.kubernetes.io/aws-load-balancer-type: {}
          f:labels:
            .: {}
            f:app: {}
            f:target: {}
        f:spec:
          f:allocateLoadBalancerNodePorts: {}
          f:externalTrafficPolicy: {}
          f:internalTrafficPolicy: {}
          f:loadBalancerClass: {}
          f:ports:
            .: {}
            k:{"port":53,"protocol":"UDP"}:
              .: {}
              f:name: {}
              f:port: {}
              f:protocol: {}
              f:targetPort: {}
          f:selector: {}
          f:sessionAffinity: {}
          f:type: {}
      manager: coredns.mandelsoft.org/hostedzone
      operation: Update
      time: "2026-05-11T10:08:08Z"
    - apiVersion: v1
      fieldsType: FieldsV1
      fieldsV1:
        f:metadata:
          f:finalizers:
            .: {}
            v:"garden.dns.gardener.cloud/service-dns": {}
      manager: dns-controller-manager
      operation: Update
      time: "2026-05-11T10:08:08Z"
    - apiVersion: v1
      fieldsType: FieldsV1
      fieldsV1:
        f:status:
          f:loadBalancer:
            f:ingress: {}
      manager: controller
      operation: Update
      subresource: status
      time: "2026-05-11T10:08:11Z"
    name: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
    namespace: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
    resourceVersion: "1469958"
    uid: c780b86f-dcb1-4890-b587-7a6df19a361d
  spec:
    allocateLoadBalancerNodePorts: true
    clusterIP: 100.109.176.68
    clusterIPs:
    - 100.109.176.68
    externalTrafficPolicy: Cluster
    internalTrafficPolicy: Cluster
    ipFamilies:
    - IPv4
    ipFamilyPolicy: SingleStack
    loadBalancerClass: service.k8s.aws/nlb
    ports:
    - name: dns-udp-port
      nodePort: 31802
      port: 53
      protocol: UDP
      targetPort: 53
    selector:
      app: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
    sessionAffinity: None
    type: LoadBalancer
  status:
    loadBalancer:
      ingress:
      - hostname: k8s-dnsservi-dnsservi-f9818d0d8c-2227d2f294f0b2fe.elb.eu-west-1.amazonaws.com
        ports:
        - port: 53
          protocol: UDP

`
var serviceDesired = `
apiVersion: v1
kind: Service
metadata:
  annotations:
    cross-cluster.io/owner-id: dataplane#1fhyywfw31te2v9c/coredns.mandelsoft.org/HostedZone/test/test
    dns.gardener.cloud/class: garden
    dns.gardener.cloud/dnsnames: test.test.dnsservice.dns-runtime.mandelsoft.shoot.canary.k8s-hana.ondemand.com
    dns.gardener.cloud/ttl: "60"
    service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: instance
    service.beta.kubernetes.io/aws-load-balancer-scheme: internet-facing
    service.beta.kubernetes.io/aws-load-balancer-type: external
  labels:
    app: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
    target: runtime
  name: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
  namespace: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
spec:
  loadBalancerClass: service.k8s.aws/nlb
  ports:
  - name: dns-udp-port
    port: 53
    protocol: UDP
    targetPort: 53
  selector:
    app: dns-service-test-dataplane--1fhyywfw31te2v9c-4eca2e03
  type: LoadBalancer
`

////////////////////////////////////////////////////////////////////////////////

var deploymentLive = `
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

var deploymentDesired = `
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

////////////////////////////////////////////////////////////////////////////////

var dep2Live = `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    cross-cluster.io/owner-id: demo#1fhyywfw31te2v9c/coredns.mandelsoft.org/HostedZone/test/test
    deployment.kubernetes.io/revision: "1"
    hashes.config.yaml: f1ffb5daea27525b7ebe3d5fc292fcebb1106a5207f25e5d63e5c5a0c1a0ae26
  creationTimestamp: "2026-05-11T13:36:38Z"
  generation: 1
  labels:
    app: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
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
            f:annotations:
              .: {}
              f:hashes.config.yaml: {}
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
    time: "2026-05-11T13:36:38Z"
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
    time: "2026-05-11T13:36:42Z"
  name: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
  namespace: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
  resourceVersion: "1585522"
  uid: cfeb5f49-b0e8-4f6e-8460-c51b78ed9ec7
spec:
  progressDeadlineSeconds: 600
  replicas: 1
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      annotations:
        hashes.config.yaml: f1ffb5daea27525b7ebe3d5fc292fcebb1106a5207f25e5d63e5c5a0c1a0ae26
      creationTimestamp: null
      labels:
        app: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
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
          name: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
        name: config-volume
status:
  availableReplicas: 1
  conditions:
  - lastTransitionTime: "2026-05-11T13:36:42Z"
    lastUpdateTime: "2026-05-11T13:36:42Z"
    message: Deployment has minimum availability.
    reason: MinimumReplicasAvailable
    status: "True"
    type: Available
  - lastTransitionTime: "2026-05-11T13:36:38Z"
    lastUpdateTime: "2026-05-11T13:36:42Z"
    message: ReplicaSet "dns-service-test-demo--1fhyywfw31te2v9c-238060a5-7f85bc6f6f"
      has successfully progressed.
    reason: NewReplicaSetAvailable
    status: "True"
    type: Progressing
  observedGeneration: 1
  readyReplicas: 1
  replicas: 1
  updatedReplicas: 1
`

var dep2Desired = `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    hashes.config.yaml: f1ffb5daea27525b7ebe3d5fc292fcebb1106a5207f25e5d63e5c5a0c1a0ae26
  labels:
    app: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
    target: runtime
  name: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
  namespace: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
spec:
  replicas: 1
  selector:
    matchLabels:
      app: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
  template:
    metadata:
      annotations:
        hashes.config.yaml: f1ffb5daea27525b7ebe3d5fc292fcebb1106a5207f25e5d63e5c5a0c1a0ae26
      labels:
        app: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
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
        volumeMounts:
        - mountPath: /etc/coredns/config
          name: config-volume
          readOnly: true
      volumes:
      - configMap:
          items:
          - key: Corefile
            path: Corefile
          name: dns-service-test-demo--1fhyywfw31te2v9c-238060a5
        name: config-volume
`

////////////////////////////////////////////////////////////////////////////////

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

	live := dep2Live
	desired := dep2Desired

	dec := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)

	liveObj := &unstructured.Unstructured{}
	_, _, err = dec.Decode([]byte(live), nil, liveObj)
	Error(err)
	desiredObj := &unstructured.Unstructured{}
	_, _, err = dec.Decode([]byte(desired), nil, desiredObj)
	Error(err)

	manager := "coredns.mandelsoft.org/hostedzone"

	merger, err := merge.NewObjectMerger(conv, cl.GetScheme(), manager)
	Error(err)
	p, err := merger.ComputeSSAPatch(liveObj, desiredObj)
	Error(err)
	fmt.Printf("PATCH: %s\n", string(p))
}
