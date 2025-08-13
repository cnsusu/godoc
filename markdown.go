package swaggermarkdown

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
)

// Swagger 结构体定义
type Swagger struct {
	Swagger     string                 `json:"swagger"`
	Info        Info                   `json:"info"`
	Paths       map[string]interface{} `json:"paths"`
	Definitions map[string]Definition  `json:"definitions"`
}

type Info struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type Definition struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
	AllOf      []AllOfItem         `json:"allOf"` // 添加 allOf 支持
}

type AllOfItem struct {
	Ref        string              `json:"$ref"`
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
}

type Property struct {
	Type        string      `json:"type"`
	Format      string      `json:"format"`
	Description string      `json:"description"`
	Ref         string      `json:"$ref"`
	Items       *Items      `json:"items"`
	Properties  interface{} `json:"properties"` // 用于处理嵌套对象
	AllOf       []AllOfItem `json:"allOf"`      // 添加 allOf 支持
}

type Items struct {
	Ref string `json:"$ref"`
}

type Operation struct {
	Summary     string                   `json:"summary"`
	Description string                   `json:"description"`
	Parameters  []Parameter              `json:"parameters"`
	Responses   map[string]ResponseValue `json:"responses"`
	Tags        []string                 `json:"tags"`
}

type Parameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Type        string `json:"type"`
	Schema      Schema `json:"schema"`
}

type Schema struct {
	Ref   string `json:"$ref"`
	Type  string `json:"type"`
	Items *Items `json:"items"`
	AllOf []struct {
		Ref string `json:"$ref"`
	} `json:"allOf"` // 添加 allOf 支持
}

type ResponseValue struct {
	Description string `json:"description"`
	Schema      Schema `json:"schema"`
}

type ResponseSchema struct {
	AllOf []struct {
		Ref        string            `json:"$ref"`
		Type       string            `json:"type"`
		Properties map[string]Schema `json:"properties"`
	} `json:"allOf"`
	Ref   string `json:"$ref"`
	Type  string `json:"type"`
	Items *Items `json:"items"`
}

type SwaggerMarkdown struct {
	mu            sync.Mutex
	CustomOrder   map[string]int
	IgnoredFields map[string]bool
}

// 使用前初始化
func NewSwaggerMarkdown() *SwaggerMarkdown {
	return &SwaggerMarkdown{
		CustomOrder:   make(map[string]int),
		IgnoredFields: make(map[string]bool),
	}
}

// SetOrder 设置排序
func (s *SwaggerMarkdown) SetOrder(param map[string]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CustomOrder = param
}

// SetIgnored 设置忽略字段
func (s *SwaggerMarkdown) SetIgnored(param map[string]bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.IgnoredFields = param
}

// Generate 读取 swagger.json，生成 Markdown 文档
func (s *SwaggerMarkdown) Generate(swaggerFilePath, markdownFilePath string) error {
	// 读取 swagger.json 文件
	data, err := os.ReadFile(swaggerFilePath)
	if err != nil {
		return err
	}

	// 解析 JSON
	var swagger Swagger
	err = json.Unmarshal(data, &swagger)
	if err != nil {
		return err
	}

	// 创建并打开 Markdown 文件
	file, err := os.Create(markdownFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 写入文档头部
	s.writeHeader(file, &swagger)

	// 处理路径和操作
	s.writePaths(file, swagger.Paths, swagger.Definitions)

	fmt.Println("Markdown 文档已生成: ", markdownFilePath)
	return nil
}

func (s *SwaggerMarkdown) writeHeader(file *os.File, swagger *Swagger) {
	header := fmt.Sprintf(`# %s

**版本**: %s  
**描述**: %s

## 接口概览
`, swagger.Info.Title, swagger.Info.Version, swagger.Info.Description)
	file.WriteString(header)
}

func (s *SwaggerMarkdown) writePaths(file *os.File, paths map[string]interface{}, defs map[string]Definition) {
	file.WriteString("\n## 接口详情\n")

	// 将路径转换为可排序的切片
	type pathItem struct {
		path   string
		method string
		detail interface{}
	}

	var pathList []pathItem
	for path, methods := range paths {
		methodMap := methods.(map[string]interface{})
		for method, details := range methodMap {
			pathList = append(pathList, pathItem{
				path:   path,
				method: method,
				detail: details,
			})
		}
	}

	// 自定义排序
	sort.Slice(pathList, func(i, j int) bool {
		// 获取自定义顺序，如果没有定义则按字母顺序
		orderI, okI := s.CustomOrder[pathList[i].path]
		orderJ, okJ := s.CustomOrder[pathList[j].path]

		if okI && okJ {
			return orderI < orderJ
		}
		if okI {
			return true
		}
		if okJ {
			return false
		}
		return pathList[i].path < pathList[j].path
	})

	// 按排序后的顺序处理接口
	for _, item := range pathList {
		// 解析操作详情
		opData, _ := json.Marshal(item.detail)
		var op Operation
		json.Unmarshal(opData, &op)

		// 写入接口标题
		file.WriteString(fmt.Sprintf("### %s\n\n", op.Summary))
		file.WriteString(fmt.Sprintf("**接口地址**：`%s`  \n", item.path))
		file.WriteString(fmt.Sprintf("**请求方式**：`%s`  \n", strings.ToUpper(item.method)))
		if strings.TrimSpace(op.Description) != "" {
			file.WriteString(fmt.Sprintf("**接口描述**：%s  \n", op.Description))
		}
		// 写入参数表
		if len(op.Parameters) > 0 {
			file.WriteString("\n**请求参数:**\n\n")
			file.WriteString("| 名称 | 类型 | 是否必填 | 描述 |\n")
			file.WriteString("|------|------|----------|------|\n")

			for _, param := range op.Parameters {
				// 处理引用类型参数
				if param.Schema.Ref != "" {
					refName := s.extractRefName(param.Schema.Ref)
					if def, exists := defs[refName]; exists {
						// 深度解析引用类型中的字段
						for field, prop := range def.Properties {
							// 检查是否需要忽略该字段
							if s.IgnoredFields[field] {
								continue
							}

							fieldType := s.getPropertyType(prop)
							required := "否"
							if s.contains(def.Required, field) {
								required = "是"
							}
							file.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s |\n",
								field, fieldType, required, prop.Description))
						}
						continue
					}
				}

				// 检查是否需要忽略该参数
				if s.IgnoredFields[param.Name] {
					continue
				}

				// 处理普通参数
				paramType := s.getParamType(param)
				required := "否"
				if param.Required {
					required = "是"
				}
				file.WriteString(fmt.Sprintf("| %s | `%s` | %s | %s |\n",
					param.Name, paramType, required, param.Description))
			}
			file.WriteString("\n")
		}

		// 写入响应
		for _, response := range op.Responses {
			// 解析响应体结构
			respSchema := response.Schema
			if respSchema.Ref != "" || respSchema.Type != "" || len(respSchema.AllOf) > 0 {
				file.WriteString("\n**响应体结构**:\n\n")
				s.writeResponseSchema(file, respSchema, defs)
			}
		}
		file.WriteString("\n---\n")
	}
}

