# ctx — anti-tokens CLI para Claude Code

> Analise seu projeto localmente. Envie ao Claude resumos densos, não arquivos brutos.

**Status:** alpha — em desenvolvimento ativo. Interfaces podem mudar.

## Por que usar

Cada token que o Claude lê tem custo e consome janela de contexto. Ao trabalhar em projetos grandes, o Claude frequentemente gasta milhares de tokens apenas lendo arquivos para construir um modelo mental antes de fazer qualquer trabalho real.

`ctx` roda localmente, analisa seu projeto com ferramentas especializadas por linguagem (Roslyn para C#, tree-sitter para TypeScript), e emite resumos compactos em markdown que dão ao Claude tudo que ele precisa em uma fração dos tokens.

### Benchmark

Teste real: análise de "o que preciso mudar para adicionar uma feature" em projeto em produção.
Ambas as sessões iniciadas com `/clear` — condições equivalentes.

| Teste | Consumo de janela de contexto |
|-------|-------------------------------|
| Sem ctx | +7% |
| Com ctx | +3% |

**Economia: 57%.** Em sessões longas de desenvolvimento, onde o Claude faz a mesma exploração várias vezes ao longo da conversa, essa economia acumula.

## Instalação

**1. Instale o Go**
Baixe em https://go.dev/dl/ e siga o instalador. Requer Go 1.22+.

**2. Instale o ctx**
```sh
go install github.com/ricarneiro/ctx/cmd/ctx@latest
```

**3. Verifique**
```sh
ctx --version
```

## Uso

```sh
# Contexto git: commits recentes, status, info da branch
ctx git

# Detectar stack e emitir visão geral do projeto
ctx auto project

# Estrutura de projeto C# (requer .NET SDK)
ctx csharp project

# Outline de arquivo C#: tipos, métodos, assinaturas
ctx csharp outline src/MyService.cs

# Listar erros de compilação
ctx csharp errors
```

Toda saída é markdown UTF-8 no stdout. Redirecione onde precisar:

```sh
ctx csharp project | clip     # Windows
ctx csharp project | pbcopy   # macOS
```

Ou referencie em um `CLAUDE.md`:

```markdown
Run `ctx csharp project` to get the project overview before making changes.
```

## Contribuindo

Contribuições são bem-vindas. Clone o repositório, faça suas alterações e abra um Pull Request. Para mudanças grandes, abra uma issue primeiro para alinhar o escopo.

```sh
git clone https://github.com/ricarneiro/CTX.git
cd CTX
go build -o ctx.exe ./cmd/ctx
```

## Licença

MIT — veja [LICENSE](LICENSE).

---

# ctx — anti-tokens CLI for Claude Code

> Analyze your codebase locally. Feed Claude dense summaries, not raw files.

**Status:** alpha — under active development. Interfaces will change.

## Why

Every token Claude reads costs money and burns context window. When working on a large codebase, Claude often spends thousands of tokens just reading files to build a mental model before doing any actual work.

`ctx` runs locally, analyzes your project with language-aware tools (Roslyn for C#, tree-sitter for TypeScript), and emits compact markdown summaries that give Claude everything it needs in a fraction of the tokens.

### Benchmark

Real test: analyzing "what needs to change to add a feature" on a production project.
Both sessions started with `/clear` — equivalent conditions.

| Test | Context window usage |
|------|----------------------|
| Without ctx | +7% |
| With ctx | +3% |

**Savings: 57%.** In long development sessions, where Claude performs the same exploration repeatedly throughout a conversation, the savings compound.

## Installation

**1. Install Go**
Download from https://go.dev/dl/ and run the installer. Requires Go 1.22+.

**2. Install ctx**
```sh
go install github.com/ricarneiro/ctx/cmd/ctx@latest
```

**3. Verify**
```sh
ctx --version
```

## Usage

```sh
# Git context: recent commits, status, branch info
ctx git

# Auto-detect stack and emit project overview
ctx auto project

# C# project structure (requires .NET SDK)
ctx csharp project

# C# file outline: types, methods, signatures
ctx csharp outline src/MyService.cs

# List compilation errors
ctx csharp errors
```

All output is UTF-8 markdown on stdout. Pipe it where you need it:

```sh
ctx csharp project | pbcopy   # macOS
ctx csharp project | clip     # Windows
```

Or reference it in a `CLAUDE.md`:

```markdown
Run `ctx csharp project` to get the project overview before making changes.
```

## Contributing

Contributions are welcome. Clone the repository, make your changes, and open a Pull Request. For large changes, open an issue first to align on scope.

```sh
git clone https://github.com/ricarneiro/CTX.git
cd CTX
go build -o ctx.exe ./cmd/ctx
```

## License

MIT — see [LICENSE](LICENSE).
