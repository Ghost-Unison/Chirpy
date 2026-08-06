package main

import (
	"log"
	"net/http"
)

/*

Handler 是一个接口契约（有 ServeHTTP 方法的都是 Handler）。
mux.Handle 的第二个参数是 Handler，mux 自己也是 Handler——这不矛盾，因为 ServeMux 实现了 ServeHTTP，
它作为 Handler 的"任务"就是把请求分发给注册在它身上的其他 Handler。

首先任何实现了ServeHTTP方法的类型都是Handler
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}

*/

func main() {
	const filepathRoot = "."
	const port = "8080"

	// handler to route request
	mux := http.NewServeMux() // mux 是 *ServeMux 类型
	// 第二个参数 http.StripPrefix(...) 返回的也是 Handler
	// 因为 StripPrefix 返回的 *http.appendsSlash... 等等，都实现了 ServeHTTP
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot))))
	// healthCheckFunc 本身是普通函数，func(ResponseWriter, *Request) 签名
	// HandlerFunc 是一个定义类型，它有 ServeHTTP 方法（方法体里就是调用自身这个函数）
	// 所以 http.HandlerFunc(healthCheckFunc) "是" Handler
	// HandleFunc 内部就是这么转换的：mux.Handle("/healthz", HandlerFunc(healthCheckFunc))
	/*

		普通的 func healthCheckFunc(...) 怎么也成了 Handler？因为 Go 定义了：
		type HandlerFunc func(ResponseWriter, *Request)

		func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
		    f(w, r)  // 直接调用自己
		}

		任何签名是 func(ResponseWriter, *Request) 的函数，都可以转成 HandlerFunc，而 HandlerFunc 有 ServeHTTP 方法，
		所以它就"是"Handler。这就是为什么 mux.HandleFunc 能接受普通函数——它内部帮你做了这层转换。
	*/
	mux.HandleFunc("/healthz", healthCheckFunc)

	//create new httpServer
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe()) //start the server

}

/*
1./healthz 天然支持所有 HTTP 方法：Go 的 ServeMux 用 "/healthz" 这种不带方法的 pattern 注册时，GET、POST 等任何方法都会匹配，
正好满足课程 "using any HTTP method" 的要求。

2./app/ 结尾的斜杠：mux.Handle("/app/", ...) 是子树匹配，能匹配 /app/ 下的所有路径；如果写成 /app（无斜杠）就只能匹配 /app 本身。
*/

/*
Readiness endpoint
The endpoint should be accessible at the /healthz path using any HTTP method.

The endpoint should simply return a 200 OK status code indicating that it has started up successfully and is listening for traffic.
The endpoint should return a Content-Type: text/plain; charset=utf-8 header,
and the body will contain a message that simply says "OK" (the text associated with the 200 status code).
*/
func healthCheckFunc(w http.ResponseWriter, r *http.Request) {
	//先设置header再调用writeHeader
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	//w.Write([]byte("OK"))
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

/*
/healthz 精确匹配
/healthz ✓
/healthz/ ✗（多了个斜杠，不是同一个路径）
/healthz/foo ✗

/app/ 子树匹配
/app/ ✓
/app/index.html ✓
/app/css/style.css ✓

补充：匹配优先级
当多个 pattern 都能匹配同一个请求时，更具体（更长）的 pattern 获胜。
这就是为什么即使注册了 /，请求 /healthz 仍然会进你的 handler 而不是 fileserver——/healthz 比 / 更具体。
/healthz/ 的情况恰恰相反：它只够得着 / 这一个 pattern，没得选。

一句话总结：/healthz 是"点名"，/app/ 是"承包一片区域"，而 / 承包的区域恰好是全世界。
*/

/*

HTTP 请求到来
    ↓
http.Server 调用它的 Handler.ServeHTTP(w, r)
    ↓  (这里的 Handler 是 mux)
mux.ServeHTTP(w, r) 被调用
    ↓  (mux 内部：根据路径查表，找到注册的那个 Handler)
找到 /healthz → 调用 healthCheckFunc 的 ServeHTTP(w, r)
找到 /app/   → 调用 FileServer 的 ServeHTTP(w, r)

所以 mux 本质上是一个分发器(dispatcher)——它实现了 Handler 接口，但它干的活是"把请求转交给其他 Handler"。这是一种经典的组合模式(Composite Pattern)。

*/

/*
可以嵌套——因为 ServeMux 是 Handler，所以你可以把一个 ServeMux 注册到另一个 ServeMux 里：
apiMux := http.NewServeMux()
apiMux.HandleFunc("/users", usersHandler)
apiMux.HandleFunc("/orders", ordersHandler)

rootMux := http.NewServeMux()
rootMux.Handle("/api/", apiMux)   // ← apiMux 作为 rootMux 的 Handler
rootMux.Handle("/app/", fileServer)

// 请求 /api/users → rootMux 发现匹配 /api/ → 转给 apiMux
//                  → apiMux 发现匹配 /api/users → 转给 usersHandler
这就是"嵌套 Handler"——如果没有"ServeMux 也是 Handler"这个设计，这种组合根本做不到。


*/
