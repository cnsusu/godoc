package swaggermarkdown

import (
	"log"
	"testing"
)

func TestGenerateMarkdown(t *testing.T) {
	// 需要忽略的字段列表
	var ignoredFields = map[string]bool{
		"_app_id":     true,
		"internal_id": true,
		// 可以在这里添加其他需要忽略的字段
	}

	// 在文件顶部添加自定义排序配置
	var customOrder = map[string]int{
		"/api/user/register": 1,
		"/api/user/query":    2,
	}
	swaggerMarkdown := NewSwaggerMarkdown()
	swaggerMarkdown.SetOrder(customOrder)
	swaggerMarkdown.SetIgnored(ignoredFields)
	err := swaggerMarkdown.Generate("swagger.json", "swagger.md")
	if err == nil {
		log.Println(err)
	}
}
