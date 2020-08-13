package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	flag "github.com/spf13/pflag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type ReceiverFunc func(key string, value float64)

func (receiver ReceiverFunc) Receive(key string, value float64) {
	receiver(key, value)
}

type Receiver interface {
	Receive(key string, value float64)
}

func WalkJSON(path string, jsonData interface{}, receiver Receiver) {
	switch v := jsonData.(type) {
	case int:
		receiver.Receive(path, float64(v))
	case float64:
		receiver.Receive(path, v)
	case bool:
		n := 0.0
		if v {
			n = 1.0
		}
		receiver.Receive(path, n)
	case string:
		// ignore
	case nil:
		// ignore
	case []interface{}:
		prefix := path + "__"
		for i, x := range v {
			WalkJSON(fmt.Sprintf("%s%d", prefix, i), x, receiver)
		}
	case map[string]interface{}:
		prefix := ""
		if path != "" {
			prefix = path + "::"
		}
		for k, x := range v {
			WalkJSON(fmt.Sprintf("%s%s", prefix, k), x, receiver)
		}
	default:
		log.Printf("unkown type: %#v", v)
	}
}

func doProbe(client *http.Client, target string) (interface{}, error) {
	resp, err := client.Get(target)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var jsonData interface{}
	err = json.Unmarshal([]byte(bytes), &jsonData)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

var httpClient *http.Client

func init() {
	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConns: 100,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
}

func probeHandler(target string, prefix string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Probing", target, prefix)
		jsonData, err := doProbe(httpClient, target)
		if err != nil {
			log.Print(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		registry := prometheus.NewRegistry()

		WalkJSON("", jsonData, ReceiverFunc(func(key string, value float64) {
			g := prometheus.NewGauge(
				prometheus.GaugeOpts{
					Name: prefix + key,
					Help: "Retrieved value",
				},
			)
			registry.MustRegister(g)
			g.Set(value)
		}))

		h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}

var indexHTML = []byte(`<html>
<head><title>Json Exporter</title></head>
<body>
<h1>Json Exporter</h1>
<p><a href="/metrics">Metrics</a></p>
</body>
</html>`)

func main() {
	// Include source line informatino in logs
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Parse flags
	addr := flag.StringP("listen-address", "a", ":9116", "The address to listen on for HTTP requests.")
	prefix := flag.StringP("prefix", "p", "", "Prefix to add to parsed metric names.")
	flag.Parse()

	// Parse command line arguments
	args := flag.Args()
	expectedArgs := 1
	if len(args) < expectedArgs {
		log.Printf("Not enough arguments. Expected %d, got %d.\n", expectedArgs, len(args))
		log.Printf("USAGE: %s <target> [flags]", os.Args[0])
		flag.Usage()
		os.Exit(1)
	} else if len(args) > expectedArgs {
		log.Printf("Warning: too many arguments received. Expected %d, got %d: %v\n", expectedArgs, len(args), args)
	}
	// Set target
	target := args[0]

	// Set handlers
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write(indexHTML)
	})
	http.HandleFunc("/metrics", probeHandler(target, *prefix))

	// Start the server
	log.Printf("Listening on %s.", *addr)
	log.Printf("Probing %s.", target)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
