# en2cn

面向语音播报场景的英语→中文谐音 SDK。通过“英文 → IPA → 拼音部件 → 中文候选”流水线，把英文词转换成真实的中文词组，保证 TTS 输出自然可读。

---

## 1. 功能概览

| 能力 | 说明 |
| --- | --- |
| 英文转 IPA | 预先生成 `eng_ipa.json`，运行期 O(1) 查表 |
| IPA 贪婪分词 | `tokenizer.go` 依据 `ipa_to_parts.json` 将音标拆成声母/韵母结构 |
| 中文候选库 | `candidate_db.json` 保存 12 万+ 词条的 `PinyinPart` 序列 |
| 声韵母相似度 | `similarity.go` 基于带权 Levenshtein 计算匹配分数 |
| 人工兜底 | `manual_overrides.json` 为高频词指定权威译名 |
| 数据生成脚本 | `cmd/generate-data` 自动下载 CMUdict、结巴词库并生成 JSON |

---

## 2. 目录结构

```
.
├── cmd/generate-data     # 数据生成脚本
├── data/                 # 各类 JSON 数据（可通过脚本刷新）
├── examples/basic        # 最小示例程序
├── engine.go             # Convert 工作流（含 override 逻辑）
├── similarity.go         # 声/韵母相似度算法
├── tokenizer.go          # IPA → PinyinPart 贪婪切分
├── types.go              # 核心结构体
tag── *_test.go            # 单元测试
└── README.md
```

---

## 3. 快速开始

### 3.1 生成/刷新数据

```bash
# 约 25 MB 数据，含 10w+ 英文词和 12w+ 中文候选
go run ./cmd/generate-data -candidate-limit 120000
```

- 原始语料缓存到 `third_party/`（已写入 `.gitignore`）。
- 新的 `data/eng_ipa.json`、`data/candidate_db.json` 会覆盖旧文件。
- 需要固定读音时，可直接编辑 `data/manual_overrides.json`，无需重跑脚本。

### 3.2 运行示例

```bash
go run ./examples/basic
```

示例输出（来源于 `manual_overrides.json`）：

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

> 希望自定义读音？直接在 `data/manual_overrides.json` 中添加条目即可。

### 3.3 嵌入到业务代码

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

1. **英文 → IPA**  
   - `cmd/generate-data` 下载 CMUdict（ARPABET），调用 `convertArpabetToIPA` 写入 `data/eng_ipa.json`。  
   - `manualIPA` 可补充漏词（如 “openai”）。
2. **IPA → 拼音部件**  
   - `data/ipa_to_parts.json` 描述单个音标到声/韵母的映射，`TokenizeIPA()` 做最长匹配并通过 `normalizeParts()` 拼合音节。
3. **中文候选库**  
   - 解析 `dict.txt.big`，借助 `github.com/mozillazg/go-pinyin` 生成候选的 `PinyinPart` 序列，默认保留 12 万词写入 `data/candidate_db.json`。
4. **声韵母相似度**  
   - `data/shengmu_similarity.json` 与 `data/yunmu_similarity.json` 存储权重；`CalculateSimilarity()` 以带权 Levenshtein 输出 0~1 分值。
5. **Engine.Convert()**  
   - 先查 `manual_overrides.json`；若无命中，再走“IPA → 拼音 → 遍历候选打分”的流程，返回最高分候选。

---

## 5. 测试与验证

| 命令 | 作用 |
| --- | --- |
| `gofmt -w ./...` | 统一代码风格 |
| `go test ./...`  | 运行 `engine_test.go`、`tokenizer_test.go`、`similarity_test.go` |
| `go run ./examples/basic` | 人工验证常见词输出 |
| `go run ./cmd/generate-data ...` | 刷新词典/候选数据 |

---

## 6. 常见扩展

1. **更换数据源**：把企业内部词库或特定领域词典喂给 `cmd/generate-data`。
2. **多音字/读音表**：接入自定义拼音词典或对 `go-pinyin` 结果做 post-process。
3. **候选打分优化**：
   - 引入词频/词性/语境特征，或对候选进行 Top-K 剪枝。
   - 将算法改为并发遍历，提升大词库下的吞吐。
4. **存储形态**：当 JSON 过大时，可换成 Badger/Bolt/SQLite 等持久化存储，再在 `NewEngine` 中按需加载。

---

## 7. 自动化纠错路线

为了让“常用英文词”读音更贴近大众习惯，可建立如下闭环：

1. **词频驱动回溯**：统计线上/语料词频，对 Top-N 词的输出与官方译名或人工黄金表对齐，差异项自动生成 override 草案。
2. **ASR 回流**：将 TTS 播报音频送入 ASR，比对识别结果与目标中文，偏差大的词打入待修正队列。
3. **多候选 PK**：保留前 K 个候选，结合语言模型、词频等特征做二次排序；若最高分低于阈值则触发人工审核。
4. **外部语料增补**：解析新闻稿、字幕、百科等双语资源，抽取常见音译词，自动写入 override 或候选库。
5. **规则/模型融合**：沉淀常见后缀、词缀的音译规则（如 "-tion" → "逊"），在算法分数偏低时由规则结果兜底。
6. **灰度与反馈**：上线灰度不同候选，收集用户纠错/投诉，自动把样本回流到 override 及训练数据中。

结合这些策略，系统可以“自动匹配 → 检测风险 → 生成修正 → 发布”循环迭代，override 机制则保障关键词即时可控。

---

## 8. FAQ

- **为什么自动结果有时“不像中文”？**  
  当前候选库来自通用词典，缺少品牌音译与语境信息；可通过 override、词频回流或规则纠偏改进。

- **如何替换词库但保留 override？**  
  先运行 `cmd/generate-data` 生成新 JSON，再把自定义 `manual_overrides.json` 覆盖回去即可。

- **能否改为在线服务？**  
  可以，在 `Engine.Convert` 之上封装 HTTP/gRPC 接口，或把 `manual_overrides` 放到数据库中动态更新。

---

## 9. License

MIT（见 `LICENSE`）。
