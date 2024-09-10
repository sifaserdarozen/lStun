package main

import (
	"log"
	"net/http"

	"github.com/prometheus/common/expfmt"
)

func main() {
	resp, err := http.Get("http://localhost:8081/metrics")
	if err != nil {
		log.Fatalln(err)
	}

	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(resp.Body)
	if err != nil {
		log.Println(err)
	}

	uptimeSec := mf["lstun_uptime_sec"]
	buildInfo := mf["lstun_build_info"]

	log.Printf("uptime: %v", uptimeSec)

	uptimeMetric := uptimeSec.GetMetric()[0].GetGauge().GetValue()
	log.Println(uptimeMetric)

	buildLabel := buildInfo.GetMetric()[0].GetLabel()
	for _, v := range buildLabel {
		log.Printf("%s : %s", v.GetName(), v.GetValue())
	}

	log.Printf("buildInfo: %v", buildInfo)

	//for k, v := range mf {
	//    log.Printf("%v ---> %v", k, v)
	//}
}
