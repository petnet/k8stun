package k8stun

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// Tunnel manages one portforward instance
type Tunnel struct {
	Name          string
	Namespace     string
	LabelSelector string
	// ListenAddress string
	ListenPort int
	TargetPort int
	config     *rest.Config
	stopChan   chan struct{}
	readyChan  chan struct{}
	wg         *sync.WaitGroup
	running    bool
	errLog     *Logger
	outLog     *Logger
}

// Initialize ...
func (tun Tunnel) Initialize(wg *sync.WaitGroup, config *rest.Config) *Tunnel {
	t := &tun
	t.outLog = &Logger{t: t, Label: "💬"}
	t.errLog = &Logger{t: t, Label: "💢"}
	t.stopChan = make(chan struct{})
	t.readyChan = make(chan struct{})
	t.wg = wg
	t.config = config
	return t
}

func (tun *Tunnel) run() {
	tun.wg.Add(1)
	defer tun.wg.Done()

	// Find pod that matches selector
	ctx := context.Background()
	clientset, _ := kubernetes.NewForConfig(tun.config)
	pods, err := clientset.CoreV1().Pods("").List(ctx, v1.ListOptions{
		Limit:         1,
		LabelSelector: tun.LabelSelector,
	})
	if err != nil {
		tun.errLog.Printf("error finding pod: %s", err)
		return
	}
	if len(pods.Items) == 0 {
		tun.errLog.Printf("no pods matching selector found")
		return
	}
	podName := pods.Items[0].Name
	tun.outLog.Printf("creating tunnel for pod %s", podName)

	// Kubernetes API connection for port forward
	host := strings.TrimLeft(tun.config.Host, "https://")
	path := fmt.Sprintf(
		"/api/v1/namespaces/%s/pods/%s/portforward",
		tun.Namespace, podName,
	)
	transport, upgrader, err := spdy.RoundTripperFor(tun.config)
	if err != nil {
		tun.errLog.Printf("Error creating roundtripper: %s", err)
		return
	}
	dialer := spdy.NewDialer(upgrader,
		&http.Client{Transport: transport}, http.MethodPost,
		&url.URL{Scheme: "https", Path: path, Host: host},
	)

	// Create forwarder
	fw, err := portforward.New(dialer,
		[]string{fmt.Sprintf("%d:%d", tun.ListenPort, tun.TargetPort)},
		tun.stopChan, tun.readyChan,
		tun.outLog, tun.errLog)
	if err != nil {
		tun.errLog.Printf("error creating new tunnel: %s", err)
		return
	}

	// Start forwarder
	if err := fw.ForwardPorts(); err != nil {
		tun.errLog.Printf("portforward err: %s", err)
	}
}

// Start ...
func (tun *Tunnel) Start() {
	tun.outLog.Printf("starting")
	go tun.run()
	<-tun.readyChan
}

// Stop ...
func (tun *Tunnel) Stop() {
	tun.outLog.Printf("stopping")
	defer func() {
		if r := recover(); r != nil {
			tun.errLog.Printf("recovered in %v", r)
		}
	}()
	close(tun.stopChan)
}
