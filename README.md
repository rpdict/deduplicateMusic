### 设计说明与可改进之处（短）

- 我使用了“中位数分块哈希”的轻量感知指纹方法，简单、并且对音量/编码差异有一定鲁棒性；如果需要更强的音频相似度判定（对变速、混响、重编码更鲁棒），建议接入成熟指纹库（如 Chromaprint / AcoustID）或基于谱图+局部最大值的特征点法。

- 当前去重为 O(N²) 比较，若文件数量非常大（上万），建议先基于文件长度 / 采样特征做桶化（hash bucketing）或使用 LSH 降低比较对数。

- 复制时尽量保留元数据；生产环境可按需保留修改时间、权限、硬链接等。

- 可以增加“dry-run”模式，仅输出将被删除的文件或将被保留的列表，便于用户核对。 

### 运行程序（需要 ffmpeg）：
``` go run ./cmd/audio-dedup -src testMusic -dst testMusic/out -workers 4 -threshold 8 -seconds 8 -v ```

### ffmpeg 安装（示例）：

- macOS (homebrew): brew install ffmpeg

- Ubuntu/Debian: sudo apt-get install ffmpeg

- Windows: 下载 ffmpeg 并把 ffmpeg.exe 放到 PATH