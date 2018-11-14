package main

import (
	"context"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/aws/aws-xray-sdk-go/plugins/ec2"
	_ "github.com/aws/aws-xray-sdk-go/plugins/ecs"
	"github.com/aws/aws-xray-sdk-go/xray"
)

const appName = "eks-workshop-x-ray-sample"

func init() {
	xray.Configure(xray.Config{
		DaemonAddr:     "xray-service.default:2000",
		LogLevel:       "info",
		ServiceVersion: "1.2.3",
	})
}

func main() {
	http.Handle("/", xray.Handler(xray.NewFixedSegmentNamer(appName), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		res := &response{Message: "42 - The Answer to the Ultimate Question of Life, The Universe, and Everything.", Random: []int{}}

		count := time.Now().Second()
		gen := random(res)

		ctx := context.Background()
		ctx, seg := xray.BeginSegment(ctx, appName + "-gen")
		defer seg.Close(nil)

		xray.Capture(ctx, "random", func(ctx1 context.Context) (err error) {
			for i := 0; i < count; i++ {
				gen()
			}
			return
		})

		out, _ := json.Marshal(res)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, string(out))

	})))
	http.ListenAndServe(":8080", nil)
}

type response struct {
	Message string `json:"message"`
	Random  []int  `json:"random"`
}

func random(res *response) func() {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return func() {
		res.Random = append(res.Random, r.Intn(42))
	}
}
