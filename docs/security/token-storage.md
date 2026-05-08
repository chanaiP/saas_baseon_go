# Token Storage Security Policy

当前前端保留 `localStorage` 中的 `access_token`，仅作为既有后台管理台兼容策略，不把它升级为长期推荐方案。

## 强制约束

- 后端必须启用 CSP、`X-Frame-Options`、`X-Content-Type-Options`、`Referrer-Policy`，生产环境启用 HSTS。
- 前端不得使用 `v-html`、`innerHTML`、`outerHTML` 或其他未消毒 HTML 注入方式。
- 前端不得把 `access_token`、`Authorization` 或 bearer token 输出到 console、URL、错误提示或持久化到非授权位置。
- `access_token` 只允许在登录写入、路由守卫读取、HTTP 拦截器注入、登出/失效清理路径中出现。
- localStorage 只允许保存用户偏好、快捷入口等低敏状态；菜单和权限以后端返回为权威。

## 回归检查

执行：

```bash
cd frontend
npm run security:check
```

该脚本会扫描 `src` 下的危险 HTML 注入、token 日志泄漏和未授权 token 存储位置。新增前端入口如果确实需要读取 token，必须先把用途收敛到认证或 HTTP 基础设施中。
