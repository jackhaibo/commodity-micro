package queue

import "github.com/common/logger"

var (
	GoroutinePool    *Pool
	preGoroutinePool *Pool
)

type Pool struct {
	size   int
	queue  chan func() error
	result chan error
}

// 初始化
func Init(size int) {
	if GoroutinePool != nil {
		preGoroutinePool = GoroutinePool
		defer preGoroutinePool.Stop()
	}
	GoroutinePool = &Pool{
		size:   size,
		queue:  make(chan func() error, size),
		result: make(chan error, size),
	}
	GoroutinePool.Start()
}

// 开门接客
func (p *Pool) Start() {
	// 开启Number个goroutine
	for i := 0; i < p.size; i++ {
		go func() {
			for {
				task, ok := <-p.queue
				if !ok {
					logger.Debug("goroutine exit")
					return
				}
				err := task()
				p.result <- err
			}
		}()
	}
}

// 关门送客
func (p *Pool) Stop() {
	close(p.queue)
	close(p.result)
}

// 添加任务
func (p *Pool) AddTask(task func() error) {
	p.queue <- task
}

func (p *Pool) GetResult() chan error {
	return p.result
}
