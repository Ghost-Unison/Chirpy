package main

import "net/http"

// 检查服务是否可用
func healthCheckFunc(w http.ResponseWriter, r *http.Request) {
	//先设置header再调用writeHeader
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte("OK"))
	w.Write([]byte(http.StatusText(http.StatusOK)))
}
