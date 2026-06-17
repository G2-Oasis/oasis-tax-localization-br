# Migration guide — adicionar localização fiscal de um novo país

Este documento explica como criar uma nova biblioteca de localização fiscal seguindo o padrão estabelecido por `oasis-tax-localization-br`. O alvo é alguém que precisa habilitar o ecossistema Oásis pra um país distinto do Brasil (ex.: México, Argentina) com mínima fricção arquitetural.

> Use `oasis-tax-localization-br` como **referência canônica** — o repo de país novo deve espelhar a sua estrutura, ajustando apenas o conteúdo regulatório local.

## Quando criar uma lib de localização nova

Crie uma `oasis-tax-localization-<país>` quando o roadmap do produto contemplar emissão fiscal, cálculo tributário ou validações regulatórias num país que **não tem** lib publicada ainda. Sintomas concretos que justificam:

- Há roadmap de emissão de documento fiscal nesse país (CFDI no México, FCE na Argentina, etc.).
- Cálculo tributário local exige regras impossíveis de modelar via DSL fiscal genérica.
- Validação de identificação (RFC, CUIT, etc.) começou a ser duplicada em mais de um componente.

Se for "talvez no futuro", **não crie ainda**. ADR-001 documenta o trade-off de YAGNI vs separação prematura.

## Pré-requisitos

- Permissão de criar repos na org GitHub `G2-Oasis`.
- Permissão de adicionar dependência em `go.mod` dos repos consumidores (OTAX-001, OTAX-002, futuros).
- Definição mínima de produto: quais elementos fiscais (validators, datasets, integrações de governo) entram na primeira versão.

## Checklist de criação

### Passo 1 — Estrutura do repo novo

1. Crie `github.com/G2-Oasis/oasis-tax-localization-<XX>` (use o código ISO 3166-1 alpha-2 do país: `mx`, `ar`, `co`, etc.). **Repositório público** — facilita CI/CD cross-repo e a lib não tem dado sensível.
2. Branches: `main` (release) + `develop` (integração). PRs `feature/* → develop`. Tags semver saem de `main`.
3. Estrutura inicial:
   ```
   /reforma_ou_calculadora/   integração com calculadora fiscal oficial (se houver)
   /aliquotas/                datasets estruturados (ICMS, IVA, etc.)
   /validators/               validators de identificadores e códigos locais
   /docs/                     ADR + guides
   README.md                  descrição + roadmap
   go.mod                     module github.com/G2-Oasis/oasis-tax-localization-<XX>
   .gitignore                 padrão Go (alinhar com oasis-tax-localization-br)
   ```
4. **Não** copie validators BR mecanicamente — recrie por país com regras locais (DV de RFC ≠ DV de CNPJ; CFOP é exclusivo BR; etc.).

### Passo 2 — Workspace local

5. Clone o repo em `c:\dev\Oasis\oasis-tax-localization-<XX>\` (ou paridade no Linux).
6. Adicione no `c:\dev\Oasis\go.work`:
   ```go
   use (
       ./OTAX-001
       ./OTAX-002
       ./oasis-tax-localization-br
       ./oasis-tax-localization-<XX>    // adicionar aqui
       ...
   )
   ```
7. `go test ./...` no novo repo verde.

### Passo 3 — Conteúdo

8. Implementar primeiro lote de validators (identificadores locais — RFC/CUIT/etc., códigos fiscais locais).
9. Implementar datasets de alíquotas (se aplicável) com fonte oficial documentada.
10. Implementar cliente da calculadora oficial (se aplicável). Padrão: interface `Client` + `MockClient` + `RealClient` + `DisabledClient`. Ver `oasis-tax-localization-br/reforma/` como referência.
11. **Testes** — paridade com `oasis-tax-localization-br`: cobertura por validator, round-trip de tabelas, casos válidos e inválidos enumerados.

### Passo 4 — Release

12. Merge PRs em `develop` → fast-forward em `main` → tag `v0.1.0`.
13. Publica releases conforme features forem adicionadas. Semver: bump minor pra features, patch pra fix de regra fiscal.

### Passo 5 — Integração nos consumidores

14. Repos que precisarem do novo país (OTAX-001, OTAX-002, etc.):
    ```bash
    go get github.com/G2-Oasis/oasis-tax-localization-<XX>@v0.1.0
    ```
15. Decida estratégia de seleção de país no consumidor:
    - **Por tenant**: cada cliente tem um país configurado, runtime seleciona a lib.
    - **Por componente**: alguns componentes só atendem um país (lib hardcoded).
    - **Híbrida**: depende do componente.

   *Esta decisão não é coberta por este guide* — depende de evolução do produto. Em junho/2026, OTAX-001 e OTAX-002 só usam BR; quando aparecer 2º país, atualizar este documento.

### Passo 6 — ADR local

16. Crie `oasis-tax-localization-<XX>/docs/adr/ADR-001-localizacao-<país>.md` documentando:
    - Regulação fiscal de referência (LC equivalente, normativas).
    - Diferenças importantes vs. localização BR.
    - Fonte oficial dos datasets.

## O que **não** colocar na lib de localização

Pra manter a fronteira limpa com o motor T01 (OTAX-001):

- **Lógica de cálculo genérica** (DSL, agregação) — fica no motor.
- **Persistência** — a lib é stateless; quem grava é o componente consumidor.
- **Endpoints HTTP** — a lib é consumida via Go module, não via REST.
- **Roteamento de tenant** — o consumidor decide qual lib carregar.

## Versionamento + breaking changes

- **`v0.x.y`**: API instável, breaking changes permitidos com aviso.
- **`v1.0.0`**: API congelada. Mudanças incompatíveis exigem `v2.0.0` em path separado (`/v2/`).
- Patch de regra fiscal (ex: "ICMS-SP mudou de 18% pra 17%"): **bump patch** (`v0.3.1`). Não é breaking change.
- Reorganização de subpacotes: **bump minor** (`v0.4.0`). Apenas estrutura, não comportamento.

## Quando deletar uma lib de localização

Se o país sair do roadmap do produto, **arquive** o repo (não delete). Arquivar:

1. Não quebra `go.mod` de consumidores que ainda tenham dependência histórica.
2. Mantém histórico auditável.
3. Bloqueia novos PRs sem perder o passado.

Reuso futuro (caso o país volte ao roadmap) fica trivial: desarquivar.
