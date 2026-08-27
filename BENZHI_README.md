# BENZHI_README

## 项目说明
- 项目：benzhi-project-83e7993f-657c-4d69-8232-d79e3af7a8b0
- 项目用途：面向考古陶器整理团队的陶片拼合证据评议工作台，使用可追溯状态机、SQLite 事务、版本化证据、独立复核和 SHA-256 摘要链完成从案件建档到只读重建档案校验的完整流程。
- Go 工具链：`golang:1.23.0`
- 前端工具链：原生 HTML、CSS 和 JavaScript，由 Go 服务直接提供

## 项目描述
- 项目名称：sherd-proof
- 项目介绍：面向考古陶器整理团队的陶片拼合假说证据评议工作台，以可追溯状态机管理陶片登记、候选拼合、反证处置、同行复核和定稿档案生成，避免未经充分论证的器形重建进入研究成果。
- 项目概述：面向考古陶器整理团队的陶片拼合假说证据评议工作台，以可追溯状态机管理陶片登记、候选拼合、反证处置、同行复核和定稿档案生成，避免未经充分论证的器形重建进入研究成果。
- 核心工作流：整理员创建拼合案件并冻结陶片基线，登记候选拼合及其形态证据，系统计算证据完整度并开放异议；异议逐项经补证、撤回候选或维持答复处理后，由未参与建档的同行复核员给出退回或通过结论，最终生成带摘要校验值的只读重建档案并将案件转为定稿状态。
- 对外接口：Go 服务直接提供无需 Node 构建的原生单页浏览器工作台和同源 JSON 端点；页面包含案件状态带、陶片基线表、候选拼合证据矩阵、异议队列、复核面板和定稿档案校验视图，可在一个案件上下文内完成唯一业务流程。

## 标准构建、运行和测试命令
进入容器后执行：
cd '/app' && GOTOOLCHAIN=local go build ./...

cd /app && GOTOOLCHAIN=local go run ./cmd/server -selftest -addr=127.0.0.1:19081

cd '/app' && GOTOOLCHAIN=local go test ./...

## Docker 构建和进入容器
chmod +x build_benzhi_docker.sh

./build_benzhi_docker.sh benzhi-project-83e7993f-657c-4d69-8232-d79e3af7a8b0-amd64 linux/amd64

./build_benzhi_docker.sh benzhi-project-83e7993f-657c-4d69-8232-d79e3af7a8b0-arm64 linux/arm64

docker run -it benzhi-project-83e7993f-657c-4d69-8232-d79e3af7a8b0-amd64:latest

## 题目验证命令
1. 预期退出码 0：`go test ./...`
2. 预期退出码 0：`go run ./cmd/server -selftest -addr=127.0.0.1:19081`
