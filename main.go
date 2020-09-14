package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cms"
	"github.com/gin-gonic/gin"
)

type serviceStats struct {
	TotalInvocations int64                `json:"totalInvocations"`
	Services         []serviceInvocations `json:"services"`
}

type serviceInvocations struct {
	Name        string `json:"name"`
	Invocations int64  `json:"invocations"`
}

type functionDuration struct {
	Name     string  `json:"name"`
	Duration float64 `json:"duration"`
}

type serviceBillableInvocationMetric struct {
	Timestamp   int64  `json:"timestamp"`
	ServiceName string `json:"serviceName"`
	Value       float64
}

type functionAvgDurationMetric struct {
	Timestamp    int64  `json:"timestamp"`
	ServiceName  string `json:"serviceName"`
	FunctionName string `json:"functionName"`
	Value        float64
}

// service总计费调用次数, 当前月
func serviceBillableInvocations(client *cms.Client) (*serviceStats, error) {
	request := cms.CreateDescribeMetricListRequest()
	request.Scheme = "https"

	request.MetricName = "ServiceBillableInvocations"
	request.Namespace = "acs_fc"
	request.Period = "86400"
	setTimeRange(request)

	response, err := client.DescribeMetricList(request)
	if err != nil {
		return nil, err
	}

	if response.Success {
		var metricList []serviceBillableInvocationMetric
		if err := json.Unmarshal([]byte(response.Datapoints), &metricList); err != nil {
			return nil, fmt.Errorf("无法解析 DescribeMetricList(ServiceTotalInvocations) 结果, %s", err.Error())
		}

		var stats serviceStats
		var totalInvocations int64
		var servicesInvocations = make(map[string]int64)
		for _, metric := range metricList {
			total, ok := servicesInvocations[metric.ServiceName]
			if ok {
				servicesInvocations[metric.ServiceName] = total + int64(metric.Value)
			} else {
				servicesInvocations[metric.ServiceName] = int64(metric.Value)
			}

			totalInvocations = totalInvocations + int64(metric.Value)
		}

		stats.TotalInvocations = totalInvocations
		var services []serviceInvocations
		for s, t := range servicesInvocations {
			services = append(services, serviceInvocations{Name: s, Invocations: t})
		}
		stats.Services = services
		return &stats, nil
	}

	return nil, errors.New(response.Message)
}

// function平均执行时间
func functionAvgDuration(client *cms.Client) ([]functionDuration, error) {
	request := cms.CreateDescribeMetricListRequest()
	request.Scheme = "https"

	request.MetricName = "FunctionAvgDuration"
	request.Namespace = "acs_fc"
	request.Period = "86400"
	setTimeRange(request)

	response, err := client.DescribeMetricList(request)
	if err != nil {
		return nil, err
	}

	if response.Success {
		var metricList []functionAvgDurationMetric
		if err := json.Unmarshal([]byte(response.Datapoints), &metricList); err != nil {
			return nil, fmt.Errorf("无法解析 DescribeMetricList(FunctionAvgDuration) 结果, %s", err.Error())
		}

		var functionsDurations = make(map[string][]float64)

		for _, metric := range metricList {
			name := fmt.Sprintf("%s/%s", metric.ServiceName, metric.FunctionName)
			durations, ok := functionsDurations[name]
			if ok {
				functionsDurations[name] = append(durations, metric.Value)
			} else {
				functionsDurations[name] = []float64{metric.Value}
			}
		}

		var stats []functionDuration
		for name, durations := range functionsDurations {
			var sum float64
			for _, d := range durations {
				sum += d
			}

			stats = append(stats, functionDuration{
				Name:     name,
				Duration: sum / float64(len(durations)),
			})
		}

		return stats, nil
	}

	return nil, errors.New(response.Message)
}

func main() {
	r := gin.New()
	r.Use(gin.Recovery())

	setupRouter(r)

	start(&http.Server{
		Addr:    fmt.Sprintf(":%s", getenv("FC_SERVER_PORT", "9000")),
		Handler: r,
	})
}

func setupRouter(r *gin.Engine) {
	client, clientErr := cms.NewClientWithAccessKey(os.Getenv("REGION_ID"), os.Getenv("ACCESS_KEY_ID"), os.Getenv("ACCESS_KEY_SECRET"))

	rg := r.Group("/stats")
	rg.GET("/service", func(c *gin.Context) {
		if clientErr != nil {
			c.JSON(http.StatusInternalServerError, failed(fmt.Sprintf("CMS客户端初始化错误, %s", clientErr.Error())))
			return
		}

		stats, err := serviceBillableInvocations(client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, failed(err.Error()))
			return
		}

		c.JSON(http.StatusOK, data(stats))
	})
	rg.GET("/function", func(c *gin.Context) {
		if clientErr != nil {
			c.JSON(http.StatusInternalServerError, failed(fmt.Sprintf("CMS客户端初始化错误, %s", clientErr.Error())))
			return
		}

		stats, err := functionAvgDuration(client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, failed(err.Error()))
			return
		}

		c.JSON(http.StatusOK, data(stats))
	})
}

func start(srv *http.Server) {
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen: %s\n", err)
		}
	}()

	log.Printf("Start Server @ %s", srv.Addr)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Print("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server Shutdown:%s", err)
	}
	<-ctx.Done()
	log.Print("Server exiting")
}

func getenv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func failed(msg string) gin.H {
	return gin.H{
		"msg":       msg,
		"timestamp": time.Now().Unix(),
	}
}

func data(data interface{}) gin.H {
	return gin.H{
		"msg":       "success",
		"data":      data,
		"timestamp": time.Now().Unix(),
	}
}

func setTimeRange(request *cms.DescribeMetricListRequest) {
	t := time.Now()
	startDate := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	startTime := fmt.Sprintf("%s 00:00:01", startDate.Format("2006-01-02"))
	request.StartTime = startTime
	endTime := fmt.Sprintf("%s 23:59:59", startDate.AddDate(0, 1, 0).Add(-time.Nanosecond).Format("2006-01-02"))
	request.EndTime = endTime
}
