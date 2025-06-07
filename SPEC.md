# MinecraftModsLocalizer 仕様書

## プロジェクト概要
Minecraftのmod翻訳を支援するクロスプラットフォームGUIアプリケーション

## 技術仕様

### フレームワーク
- **言語**: Go
- **GUI**: Wails v2 (OSネイティブUI)
- **対応OS**: Windows/Mac/Linux

### 対応ファイル形式
1. **JSON形式** - `{"key": "value"}` (例: `en_us.json`, `ja_jp.json`)
2. **.lang形式** - `key=value` (例: `en_US.lang`, `ja_JP.lang`)
3. **SNBT形式** - Structure NBT テキスト表現

### 翻訳機能
- **翻訳元**: 英語（デフォルト）→ 将来的に自動検出
- **翻訳先**: 利用者指定の言語
- **翻訳エンジン**:
  - Google翻訳API
  - DeepL翻訳API
  - OpenAI API準拠LLM
  - **デフォルト**: gpt-4.1-mini（コスパ重視）

### GUI要件
- ファイル選択ダイアログ
- 翻訳エンジン選択
- 翻訳先言語選択
- 進捗表示
- ログ出力
- OSネイティブルック&フィール

### 将来実装予定
- 翻訳元言語自動検出
- バッチ処理機能
- 翻訳履歴管理

## 開発優先度
1. Wails v2プロジェクト初期化
2. ファイル形式パーサー実装
3. 翻訳API連携
4. GUI実装