# en2cn

一个用 Go 实现的“英语 → 中文谐音” SDK。它将英文单词转换为 TTS（文字转语音）友好的真实中文词组，适合在语音助手、智能客服等场景中播报外文专有名词。

## 1. 背景与目标

- **输入**：任意英文单词或短语（如 `Tesla`、`OpenAI`）。
- **输出**：来自中文词库的真实词组（如 `特斯拉`、`偶也来`）。
- **约束**：输出必须是现成词条，避免拼凑字导致 TTS 断句或读音错误。
- **核心思路**：三段式流水线
  1. 英语 → IPA（国际音标）
  2. IPA → 汉语拼音结构（声母/韵母）
  3. 拼音结构 → 中文候选词匹配

## 2. 系统架构

| 阶段 | 说明 | 本项目实现 |
| --- | --- | --- |
| 1. 英语 → IPA | 预先计算大量英文词与 IPA 的映射，运行时只查表 | `data/eng_ipa.json`（由 `cmd/generate-data` 从 CMUdict 生成） |
| 2. IPA → 拼音部件 | 以贪婪匹配方式把 IPA 拆成 Go 结构体 `PinyinPart` | `data/ipa_to_parts.json` + `tokenizer.go` |
| 3. 拼音部件 → 中文词组 | 以声母/韵母相似度为权重跑加权 Levenshtein，选出最接近的中文候选 | 候选集来自 `data/candidate_db.json`，相似度矩阵在 `data/shengmu_similarity.json` / `data/yunmu_similarity.json` |

所有数据通过 `embed` 打包，`NewEngine()` 启动时一次性加载，`Engine.Convert(word)` 即可完成转换。

## 3. 代码结构

```
.
├── cmd/generate-data     # 数据生成脚本，自动下载 CMUdict + 结巴词库
├── data/                 # 运行所需的全部 JSON 数据
├── examples/basic        # 演示如何调用 SDK
├── similarity.go         # 声韵母相似度（加权 Levenshtein）
├── tokenizer.go          # IPA -> PinyinPart 贪婪 tokenizer
├── engine.go             # Engine.Convert 工作流（含 overrides 支持）
└── ...                   # 其它辅助文件（types.go、errors.go 等）
```

## 4. 快速开始

### 4.1 生成数据（一次性）

```bash
go run ./cmd/generate-data -candidate-limit 120000
```

- CMUdict 与结巴词库会被下载到 `third_party/`（已加入 `.gitignore`）。
- 新生成的 `data/eng_ipa.json`（约 3.5 MB）和 `data/candidate_db.json`（约 23 MB）会覆盖原文件。
- 需要强制指定发音时，可直接修改 `data/manual_overrides.json`，不必重新生成大词典。
- 通过 `-candidate-limit` 或自定义 URL 可以控制数据规模与来源。

### 4.2 运行示例程序

```bash
go run ./examples/basic
```

示例输出（结果会随词库差异而变化；以下条目来自 `data/manual_overrides.json`）：

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

> 如果希望示例输出符合业务习惯，可直接编辑 `data/manual_overrides.json`，无需重新跑数据生成脚本。

### 4.3 在业务代码中使用

```go
engine, err := en2cn.NewEngine()
if err != nil {
    log.Fatalf("init engine: %v", err)
}

word := "openai"
zh, err := engine.Convert(word)
if err != nil {
    log.Printf("convert error: %v", err)
}
fmt.Println(zh)
```

## 5. 数据生成流水线详解

1. **英语 → IPA**  
   - `cmd/generate-data` 下载 CMUdict（约 3.7 MB），解析 ARPABET，调用 `convertArpabetToIPA` 写入 `data/eng_ipa.json`。  
   - 支持通过 `manualIPA` 注入自定义条目（例如 CMU 没有收录的 “openai”）。

2. **IPA → PinyinPart**  
   - 运行时加载 `data/ipa_to_parts.json`，`TokenizeIPA()` 先清洗 IPA，再做最长匹配切分。  
   - `normalizeParts()` 负责把纯声母或纯韵母的切片合并成完整音节。

3. **中文候选库**  
   - 生成脚本下载 `dict.txt.big`，用 `github.com/mozillazg/go-pinyin` 转为拼音。  
   - 每个词条拆成 `PinyinPart` 序列后写入 `data/candidate_db.json`，默认保留 120k 词。

4. **相似度计算**  
   - `similarity.go` 使用带权 Levenshtein：替换成本 = `1 - avg(shengmuScore, yunmuScore)`。  
   - 声母/韵母权重矩阵定义在 `data/shengmu_similarity.json` / `data/yunmu_similarity.json`。

5. **引擎工作流**  
   - `Engine.Convert()`：查 IPA → Tokenize → 遍历候选库打分 → 返回得分最高的词。  
   - 若词典缺少该单词，会返回 `ErrIPAUnavailable`。
6. **人工兜底**  
   - `data/manual_overrides.json` 中的词条会在自动匹配前被直接返回，适合为品牌名称、口语热词等提供权威谐音。

## 6. 开发与验证

- **格式化**：`gofmt -w ./...`
- **测试**：`go test ./...`
- **示例**：`go run ./examples/basic`
- **数据更新**：`go run ./cmd/generate-data -candidate-limit 120000`

## 7. 常见扩展方向

1. **替换数据源**：可自行准备更大的英文/中文词典，并通过参数传给 `cmd/generate-data`。  
2. **多音字处理**：目前依据 `go-pinyin` 的默认读音，可按需引入自定义读音表。  
3. **候选词打分**：若词库较大，可在 `Engine.Convert` 中加入并发或 Top-K 剪枝。  
4. **嵌入式存储**：若 JSON 太大，可改为 Badger/Bolt 等 KV 库，并在 `NewEngine` 中按需加载。

---

如需进一步定制或接入生产环境，建议先运行生成脚本获得完整数据，再根据业务要求（词库、权重、过滤规则等）调整相关 JSON 或算法。EOF
