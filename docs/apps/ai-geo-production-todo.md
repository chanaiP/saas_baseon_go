# AI GEO 生产级整改 TODO

## 必须完成

- [x] 前端 AI Gateway 调用必须走 AI GEO 后端受控代理，不能直接调用 `/api/ai-gateway/v1/invoke/stream` 绕过应用权限、套餐和审计边界。
- [x] 母稿生成、发布计划创建的配额消费必须和业务写入保持事务一致；业务失败不能扣减额度。
- [x] 逻辑归档必须释放租户内唯一编码，归档后允许复用原编码新建业务数据。
- [x] 母稿、素材等带业务关联的写入必须校验 brand/product/channel 等父子归属，避免错链数据。
- [x] AI GEO HTTP 响应不得直接透出 GORM model，需要收口为稳定 DTO。
- [x] `current_schema.sql` 必须同步 AI GEO 当前表结构与索引，生产 baseline 和增量迁移保持一致。
- [x] 为上述生产门禁补回归测试，并跑通 AI GEO 相关测试。
