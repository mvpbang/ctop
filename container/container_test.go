package container

import (
	"sync"
	"testing"
	"time"

	"github.com/bcicen/ctop/connector/collector"
	"github.com/bcicen/ctop/connector/manager"
	"github.com/bcicen/ctop/models"
)

func TestContainerConcurrentAccess(t *testing.T) {
	c := New("test12345678", collector.NewMock(1), manager.NewMock())

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// 模拟写入 Meta
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				c.SetMeta("name", "container-test")
				c.SetMeta("state", "running")
			}
		}
	}()

	// 模拟读取 Meta
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_ = c.GetMeta("name")
				_ = c.GetMeta("state")
			}
		}
	}()

	// 模拟写入 Metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				m := models.Metrics{CPUUtil: 50, MemUsage: 1024}
				c.SetMetrics(m)
			}
		}
	}()

	// 模拟读取 Metrics
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stop:
				return
			default:
				_ = c.GetMetrics().CPUUtil
			}
		}
	}()

	time.Sleep(100 * time.Millisecond)
	close(stop)
	wg.Wait()
}
