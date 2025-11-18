//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func fanIn(chs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(chs))
	for _, c := range chs {
		go func(c <-chan int) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func slowChan() <-chan int {
	fmt.Println("slowChan() called, will sleep 4s...")
	time.Sleep(4 * time.Second)
	ch := make(chan int, 1)
	ch <- 42
	fmt.Println("slowChan() returns channel with data")
	return ch
}

func main() {
	// 非阻塞接收，注意这个如果加上for循序，会造成忙等
	ch1 := make(chan int)
	select {
	case v := <-ch1:
		fmt.Println("received from ch1:", v)
	default:
		fmt.Println("no value received from ch1")
	}
	// 注意：下面这种写法会导致 CPU 占用过高，因为它会不断地轮询 channel
	//     for {
	//     select {
	//     case v := <-ch:
	//         fmt.Println(v)
	//     default:
	//         // 立即返回，导致 CPU 占用高
	//     }
	// }
	// 修改为使用 time.After 来避免忙等，阻塞式等待
	// 	for {
	//     select {
	//     case v := <-ch:
	//         fmt.Println(v)
	//     case <-time.After(100 * time.Millisecond): // 这里注意，高频率可能会造成timer泄漏
	//         // 批量处理或休眠
	//     }
	// }
	// 推荐使用可重用的 Timer 来避免频繁分配
	// 	t := time.NewTimer(time.Second)
	// defer t.Stop()
	// for {
	//     // 在每轮开始前重置（先 Stop 并排空可能的残留）
	//     if !t.Stop() {
	//         select { case <-t.C: default: }
	//     }
	//     t.Reset(time.Second)

	//     select {
	//     case <-ch:
	//         // ...
	//     case <-t.C:
	//         // timeout
	//     }
	// }

	// 超时接收
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case v := <-ch1:
		fmt.Println("received from ch1:", v)
	case <-timer.C:
		fmt.Println("timeout waiting for value from ch1")
	}

	// 使用 context 控制超时
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()
	ch2 := make(chan string)

	select {
	case v := <-ch2:
		fmt.Println("data", v)
	case <-ctx2.Done():
		fmt.Println("cancelled or timed out:", ctx2.Err())
	}

	// 动态控制chan的接收
	var sendCh chan int // 默认为nil
	readyToSend := true // 根据某些条件动态决定
	if readyToSend {
		sendCh = make(chan int, 1)
	}

	select {
	case sendCh <- 42:
		fmt.Println("sent 42 to sendCh")
	case <-time.After(time.Second):
		fmt.Println("sendCh is not ready for sending")
	}

	sendCh = nil // 之后将其设为nil，禁用发送
	select {
	case sendCh <- 42:
		fmt.Println("sent 42 to sendCh")
	case <-time.After(time.Second):
		fmt.Println("sendCh is not ready for sending")
	}

	// 合并多个生产者
	producer1 := make(chan int)
	producer2 := make(chan int)

	go func() {
		for i := 0; i < 5; i++ {
			producer1 <- i
			time.Sleep(500 * time.Millisecond)
		}
		close(producer1)
	}()

	go func() {
		for i := 100; i < 105; i++ {
			producer2 <- i
			time.Sleep(700 * time.Millisecond)
		}
		close(producer2)
	}()

	merged := fanIn(producer1, producer2)
	for v := range merged {
		fmt.Println("received:", v)
	}
	// 重点，select的分值执行顺序
	// 1. select会获取监听的channel的表达式
	// 2. 在获取监听的channel的时候，是从上到下顺序执行的，没有并发
	// 3. 等到都获取完监听的channel表达式后，才会去判断哪个channel有数据准备好
	// 4. 会伪随机的选择一个有数据的channel进行处理，执行完获取channel表达式，刚好有数据的会优先执行
	// 5. 如果某个channel的表达式执行很慢，会导致后续channel的表达式获取也被阻塞，影响整体的select性能

	fmt.Println("--- select 慢函数与 timer 演示 ---")
	start := time.Now()
	select {
	case v := <-slowChan(): // slowChan() 会被调用获取channel，执行完刚好有数据
		fmt.Printf("select got slowChan value: %d, elapsed: %.2fs\n", v, time.Since(start).Seconds())
	case <-time.After(2 * time.Second): // slowChan() 执行完毕前，time.After 先返回channel,但是slowChan已经有数据，无法触发
		fmt.Printf("select timeout, elapsed: %.2fs\n", time.Since(start).Seconds())
	}
	fmt.Println("--- end ---")

	// 另一个例子，timer有一定概率执行，但是还是slowChan优先触发
	timer2 := time.NewTimer(2 * time.Second) // 定时器已经启动
	defer timer2.Stop()
	select {
	case v := <-slowChan(): // slowChan() 必定被执行，获取channel,但是定时器已经启动，slowChan时间大于timer2的超时
		fmt.Println(v)
	case <-timer2.C: // 获取定时器channel
		fmt.Println("timeout")
	}

	// 要点

}
