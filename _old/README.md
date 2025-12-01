# ⚠️ 待清理目录

> 此目录包含项目重构后的旧文件和目录，保留用于数据安全

**创建日期**: 2025-12-01
**状态**: 待清理

---

## 📁 目录说明

此目录包含 ES Serverless Platform v2.0 重构前的旧文件和目录。所有内容已迁移到新的标准化结构，此处仅作备份保留。

---

## 🗂️ 包含内容

### 旧源代码目录

| 目录 | 新位置 | 说明 |
|------|--------|------|
| `server/` | `src/control-plane/` | Go 控制平面服务 |
| `es-plugin/` | `src/es-plugin/` | ES IVF 向量插件 |
| `frontend/` | `src/frontend/` | Web 管理界面 |

### 旧部署配置目录

| 目录 | 新位置 | 说明 |
|------|--------|------|
| `terraform/` | `deployments/terraform/` | Terraform IaC |
| `helm/` | `deployments/helm/` | Helm Charts |
| `k8s/` | `deployments/kubernetes/` | Kubernetes YAML |
| `demo/` | `examples/` | 示例代码 |

### 旧文档文件

以下文档已归档或替换：

**已替换的文档**:
- `FILE_STRUCTURE.md` → `PROJECT_STRUCTURE.md` (新版)
- `TERRAFORM_HELM_README.md` → `docs/deployment/terraform-helm.md`

**已归档的文档**:
- `IVF实现*.md` → `docs/archive/implementation-summary/`
- `具体要求.md`, `说明.md` 等 → `docs/archive/requirements/`

**临时文档（可删除）**:
- `DELIVERY_CHECKLIST.md`
- `DEPLOYMENT_STATUS.md`
- `DEPLOYMENT_SUCCESS.md`
- `IMPLEMENTATION_SUMMARY.md`
- `QUICK_REFERENCE.md`
- `PROJECT_RESTRUCTURE_PLAN.md`
- `README_OLD.md`
- `下一步操作指南.md`
- `实现情况清单.md`
- `文档说明README.md`

---

## ⏰ 清理计划

### 阶段 1: 验证期（当前）

**时间**: 2025-12-01 ~ 2025-12-31 (1个月)
**状态**: 保留所有文件
**目的**: 确保新结构正常工作，万一需要可以回滚

### 阶段 2: 过渡期

**时间**: 2026-01-01 ~ 2026-01-31 (1个月)
**状态**: 标记为废弃
**操作**:
- 添加 `.deprecated` 标记
- 在项目文档中提醒
- 收集用户反馈

### 阶段 3: 清理期

**时间**: 2026-02-01
**状态**: 可以安全删除
**操作**:
- 创建最终备份归档
- 删除 `_old/` 目录
- 更新项目文档

---

## 🔍 如何使用旧文件

### 如果需要查看旧代码

```bash
cd _old/server
# 查看旧的控制平面代码
```

### 如果需要恢复文件

```bash
# 复制到新位置（不推荐）
cp _old/server/some_file.go src/control-plane/

# 或比较差异
diff _old/server/some_file.go src/control-plane/some_file.go
```

### 如果需要查看旧文档

```bash
# 查看旧的实现文档
less _old/IVF实现总结.md

# 对应的新文档位置
less docs/archive/implementation-summary/IVF实现总结.md
```

---

## ⚠️ 重要提醒

1. **不要修改此目录中的文件**
   - 这些文件仅供参考
   - 所有开发工作应在新结构中进行

2. **不要在新代码中引用此目录**
   - 所有路径引用应指向新结构
   - 如需旧文件，应先迁移到新位置

3. **定期检查是否可以清理**
   - 验证新结构工作正常
   - 确认不再需要旧文件
   - 按计划安全删除

---

## 📋 清理检查清单

在删除此目录前，请确认：

- [ ] 新的项目结构已正常工作至少 1 个月
- [ ] 所有开发人员已迁移到新结构
- [ ] 所有 CI/CD 配置已更新
- [ ] 所有文档链接已更新
- [ ] 没有脚本引用 `_old/` 目录
- [ ] 已创建最终备份归档
- [ ] 团队同意删除此目录

---

## 🔗 相关文档

- [迁移指南](../MIGRATION_GUIDE.md) - 完整的迁移说明
- [项目结构](../PROJECT_STRUCTURE.md) - 新的目录结构
- [项目整理归档](../docs/archive/2025-12-01-project-cleanup/) - 重构和整理详情

---

## 📞 问题反馈

如果您在使用新结构时遇到问题，或者需要从 `_old/` 中恢复某些内容：

1. 提交 GitHub Issue（标注 `migration` 标签）
2. 联系项目维护者
3. 查看迁移文档

---

**维护说明**:
- 定期检查此目录大小
- 记录任何从此目录恢复文件的情况
- 在清理前创建归档备份

---

**最后更新**: 2025-12-01
**下次检查**: 2026-01-01
