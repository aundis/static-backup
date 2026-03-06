# 文件备份系统

一个基于 Go 语言和 Vue3 开发的文件备份系统，支持指定文件夹备份、计划任务、备份记录管理和 OSS 配置管理。

## 功能特性

- ✅ 文件备份：支持指定文件夹打包压缩备份
- ✅ 文件恢复：支持从备份文件恢复到指定目录
- ✅ 计划任务：支持定时备份，可设置保留份数
- ✅ 备份记录：支持查看、删除、恢复和下载备份文件
- ✅ OSS 配置：支持配置阿里云 OSS，备份文件可同步到云端
- ✅ 登录验证：支持密码登录验证
- ✅ 目录选择：支持通过弹窗选择目录
- ✅ 多级目录：支持选择二级及更深层次目录
- ✅ 文件排序：文件夹排在上面，然后按名称升序排序

## 技术栈

- **后端**：Go 语言、Gin 框架、cron 库、阿里云 OSS SDK
- **前端**：Vue3、Vite、Axios

## 安装部署

### 本地运行

1. **克隆项目**
   ```bash
   git clone <项目地址>
   cd static-backup
   ```

2. **运行后端**
   ```bash
   cd backend
   go run main.go backup.go restore.go schedule.go directory.go
   ```

3. **运行前端**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

4. **访问应用**
   打开浏览器，访问 http://localhost:5173

### Docker 部署

1. **构建镜像**
   ```bash
   docker build -t backup-system .
   ```

2. **运行容器**
   ```bash
   docker run -d -p 8080:8080 --name backup-system backup-system
   ```

3. **访问应用**
   打开浏览器，访问 http://localhost:8080

### GitHub Actions 自动构建

项目已配置 GitHub Actions 工作流，当代码推送到 main 分支时，会自动构建并推送 Docker 镜像到 Docker Hub。

需要在 GitHub 仓库设置以下 Secrets：
- `DOCKER_USERNAME`：Docker Hub 用户名
- `DOCKER_PASSWORD`：Docker Hub 密码或访问令牌

## 使用说明

### 登录
- 默认密码：admin123
- 登录后可访问系统所有功能

### 备份
1. 点击「备份」标签页
2. 点击「选择源目录」按钮，在弹窗中选择要备份的目录
3. 点击「开始备份」按钮
4. 备份完成后，会显示备份结果

### 恢复
1. 点击「恢复」标签页
2. 点击「选择备份文件」按钮，在弹窗中选择要恢复的备份文件
3. 点击「选择目标目录」按钮，在弹窗中选择恢复的目标目录
4. 点击「开始恢复」按钮
5. 恢复完成后，会显示恢复结果

### 计划任务
1. 点击「计划任务」标签页
2. 点击「添加计划任务」按钮
3. 填写任务名称、选择源目录、设置 Cron 表达式、设置保留份数
4. 可选：选择 OSS 配置，将备份文件同步到云端
5. 点击「保存」按钮
6. 可手动触发、编辑或删除计划任务

### 备份记录
1. 点击「备份记录」标签页
2. 可按计划任务筛选备份记录
3. 可删除、恢复或下载备份文件

### OSS 配置
1. 点击「OSS配置」标签页
2. 点击「添加OSS配置」按钮
3. 填写配置名称、Endpoint、Access Key ID、Access Key Secret、Bucket名称、前缀
4. 点击「测试连接」按钮，测试配置是否正确
5. 点击「保存」按钮
6. 可编辑或删除 OSS 配置

## 目录结构

```
static-backup/
├── backend/            # 后端代码
│   ├── backup/         # 备份文件存储目录
│   ├── config/         # 配置文件目录
│   ├── backup.go       # 备份功能
│   ├── directory.go    # 目录遍历功能
│   ├── main.go         # 主入口
│   ├── restore.go      # 恢复功能
│   └── schedule.go     # 计划任务功能
├── frontend/           # 前端代码
│   ├── dist/           # 构建产物
│   ├── src/            # 源代码
│   ├── index.html      # 入口 HTML
│   ├── package.json    # 依赖配置
│   └── vite.config.js  # Vite 配置
├── .github/            # GitHub Actions 配置
├── Dockerfile          # Docker 构建文件
└── README.md           # 项目说明
```

## 配置说明

- **计划任务配置**：存储在 `backend/config/schedules.json`
- **OSS 配置**：存储在 `backend/config/oss_configs.json`
- **备份文件**：存储在 `backend/backup/` 目录

## 注意事项

1. 首次运行时，系统会自动创建必要的目录和配置文件
2. 备份文件默认存储在 `backend/backup` 目录
3. 计划任务和 OSS 配置会持久化到本地 JSON 文件
4. 登录状态会保存在浏览器的 localStorage 中
5. 实际生产环境中，建议修改默认密码，并使用更安全的认证方式

## 许可证

MIT
