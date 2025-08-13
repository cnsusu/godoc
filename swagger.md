# API接口文档

**版本**: v1.0.0  
**描述**: 

## 接口概览

## 接口详情
### 注册

**接口地址**：`/api/user/register`  
**请求方式**：`POST`  

**请求参数:**

| 名称 | 类型 | 是否必填 | 描述 |
|------|------|----------|------|
| nick_name | `string` | 否 | 用户昵称 |
| user_mobile | `string` | 是 | 手机号码 |
| user_name | `string` | 是 | 用户姓名 |


**响应体结构**:

| 字段 | 类型 | 描述 |
|------|------|------|
| msg | `string` | 错误信息 |
| code | `integer` | 错误码 |


---
### 查询

**接口地址**：`/api/user/query`  
**请求方式**：`GET`  

**请求参数:**

| 名称 | 类型 | 是否必填 | 描述 |
|------|------|----------|------|
| user_mobile | `string` | 是 | 手机号码 |


**响应体结构**:

| 字段 | 类型 | 描述 |
|------|------|------|
| code | `integer` | 错误码 |
| data | `[]response.UserInfo` | 数据 |
| msg | `string` | 错误信息 |

**data 数组元素结构**:

| 字段 | 类型 | 描述 |
|------|------|------|
| user_mobile | `string` | 手机号码 |
| user_name | `string` | 用户姓名 |
| nick_name | `string` | 用户昵称 |


---