// 解决嵌套引用
func (s *SwaggerMarkdown) resolveSchema(schema Schema, defs map[string]Definition) Definition {
	if schema.Ref != "" {
		refName := s.extractRefName(schema.Ref)
		if def, exists := defs[refName]; exists {
			return def
		}
	}
	return Definition{}
}

// 写入定义字段
func (s *SwaggerMarkdown) writeDefinitionFields(file *os.File, def Definition, defs map[string]Definition) {
	// 处理 allOf 结构
	for _, allOfItem := range def.AllOf {
		if allOfItem.Ref != "" {
			refName := s.extractRefName(allOfItem.Ref)
			if nestedDef, exists := defs[refName]; exists {
				s.writeDefinitionFields(file, nestedDef, defs)
			}
		} else if allOfItem.Properties != nil {
			// 处理内联属性
			file.WriteString("| 字段 | 类型 | 描述 |\n")
			file.WriteString("|------|------|------|\n")
			for field, prop := range allOfItem.Properties {
				fieldType := s.getPropertyType(prop)
				file.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", field, fieldType, prop.Description))
			}
		}
	}

	if len(def.Properties) > 0 {
		file.WriteString("| 字段 | 类型 | 描述 |\n")
		file.WriteString("|------|------|------|\n")

		for field, prop := range def.Properties {
			fieldType := s.getPropertyType(prop)
			file.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", field, fieldType, prop.Description))
		}
		file.WriteString("\n")

		// 处理嵌套结构
		for field, prop := range def.Properties {
			// 处理嵌套对象
			if prop.Properties != nil {
				if nestedProps, ok := prop.Properties.(map[string]interface{}); ok {
					file.WriteString(fmt.Sprintf("**%s 结构详情**:\n\n", field))
					s.writeNestedFields(file, nestedProps, defs)
				}
			}

			// 处理嵌套引用
			if prop.Ref != "" {
				refName := s.extractRefName(prop.Ref)
				if nestedDef, exists := defs[refName]; exists {
					file.WriteString(fmt.Sprintf("**%s 结构详情**:\n\n", field))
					s.writeDefinitionFields(file, nestedDef, defs)
				}
			}

			// 处理 allOf 结构
			if len(prop.AllOf) > 0 {
				for _, allOfItem := range prop.AllOf {
					if allOfItem.Ref != "" {
						refName := s.extractRefName(allOfItem.Ref)
						if nestedDef, exists := defs[refName]; exists {
							file.WriteString(fmt.Sprintf("**%s 结构详情**:\n\n", field))
							s.writeDefinitionFields(file, nestedDef, defs)
						}
					}
				}
			}

			// 处理数组类型
			if prop.Type == "array" && prop.Items != nil && prop.Items.Ref != "" {
				refName := s.extractRefName(prop.Items.Ref)
				if df, exists := defs[refName]; exists {
					file.WriteString(fmt.Sprintf("**%s 数组元素结构**:\n\n", field))
					s.writeDefinitionFields(file, df, defs)
				}
			}
		}
	}
}

