//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// 链路追踪
type traceIDKey struct{}

func HandRequest(ctx context.Context) {
	traceID := ctx.Value(traceIDKey{}).(string)
	log.Println("trace:", traceID)
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	select {
	case <-time.After(3 * time.Second):
		w.Write([]byte("done"))
	case <-ctx.Done():
		fmt.Println("client canceled:", ctx.Err())
		return
	}
}

func worker(ctx context.Context, jobs <-chan int) {
	for {
		select {
		case <-ctx.Done():
			return
		case j := <-jobs:
			fmt.Println("process", j)
		}
	}
}

func worker1(ctx context.Context, id int, jobs <-chan int, results chan<- int) {
	for {
		select {
		case <-ctx.Done():
			// 收到统一取消信号，优雅退出
			fmt.Printf("worker %d exit: %v\n", id, ctx.Err())
			return
		case j, ok := <-jobs:
			if !ok {
				// jobs 关了，正常处理完退出
				fmt.Printf("worker %d: jobs closed, exit\n", id)
				return
			}
			// 正常处理任务
			fmt.Printf("worker %d processing job %d\n", id, j)
			results <- j * 2
		}
	}
}

func main() {
	// 取消goroutine(避免goroutine泄露)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("done")
				return
			default:
				fmt.Println("running...")
				time.Sleep(time.Second)
			}
		}
	}()
	time.Sleep(time.Second * 3)
	cancel()

	// 超时控制，防止服务 卡死
	ctx1, cancel1 := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel1()
	select {
	case <-time.After(3 * time.Second):
		fmt.Println("operation done")
	case <-ctx1.Done():
		fmt.Println("timeout", ctx.Err())
	}

	// 请求范围内的数据传递（链路追踪）
	// 用途：

	// 传递用户信息

	// 传递 trace_id、span_id（链路追踪）

	// 传递请求范围的参数

	// ⚠️ 注意：Context 不用于存业务数据，值应该是 metadata。
	ctx2 := context.WithValue(context.Background(), traceIDKey{}, "abc123")
	HandRequest(ctx2)

	// 在 Go 的 http.Server 内：

	// 请求开始时自动创建 context

	// 请求结束时自动 cancel

	// handler 使用 ctx 来控制数据库访问、API 调用等
	// httpHandler()
	// 作为 goroutine 之间的“退出信号”
	ctx3, cancel3 := context.WithCancel(context.Background())
	defer cancel3()
	jobs := make(chan int)
	go worker(ctx3, jobs)
	jobs <- 1
	jobs <- 2
	cancel3()

	// 控制子任务的层级取消
	parent := context.Background()
	ctxA, cancelA := context.WithCancel(parent)
	context.WithCancel(parent)
	_, cancleA1 := context.WithCancel(ctxA)
	cancleA1() // 子级取消，不会影响ctxA
	cancelA()  // 子级全部受影响

	// 竞速， 速度快的cancel,速度慢的取消
	ctx4, cancel4 := context.WithCancel(context.Background())
	defer cancel4()

	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("task1 done")
		cancel4() // 结束其他任务
	}()

	go func() {
		<-ctx4.Done()
		fmt.Println("task2 canceled")
	}()

	//结合 WaitGroup 的标准用法
	var wg sync.WaitGroup
	ctx5, cancel5 := context.WithCancel(context.Background())

	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for {
				select {
				case <-ctx5.Done():
					fmt.Println("exit:", i)
					return
				default:
					time.Sleep(200 * time.Millisecond)
				}
			}
		}(i)
	}

	time.Sleep(time.Second)
	cancel5() // 通知全部 goroutine 退出

	wg.Wait() // 等全部退出

	// workpool 用法
	ctx6, cancel6 := context.WithCancel(context.Background())
	defer cancel6() // 程序退出时兜底取消

	jobs1 := make(chan int, 100)
	results := make(chan int, 100)

	// 启动 5 个 worker
	for w := 1; w <= 5; w++ {
		go worker1(ctx6, w, jobs1, results)
	}

	// 投递任务（这里先简单全部投入）
	for j := 1; j <= 100; j++ {
		jobs <- j
	}
	close(jobs)

}
