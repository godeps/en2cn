# en2cn

面向语音播报场景的英语 → 中文谐音 SDK。通过“英文 → IPA → 拼音部件 → 中文候选”流水线，让 TTS 能更自然地读出英文专有名词。

---

## 1. 功能概览

| 能力 | 描述 |
| --- | --- |
| 英文转 IPA | 预先生成 `eng_ipa.json`，运行时 O(1) 查表 |
| IPA 贪婪切分 | `tokenizer.go` 根据 `ipa_to_parts.json` 将 IPA 拆成声母/韵母部件 |
| 中文候选库 | `candidate_db.json` 存储 10 万+ 真实词条及其拼音部件 |
| 声韵母相似度 | `similarity.go` 使用带权 Levenshtein 计算匹配分数 |
| 人工兜底 | `manual_overrides.json` 为高频词配置固定译名 |
| 数据生成脚本 | `cmd/generate-data` 自动下载 CMUdict、结巴词库并生成 JSON |

---

## 2. 目录结构

```
.
├── cmd/generate-data     # 数据生成脚本
├── data/                 # 所有嵌入式 JSON 数据
├── examples/basic        # CLI 示例
├── engine.go             # Convert 入口，含 override 支持
├── similarity.go         # 声韵母相似度算法
├── tokenizer.go          # IPA → 拼音部件
├── types.go / errors.go  # 基础定义
├── *_test.go             # 单元测试
tag── README.md
```

---

## 3. 快速开始

### 3.1 生成 / 刷新数据

```bash
# 约 25 MB 数据，含 10w+ 英文词、12w+ 中文候选
go run ./cmd/generate-data -candidate-limit 120000
```

- 原始语料缓存到 `third_party/`（已忽略提交）。
- `eng_ipa.json`、`candidate_db.json` 会被覆盖。
- 想固定读音时，直接编辑 `data/manual_overrides.json`，无需重跑脚本。

### 3.2 运行示例

```bash
go run ./examples/basic
```

示例输出（来自 `manual_overrides.json`）：

```
hello -> 哈喽
coffee -> 咖啡
tiger -> 太格
banana -> 巴娜娜
tesla -> 特斯拉
apple -> 苹果
google -> 谷歌
microsoft -> 迈克软
openai -> 欧朋爱
```

> 想要不同读法，只需在 `manual_overrides.json` 中增加条目。

### 3.3 嵌入业务代码

```go
engine, err := en2cn.NewEngine()
if err != nil {
    log.Fatalf("init engine: %v", err)
}

zh, err := engine.Convert("openai")
if err != nil {
    log.Printf("convert error: %v", err)
}
fmt.Println(zh)
```

---

## 4. 数据流水线

1. **英文 → IPA**：`cmd/generate-data` 下载 CMUdict，使用 `convertArpabetToIPA` 生成 `eng_ipa.json`；`manualIPA` 可补漏词。
2. **IPA → 拼音部件**：`tokenizer.go` 依据 `ipa_to_parts.json` 做最长匹配并通过 `normalizeParts` 组合音节。
3. **中文候选库**：脚本解析 `dict.txt.big`，用 `github.com/mozillazg/go-pinyin` 得到 `PinyinPart` 列表并写入 `candidate_db.json`。
4. **声韵母相似度**：`similarity.go` 读取 `shengmu/yunmu_similarity.json`，以带权 Levenshtein 计算 0~1 的相似度分。
5. **Engine.Convert**：优先检查 `manual_overrides.json`，否则按“IPA→拼音→遍历候选打分”选取最高分词。

---

## 5. 测试与验证

| 命令 | 作用 |
| --- | --- |
| `gofmt -w ./...` | 统一代码风格 |
| `go test ./...`  | 运行 `engine_test.go`、`tokenizer_test.go`、`similarity_test.go` |
| `go run ./examples/basic` | 人工检验关键词输出 |
| `go run ./cmd/generate-data ...` | 刷新词库/候选数据 |

---

## 6. 扩展与优化

1. **替换数据源**：把企业词库或垂直领域词典喂给 `cmd/generate-data`。
2. **多音字支持**：在 `go-pinyin` 结果上套自定义读音表或词性标注。
3. **候选打分改进**：引入词频、词性、语境特征；或在 `Engine.Convert` 中并发遍历、做 Top-K 剪枝。
4. **存储形态**：当 JSON 过大时，可替换成 Badger/Bolt/SQLite 等，并在 `NewEngine` 中按需加载。
5. **自动化纠错**：
   - 词频驱动回溯：统计线上词频，与官方译名比对后自动生成 override 草案。
   - ASR 回流：把 TTS 音频送入 ASR，偏差大的词进入待修正队列。
   - 多候选 PK：保留前 K 个候选并结合语言模型/词频做二次排序，低置信度词触发人工审核。
   - 外部语料增补：解析新闻稿、字幕、百科双语资源，自动补充常见音译词。
   - 规则/模型融合：针对 “-tion”等后缀沉淀规律，在算法分数偏低时由规则结果兜底。
   - 灰度与反馈：线上灰度不同候选，收集用户纠错数据，回流到 override 与训练集。

---

## 7. FAQ

- **为什么自动结果有时“不像中文”？**  
  通用词典缺少品牌音译/语境信息，且相似度算法只看发音；可通过 override、词频回流或规则纠偏提升质量。

- **如何替换词库但保留 override？**  
  运行 `cmd/generate-data` 后再覆盖 `data/manual_overrides.json`，或在脚本中合并自定义表。

- **能否部署成服务？**  
  可以在 `Engine.Convert` 外包一层 HTTP/gRPC，同步提供 override 管理接口。

---

## 8. License

MIT（见 `LICENSE`）。
