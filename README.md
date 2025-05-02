minisocks - 轻量级 SOCKS5 网络代理工具

![GitHub release](https://img.shields.io/github/release/beijian128/minisocks) ![License](https://img.shields.io/badge/license-MIT-blue)

minisocks 是一个轻量级的 SOCKS5 代理工具，提供简单易用的网络代理解决方案。

功能特性

• ✅ 轻量级 SOCKS5 协议实现

• 🔒 内置数据混淆功能

• ⚡ 高性能网络传输

• 🔄 自动生成安全密码

• 📁 简洁的 JSON 配置文件

• 🖥️ 跨平台支持


快速开始

1. 下载安装

```bash
# 使用 curl 下载最新版本 (Linux/macOS)
curl -L https://github.com/beijian128/minisocks/releases/latest/download/minisocks-$(uname -s)-$(uname -m).tar.gz | tar xz
```

或前往 [GitHub Releases](https://github.com/beijian128/minisocks/releases) 手动下载适合您系统的版本。

2. 服务端部署

```bash
# 在服务器上运行
./minisocks-server
```

首次运行会自动生成配置文件 `~/.minisocks.json` 并显示初始配置：

```
[INFO] 服务启动成功
监听地址: 0.0.0.0:7448
认证密码: ******** (请妥善保存)
```

3. 客户端配置

```bash
# 在本地运行
./minisocks-local
```

修改生成的配置文件 `./minisocks.json`：

```json
{
  "remote": "your.server.ip:7448",
  "password": "server_password_here",
  "listen": "127.0.0.1:7448"
}
```

重新启动客户端：

```bash
./minisocks-local
```

4. 配置代理

配置您的系统或浏览器使用 SOCKS5 代理：

• 地址：`127.0.0.1`

• 端口：`7448`


推荐浏览器扩展：
• Chrome/Edge: [SwitchyOmega](https://chrome.google.com/webstore/detail/proxy-switchyomega/padekgcemlokbadohgkifijomclgjgif)

• Firefox: [FoxyProxy](https://addons.mozilla.org/firefox/addon/foxyproxy-standard/)


详细配置

客户端配置 (minisocks-local)

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `password` | 加密密码（需与服务端一致） | 自动生成 | "your_password" |
| `listen` | 本地监听地址 | "0.0.0.0:7448" | "127.0.0.1:7448" |
| `remote` | 远程服务器地址 | "0.0.0.0:7448" | "45.56.76.5:7448" |

服务端配置 (minisocks-server)

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `password` | 加密密码 | 自动生成 | "your_password" |
| `listen` | 服务监听地址 | "0.0.0.0:7448" | ":7448" |

配置文件示例

```json
{
  "remote": "45.56.76.5:7448",
  "password": "your_secure_password_here",
  "listen": "127.0.0.1:7448"
}
```

注意事项

1. 🔐 客户端和服务端的 `password` 必须完全一致
2. ⚠️ 自动生成的密码强度更高，建议不要手动修改
3. 🔄 修改配置后需要重启服务生效
4. 📍 默认配置文件路径为 `./minisocks.json`
5. 🌐 确保服务器防火墙已开放相应端口

