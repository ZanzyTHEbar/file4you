# Agentic RAG Flow

OpenAI's Agentic RAG Flow is a framework that combines retrieval-augmented generation (RAG) with agentic capabilities. This allows for the generation of responses based on both pre-existing knowledge and real-time information retrieval.

## Key Components

1. **Load the Entire Document** into the context window.
2. **Chunk the Document** into 20 chunk that respect sentence boundaries.
3. **Prompt the model** for which chunks might contain relevant information.
4. **Drill down** into the selected relevant chunks.
5. **Recursively call this function** until we reach paragraph-level content.
6. **Generate a response** based on the retrieved information.
7. **Verify the answer** for factual accuracy.

```mermaid
flowchart TD
    A[Start: CLI Invocation]
    B[Load Config and Environment Variables]
    C[Initialize LLM Provider Wrapper]
    D{Select LLM Provider}
    E[OpenAI Client]
    F[Anthropic Client]
    G[Local LLM Client]
    H[Unified API Abstraction]
    I[Extensibility and Plugin Manager]
    J[Advanced Prompt Engineering]
    K[Dynamic Prompt Templating and Chaining]
    L[Generate Custom Prompts]
    M[Send Prompt to LLM]
    N[Concurrent Processing - Goroutines and Channels]
    O[Custom Agent Module]
    P[Inter-Agent Messaging - Message Bus / Actor Model]
    Q[Mixture of Agents - Ensemble and Iterative Refinement]
    R[RAG Workflow Integration - Vector Retrieval and Context Augmentation]
    S[LLM Final Response Generation]
    T[Structured Output - JSON Schema Validation]
    U[Observability and Metrics Logging]
    V[Self-Optimization and Adaptive Tuning]
    W[Robust Error Handling and Retry Logic]
    X[Return Final Response to CLI]

    A --> B
    B --> C
    C --> D
    D -->|OpenAI| E
    D -->|Anthropic| F
    D -->|Local| G
    E --> H
    F --> H
    G --> H
    H --> I
    I --> J
    J --> K
    K --> L
    L --> M
    M --> N
    N --> O
    O --> P
    P --> Q
    Q --> R
    R --> S
    S --> T
    T --> U
    U --> V
    V --> W
    W --> X
```