// 写入响应体结构
func (s *SwaggerMarkdown) writeResponseSchema(file *os.File, schema Schema, defs map[string]Definition) {
	// 处理 allOf 结构
	if len(schema.AllOf) > 0 {
		for _, allOfItem := range schema.AllOf {
			if allOfItem.Ref != "" {
				refName := s.extractRefName(allOfItem.Ref)
				if def, exists := defs[refName]; exists {
					s.writeDefinitionFields(file, def, defs)
				}
			}
		}
		return
	}

	// 处理引用类型
	if schema.Ref != "" {
		refName := s.extractRefName(schema.Ref)
		if def, exists := defs[refName]; exists {
			s.writeDefinitionFields(file, def, defs)
			return
		}
	}

	// 处理数组类型
	if schema.Type == "array" && schema.Items != nil {
		if schema.Items.Ref != "" {
			refName := s.extractRefName(schema.Items.Ref)
			file.WriteString(fmt.Sprintf("数组类型: `[]%s`\n\n", refName))
			if def, exists := defs[refName]; exists {
				file.WriteString(fmt.Sprintf("**数组元素结构**:\n\n"))
				s.writeDefinitionFields(file, def, defs)
			}
			return
		}
	}

	// 处理基本类型
	if schema.Type != "" {
		file.WriteString(fmt.Sprintf("`%s`\n", schema.Type))
	}
}

// 写入嵌套字段
func (s *SwaggerMarkdown) writeNestedFields(file *os.File, props map[string]interface{}, defs map[string]Definition) {

	file.WriteString("| 字段 | 类型 | 描述 |\n")
	file.WriteString("|------|------|------|\n")

	for field, prop := range props {
		if propMap, ok := prop.(map[string]interface{}); ok {
			fieldType := "object"
			if t, exists := propMap["type"]; exists {
				fieldType = t.(string)
			}
			if ref, exists := propMap["$ref"]; exists {
				fieldType = s.extractRefName(ref.(string))
			}
			if items, exists := propMap["items"]; exists {
				if itemsMap, ook := items.(map[string]interface{}); ook {
					if ref, ext := itemsMap["$ref"]; ext {
						fieldType = "[]" + s.extractRefName(ref.(string))
					}
				}
			}

			description := ""
			if desc, exists := propMap["description"]; exists {
				description = desc.(string)
			}

			file.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n", field, fieldType, description))

			// 递归处理嵌套引用
			if ref, exists := propMap["$ref"]; exists {
				refName := s.extractRefName(ref.(string))
				if nestedDef, ext := defs[refName]; ext {
					file.WriteString(fmt.Sprintf("\n**嵌套引用**: %s\n\n", refName))
					s.writeDefinitionFields(file, nestedDef, defs)
				}
			}

			// 处理嵌套属性
			if nestedProps, exists := propMap["properties"]; exists {
				if nestedMap, ok := nestedProps.(map[string]interface{}); ok {
					file.WriteString(fmt.Sprintf("\n**%s 结构详情**:\n\n", field))
					s.writeNestedFields(file, nestedMap, defs)
				}
			}

			// 处理 allOf 结构
			if allOf, exists := propMap["allOf"]; exists {
				if allOfList, ok := allOf.([]interface{}); ok {
					for _, item := range allOfList {
						if allOfMap, ok := item.(map[string]interface{}); ok {
							if ref, exists := allOfMap["$ref"]; exists {
								refName := s.extractRefName(ref.(string))
								if nestedDef, ext := defs[refName]; ext {
									file.WriteString(fmt.Sprintf("\n**%s 结构详情**:\n\n", field))
									s.writeDefinitionFields(file, nestedDef, defs)
								}
							}
						}
					}
				}
			}
		}
	}
}

// 获取参数类型（处理嵌套引用）
func (s *SwaggerMarkdown) getParamType(param Parameter) string {
	if param.Schema.Ref != "" {
		return s.extractRefName(param.Schema.Ref)
	}
	if param.Schema.Type == "array" && param.Schema.Items != nil {
		return "[]" + s.extractRefName(param.Schema.Items.Ref)
	}
	return param.Type
}

// 获取属性类型（处理嵌套类型）
func (s *SwaggerMarkdown) getPropertyType(prop Property) string {
	if prop.Ref != "" {
		return s.extractRefName(prop.Ref)
	}
	if prop.Type == "array" && prop.Items != nil {
		if prop.Items.Ref != "" {
			return "[]" + s.extractRefName(prop.Items.Ref)
		}
		return "[]" + prop.Type
	}
	if prop.Format != "" {
		return prop.Type + " (" + prop.Format + ")"
	}
	if prop.Properties != nil {
		return "object"
	}
	if len(prop.AllOf) > 0 {
		return "object (allOf)"
	}
	return prop.Type
}

// 检查字符串是否在数组中
func (s *SwaggerMarkdown) contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// 从 $ref 中提取模型名称
func (s *SwaggerMarkdown) extractRefName(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1]
}
