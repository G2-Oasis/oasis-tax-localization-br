# ADR-001 — Separação da localização fiscal em biblioteca própria

- **Status**: Aceita
- **Data**: 2026-06-16
- **Decisores**: Juliana Najara (Arquitetura), Silvyo Vieira (Backend)
- **Sprint**: T02 — Localização Fiscal Brasil

## Contexto

O motor fiscal do ecossistema Oásis (OTAX-001) nasceu monoBR. Toda regra, tabela e validação específica do regime tributário brasileiro vive misturada com o motor genérico de cálculo. Em concreto, três níveis de acoplamento foram observados em junho/2026:

1. **Estrutural**: `domain.TaxRule` tem campos BR hardcoded no struct (CFOP, CST, NCM, UFOrigem/Destino, CNAE) e na tabela SQL `tax_rule`.
2. **Operacional**: lookups, validações e mapeamentos BR ficam embutidos em `engine.go::rankRule/isRuleEligible`, `engine_reforma.go::buildReformaRequest` e similares.
3. **Espalhamento**: validators e tabelas auxiliares (UF→IBGE, CNPJ/CPF mod11) viviam **duplicados em 2–3 lugares** entre OTAX-001 e OTAX-002.

Quando o roadmap começou a contemplar **outros países** (sugerido inicialmente em conversa de 05/06 entre Juliana e o assistente do Tax), ficou claro que o T02 do diagrama arquitetural — "Localização Fiscal" — precisa existir como **componente real**, não diluído.

## Decisão

Criar a biblioteca **`github.com/G2-Oasis/oasis-tax-localization-br`** (módulo Go público) hospedando:

- **`reforma/`** — cliente HTTP da Calculadora oficial da Reforma Tributária do Consumo (CBS/IBS/IS, LC 214/2025).
- **`validators/`** — códigos fiscais BR: CPF, CNPJ (mod11), NCM (8d), CFOP (4d, prefixos {1,2,3,5,6,7}), CNAE (7d), UF↔IBGE (27 entradas).
- **`aliquotas/`** _(reservado)_ — datasets de alíquotas brasileiras (bloqueado em definição fiscal, ver BL-T02.04).

Os repos consumidores (OTAX-001 e OTAX-002) passam a importar a lib via `go.mod` em vez de duplicar tabelas/validators.

Quando outro país entrar no roadmap, segue o mesmo padrão: `oasis-tax-localization-mx`, `oasis-tax-localization-ar`, etc., conforme [MIGRATION_NOVO_PAIS.md](../MIGRATION_NOVO_PAIS.md).

## Consequências

### Positivas

- **Single source of truth** — tabela UF→IBGE existia em 3 cópias (OTAX-001 numbering, OTAX-001 nfcom_envelope consumindo, OTAX-002 domain). Idem CPF/CNPJ que viviam em OTAX-001 e seriam replicados pra OTAX-002 conforme demanda surgisse.
- **Reuso entre componentes** — OTAX-001, OTAX-002 (e futuros TMPO*, TMFC*) consomem a mesma lib. Mudança de regra fiscal BR (ex: nova UF, nova alíquota CBS) é atualizada em um lugar.
- **Caminho claro pra multi-país** — adicionar Brasil v2/outros países exige criar nova lib seguindo o template, sem mexer no motor genérico T01.
- **Testabilidade** — lib tem testes isolados. Repos consumidores testam só integração.

### Negativas

- **Versionamento adicional** — cada release da lib (tag semver) precisa ser propagada nos `go.mod` dos consumidores. Hoje: v0.1.0 → v0.2.0 → v0.3.0 em 1 dia, manageable.
- **Dependência cross-repo** — CI dos consumidores precisa de acesso à lib. Mitigação: lib é **pública**, evita necessidade de token compartilhado.
- **Latência de mudança em emergência** — se uma alíquota mudar amanhã, precisa fluxo: PR na lib → tag → bump na lib → PRs nos consumidores. Mitigação: branches/forks locais em emergência.

### Neutras

- **Fronteira ainda imperfeita** — `domain.TaxRule` em OTAX-001 segue com campos BR hardcoded. Refatorar `TaxRule` para suportar localização configurável é decisão futura (depende de pressão real de 2º país; ver "Alternativas").

## Alternativas consideradas

### A. Não fazer (YAGNI)

- **Razão**: OTAX-001 é monoBR, ainda em homologação. Pode-se argumentar que separação é prematura.
- **Por que rejeitada**: a duplicação já é dor concreta (3 cópias da tabela UF; CPF/CNPJ replicado seria 4ª). E a sugestão veio com motivação clara da arquitetura (Juliana), não é especulação técnica isolada.

### B. Sub-pacotes dentro do mesmo repo (internal/localization/br/)

- **Razão**: mais simples, sem novo módulo, sem novo CI.
- **Por que rejeitada**: não resolve duplicação cross-repo (OTAX-001 vs OTAX-002). E mantém localização "presa" ao OTAX-001, prejudicando reuso em outros componentes do ecossistema.

### C. Refactor profundo do `domain.TaxRule` agora

- **Razão**: faria sentido pra um sistema verdadeiramente multi-país desde o nascimento.
- **Por que rejeitada (parcialmente)**: custo estimado 3+ meses, alto risco em motor já validado em homologação NFCom. Adiada — vira tema próprio quando segundo país concreto entrar no roadmap, idealmente com requisitos reais em mãos pra não desenhar abstração errada.

## Referências

- Diagrama arquitetural do Tax (Juliana, 05/06/2026) — bloco T02 "Localização Fiscal (Brasil)" em alíquotas + UF + Reforma.
- Backlog Sprint T02 — BLs T02.01 a T02.07.
- Memory `project_oasis_arquitetura_revisao_19_05.md` — discussão paralela sobre fronteiras OTAX-001/OTAX-002.
