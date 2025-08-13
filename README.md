<p align="right">
   <strong>中文</strong> | <a href="./README.en.md">English</a>
</p>

# Swagger 转 Markdown 工具 (Go 实现)

[![Go 代码质量](https://goreportcard.com/badge/github.com/cnsusu/swaggermarkdown)](https://goreportcard.com/report/github.com/cnsusu/swaggermarkdown)
[![许可证: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

> 轻量级、无依赖的 Go 库，将 Swagger/OpenAPI JSON 规范转换为结构清晰的 Markdown 文档，支持自定义格式化选项。特别适合静态站点部署和离线使用。

## ✨ 功能特性
- **动态排序控制** - 通过优先级映射自定义接口显示顺序
- **字段过滤** - 排除敏感或不必要字段（如 `_app_id`）
- **单文件输出** - 生成整合的 Markdown 文件
- **零依赖** - 纯 Go 实现，无外部依赖
- **支持 Swagger 2.0 & OpenAPI 3.0** - 完整 Markdown 格式支持（含表格、代码块和链接）:cite[5]

## 📥 安装方式
```bash
go get github.com/cnsusu/swaggermarkdown
```

## 🚀 Basic Usage
```bash
package main

import (
	"log"
	swaggermarkdown "github.com/cnsusu/swaggermarkdown"
)

func main() {
	// 定义需要忽略的字段
	ignoredFields := map[string]bool{
		"_app_id":    true,  // 敏感应用ID
		"internal_id": true, // 在此添加其他字段
	}

	// 配置接口显示顺序（数值越小越靠前）
	customOrder := map[string]int{
		"/api/user/register": 1, // 最高优先级
		"/api/user/login":    2,
	}

	// 初始化生成器
	swaggerMarkdown := swaggermarkdown.NewSwaggerMarkdown()
	swaggerMarkdown.SetOrder(customOrder)
	swaggerMarkdown.SetIgnored(ignoredFields)
	swaggerMarkdown.SetTitle("我的API文档") // 自定义标题

	// 生成文档
	err := swaggerMarkdown.Generate("swagger.json", "API文档.md")
	if err != nil {
		log.Fatal("文档生成失败: ", err)
	}
}
```

