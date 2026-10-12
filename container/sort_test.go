package container

import (
	"testing"

	"github.com/bcicen/ctop/config"
	"github.com/bcicen/ctop/models"
)

func TestContainersFilter(t *testing.T) {
	config.Init()
	config.Update("allContainers", "true")

	c1 := &Container{Id: "abcdef123456789", Meta: models.NewMeta("id", "abcdef123456")}
	c1.Meta["name"] = "web-server"
	c1.Meta["image"] = "nginx:alpine"
	c1.Meta["state"] = "running"

	c2 := &Container{Id: "987654321fedcba", Meta: models.NewMeta("id", "987654321fed")}
	c2.Meta["name"] = "db-redis"
	c2.Meta["image"] = "redis:7.0"
	c2.Meta["state"] = "running"

	list := Containers{c1, c2}

	// 测试 1: 按名称忽略大小写匹配
	config.Update("filterStr", "WEB")
	list.Filter()
	if !c1.Display || c2.Display {
		t.Errorf("expected c1 true, c2 false for 'WEB'; got c1=%v, c2=%v", c1.Display, c2.Display)
	}

	// 测试 2: 按镜像名匹配
	config.Update("filterStr", "redis")
	list.Filter()
	if c1.Display || !c2.Display {
		t.Errorf("expected c1 false, c2 true for 'redis'; got c1=%v, c2=%v", c1.Display, c2.Display)
	}

	// 测试 3: 按短 ID 匹配
	config.Update("filterStr", "abcdef")
	list.Filter()
	if !c1.Display || c2.Display {
		t.Errorf("expected c1 true, c2 false for 'abcdef'; got c1=%v, c2=%v", c1.Display, c2.Display)
	}

	// 测试 4: 特殊正则字符不应引发崩溃
	config.Update("filterStr", "nginx[")
	list.Filter()
	if c1.Display || c2.Display {
		t.Errorf("expected neither to match 'nginx['; got c1=%v, c2=%v", c1.Display, c2.Display)
	}
}
