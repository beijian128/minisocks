minisocks - 轻量级 SOCKS5 网络混淆代理工具


快速开始

1. 下载安装
前往 [GitHub Releases](https://github.com/beijian128/minisocks/releases) 下载最新版本，选择适合您操作系统的版本。

2. 服务端配置 (minisocks-server)
1. 将 `minisocks-server` 上传到您的境外服务器
2. 首次运行：
   ```bash
   ./minisocks-server
   ```
3. 程序会自动生成配置文件 `~/.minisocks.json` 并显示初始配置：
   ```
   本地监听地址 listen：:7448
   密码 password：******
   ```
4. 记录下自动生成的密码（后续客户端需要使用）

3. 客户端配置 (minisocks-local)
1. 在本地电脑运行：
   ```bash
   ./minisocks-local
   ```
2. 修改自动生成的配置文件 `~/.minisocks.json`：
   ```json
   {
     "remote": "您的服务器IP:7448",
     "password": "服务器显示的密码"
   }
   ```
3. 重新启动客户端：
   ```bash
   ./minisocks-local
   ```

4. 使用代理
配置您的浏览器或系统使用 SOCKS5 代理：
• 地址：127.0.0.1

• 端口：7448


详细配置

客户端配置 (minisocks-local)
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `password` | 加密密码（需与服务端一致） | 自动生成 |
| `listen` | 本地监听地址（ip:port） | 0.0.0.0:7448 |
| `remote` | 远程服务器地址（ip:port） | 0.0.0.0:7448 |

服务端配置 (minisocks-server)
| 参数 | 说明 | 默认值 |
|------|------|--------|
| `password` | 加密密码（需与客户端一致） | 自动生成 |
| `listen` | 服务监听地址（ip:port） | 0.0.0.0:7448 |

注意事项
1. 客户端和服务端的 `password` 必须完全一致
2. 密码会自动生成，建议不要手动修改
3. 配置文件路径为 `~/.minisocks.json`
4. 修改配置后需要重启服务生效

示例配置文件
```json
{
  "remote": "45.56.76.5:7448",
  "password": "your_secure_password_here",
  "listen": "127.0.0.1:7448"
}
```
