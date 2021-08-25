package k8stun

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

// Tunnel manages one portforward instance
type Tunnel struct {
	id            int
	Name          string `yaml:"name"`
	Namespace     string `yaml:"namespace"`
	LabelSelector string `yaml:"labelSelector"`
	ListenPort    int    `yaml:"listenPort"`
	TargetPort    int    `yaml:"targetPort"`
	config        *rest.Config
	stopChan      chan struct{}
	readyChan     chan struct{}
	wg            *sync.WaitGroup
	running       bool
	errLog        *Logger
	outLog        *Logger
	bo            backoff.BackOff
}

// Initialize ...
func (tun Tunnel) Initialize(id int, wg *sync.WaitGroup, config *rest.Config) *Tunnel {
	t := &tun
	t.id = id
	t.outLog = &Logger{t: t, Label: "💬"}
	t.errLog = &Logger{t: t, Label: "💢"}
	t.stopChan = make(chan struct{})
	t.readyChan = make(chan struct{})
	t.wg = wg
	t.config = config
	t.bo = backoff.NewExponentialBackOff()
	return t
}

func (tun *Tunnel) run() error {
	tun.running = true
	tun.outLog.Printf("starting tunnel")

	tun.wg.Add(1)
	defer tun.wg.Done()

	// Find pod that matches selector
	ctx := context.Background()
	clientset, _ := kubernetes.NewForConfig(tun.config)
	pods, err := clientset.CoreV1().
		Pods(tun.Namespace).
		List(ctx, v1.ListOptions{
			Limit:         1,
			LabelSelector: tun.LabelSelector,
		})
	if err != nil {
		tun.errLog.Printf("error finding pod: %s", err)
		return err
	}
	if len(pods.Items) == 0 {
		tun.errLog.Printf("no pods matching selector found")
		return err
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
		return fmt.Errorf("Error creating roundtripper: %s", err)
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
		return fmt.Errorf("error creating new tunnel: %w", err)
	}

	// Start forwarder
	if err := fw.ForwardPorts(); err != nil {
		return fmt.Errorf("portforward err: %w", err)
	}

	return nil
}

// NextBackOff implements backoff.Backoff interface
func (tun *Tunnel) NextBackOff() time.Duration {
	if !tun.running {
		return backoff.Stop
	}
	return tun.bo.NextBackOff()
}

// Reset implements backoff.Backoff interface
func (tun *Tunnel) Reset() {
	tun.bo.Reset()
}

func (tun *Tunnel) runUntilStopped() {
	backoff.RetryNotify(
		tun.run,
		tun,
		func(err error, d time.Duration) {
			tun.errLog.Printf(
				"tunnel failed: %s, waiting %s",
				err, d.Truncate(time.Millisecond))
		},
	)
}

// Start ...
func (tun *Tunnel) Start() {
	go tun.runUntilStopped()
	<-tun.readyChan
}

// Stop ...
func (tun *Tunnel) Stop() {
	tun.running = false
	tun.outLog.Printf("stopping tunnel")
	defer func() {
		if r := recover(); r != nil {
			tun.errLog.Printf("recovered in %v", r)
		}
	}()
	close(tun.stopChan)
}
