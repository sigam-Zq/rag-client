# Qdrant CLI (Go Agents CLI)

[English](#english) | [中文](#chinese)

---

<a name="english"></a>
## English

A command-line tool for indexing and searching documents in the Qdrant vector database. It supports extracting text from `.md`, `.txt`, `.docx`, and `.pdf` files, generating embeddings via Ollama (BGE-M3 model), and storing them in Qdrant.

### Features
- **File Ingestion**: Support for Markdown, Plain Text, Word (Docx), and PDF documents.
- **Dynamic Collections**: Automatically creates a new Qdrant collection for each new document, named after the file (e.g., `my_document.pdf` -> `my_document` collection).
- **Unique Document IDs**: Assigns a unique UUID to each indexed document, ensuring all its text chunks are linked.
- **Embedding Generation**: Uses Ollama with the `bge-m3` model for high-quality text embeddings.
- **Advanced Vector Search**:
    - **Global Search**: Search across all collections simultaneously to find the most relevant results.
    - **Scoped Search**: Specify a collection to narrow down the search scope.
- **RAG Support (Chat with LLM)**: Ask questions and get answers powered by Ollama, using retrieved document context to improve accuracy.
- **Clean Architecture**: A well-organized project structure with modular packages for clarity and maintainability.

### Project Structure
- `cmd/`: Command-line interface definitions using Cobra.
- `pkg/`: Core business logic and service integrations.
    - `embedding/`: Ollama embedding service client.
    - `orchestrator/`: Main indexing and query orchestration.
    - `qdrant/`: Qdrant vector database client.
    - `util/`: Shared utility functions (file parsing, text processing).
- `main.go`: Application entry point.

### Prerequisites
- [Go](https://golang.org/) 1.25+
- [Ollama](https://ollama.com/) (running locally on port 11434 with `bge-m3` model)
- [Qdrant](https://qdrant.tech/) (running locally on port 6333)

### Usage

#### Install Dependencies
```bash
go mod tidy
```

#### Build the Application
```bash
go build -o qdrant-cli
```

#### Insert a Document
This will create a collection named after the file (e.g., `my_report`).
```bash
./qdrant-cli InsertFile --file path/to/my_report.pdf
```

#### Search

**Global Search (across all documents):**
```bash
./qdrant-cli search --query "your search text"
```

**Search within a specific document:**
Use the `-c` or `--collection` flag to specify the document collection.
```bash
./qdrant-cli search --query "your search text" --collection my_report
```

**Additional search options:**
- `--limit <N>`: Number of results to return (default: 5).
- `--score`: Show the similarity score for each result.

#### Ask with RAG (LLM)
Ask a question and get an answer based on your documents.
```bash
./qdrant-cli llm "How to update CKA exam?"
```

**Options:**
- `-m, --model <name>`: Specify Ollama model (default: `deepseek-r1:7b`).
- `-l, --limit <N>`: Number of context chunks to retrieve (default: 5).

---

<a name="chinese"></a>
## 中文

一个用于在 Qdrant 向量数据库中索引和搜索文档的命令行工具。支持从 `.md`、`.txt`、`.docx` 和 `.pdf` 文件中提取文本，通过 Ollama（BGE-M3 模型）生成嵌入向量，并将其存储在 Qdrant 中。

### 功能特性
- **文件导入**：支持 Markdown、纯文本、Word (Docx) 和 PDF 文档。
- **动态 Collection**：为每个新文档自动创建一个以文件名命名的 Qdrant Collection（例如，`我的文档.pdf` -> `我的文档` Collection）。
- **唯一文档 ID**：为每个索引的文档分配一个唯一的 UUID，确保其所有文本块都被关联起来。
- **嵌入生成**：使用 Ollama 和 `bge-m3` 模型生成高质量的文本向量。
- **高级向量搜索**：
    - **全局搜索**：同时在所有 Collection 中搜索，以找到最相关的结果。
    - **范围搜索**：可指定单个 Collection，以缩小搜索范围。
- **RAG 支持 (对话与 LLM)**：通过 Ollama 发起对话，并结合检索到的文档内容提供准确的回答。
- **清晰架构**：组织良好的项目结构，采用模块化包，清晰易维护。

### 项目结构
- `cmd/`：使用 Cobra 定义的命令行界面。
- `pkg/`：核心业务逻辑和服务集成。
    - `embedding/`：Ollama 嵌入服务客户端。
    - `orchestrator/`：主要的索引和查询编排逻辑。
    - `qdrant/`：Qdrant 向量数据库客户端。
    - `util/`：共享工具函数（文件解析、文本处理）。
- `main.go`：应用程序入口。

### 环境准备
- [Go](https://golang.org/) 1.25+
- [Ollama](https://ollama.com/) (运行在本地 11434 端口，已安装 `bge-m3` 模型)
- [Qdrant](https://qdrant.tech/) (运行在本地 6333 端口)

### 使用方法

#### 安装依赖
```bash
go mod tidy
```

#### 编译应用
```bash
go build -o qdrant-cli
```

#### 插入文档
该命令将创建一个以文件名命名的 Collection（例如 `my_report`）。
```bash
./qdrant-cli InsertFile --file path/to/my_report.pdf
```

#### 搜索

**全局搜索（跨所有文档）：**
```bash
./qdrant-cli search --query "你的搜索关键词"
```

**在指定文档内搜索：**
使用 `-c` 或 `--collection` 标志指定文档对应的 Collection。
```bash
./qdrant-cli search --query "你的搜索关键词" --collection my_report
```

**其他搜索选项：**
- `--limit <N>`：返回结果的数量（默认为 5）。
- `--score`：显示每个结果的相似度分数。

#### 使用 RAG 提问 (LLM)
基于你的文档进行提问。
```bash
./qdrant-cli llm "CKA 考试如何更新？"
```

**选项：**
- `-m, --model <name>`: 指定 Ollama 模型（默认：`deepseek-r1:7b`）。
- `-l, --limit <N>`: 检索的上下文块数量（默认：5）。
