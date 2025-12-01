# 📦 目录已迁移

此目录 `frontend/` 已迁移到新位置。

## ✅ 新位置

```
frontend/ → src/frontend/
```

## 📖 为什么迁移？

项目已进行标准化重构，采用更清晰的目录结构：
- `src/` - 所有源代码
- `deployments/` - 所有部署配置
- `scripts/` - 工具脚本（按功能分类）
- `docs/` - 文档中心（按类型分类）

## 🚀 启动前端

新路径下的启动命令：

```bash
cd src/frontend
python -m http.server 8000

# 访问
# http://localhost:8000
```

## 🔗 更多信息

查看完整的项目结构说明：
- [PROJECT_STRUCTURE.md](../PROJECT_STRUCTURE.md)
- [前端开发文档](../docs/development/frontend.md)

---

**注意**：此目录仅保留用于向后兼容，后续版本可能移除。请尽快更新您的引用路径。
