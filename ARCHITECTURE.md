# Arquitetura e Decisões do Projeto

Este documento detalha as estratégias arquiteturais e os padrões de projeto adotados para garantir que o sistema atenda aos requisitos rigorosos de consistência financeira, idempotência, alta concorrência e tolerância a falhas exigidos pelo domínio de apostas distribuídas.

## 1. Padrão Arquitetural: Clean Architecture & Uber Fx
A aplicação foi estruturada baseada nos princípios da **Clean Architecture**, promovendo o isolamento total das regras de negócio (`domain`) de detalhes de I/O, infraestrutura (`infrastructure`) e apresentação HTTP/Mensageria (`presentation`). O domínio não conhece banco de dados, SQS ou Echo.

Para gerenciar a composição de dependências e o ciclo de vida dos processos, adotou-se o **Uber Fx**. Através do `fx.Lifecycle`, garantimos *Graceful Shutdown*: o servidor HTTP web e os *workers* de background (SQS Consumer e Outbox Publisher) são iniciados e encerrados de maneira orquestrada, garantindo que conexões com o banco não sejam cortadas enquanto uma transação financeira está em andamento.

## 2. Modelagem de Dinheiro (Value Object)
Para cumprir o requisito estrito de evitar ponto flutuante (`float32`/`float64`), o dinheiro foi modelado como um **Value Object imutável** (`domain.Money`). 
*   **Representação:** O valor financeiro é trafegado e armazenado como `int64` (representando a menor unidade fracionária da moeda, como centavos).
*   **Segurança:** Toda a aritmética (soma e subtração) é encapsulada em métodos do domínio que validam *overflow* e checam se as moedas (ex: `BRL` vs `USD`) são idênticas antes de operar, evitando anomalias financeiras.

## 3. Controle de Concorrência e Transações Atômicas
O sistema utiliza o pacote `pgx` para interagir nativamente com o PostgreSQL. A coordenação de concorrência por carteira foi resolvida utilizando **Pessimistic Locking** (`SELECT ... FOR UPDATE`).
*   **Unit of Work:** Foi implementada uma interface no repositório que aceita *callbacks* (`RunInTransaction`). 
*   **Atomicidade Absoluta:** O saldo da carteira, a geração do recibo (`WagerTransaction`), o lançamento contábil (`WalletLedgerEntry`) e os eventos de mensageria (`Outbox`) são todos gravados sob o mesmo Lock transacional e o mesmo Contexto de Banco de Dados. Se um erro ocorrer em qualquer etapa, é feito o *Rollback* unificado. Duas requisições simultâneas para a mesma carteira enfileiram no banco, prevenindo o problema de *Lost Updates*.

## 4. Idempotência Persistente e Deduplicação (Inbox Pattern)
O sistema deve sobreviver a envios duplicados independentemente da via de entrada (HTTP ou SQS).
*   **No HTTP:** O cabeçalho `Idempotency-Key` é exigido. Antes de processar a aposta, o sistema busca um lock seguro e verifica se a chave já existe no repositório de transações. Caso exista, uma resposta de sucesso simulada (`idempotentReplay: true`) é devolvida imediatamente.
*   **No SQS (Inbox Pattern):** Para o processamento em background, o sistema extrai o `messageId` original entregue pelo SQS e grava em uma tabela SQL `inbox` atrelada à mesma transação do negócio (usando *Unique Constraints*). Se a AWS re-entregar a mensagem, o banco recusa a duplicação atamicamente.

## 5. Garantia de Eventos (Transactional Outbox)
Para assegurar a entrega *at-least-once* de eventos de negócio (`WalletBalanceChanged`, `WagerTransactionProcessed`) sem o risco de inconsistência (ex: atualizar o banco e o servidor desligar antes de publicar no broker), foi adotado o padrão **Transactional Outbox**.
Os eventos são serializados em JSON e salvos na tabela `outbox` no status de `PENDING` durante o *commit* bancário. Um processo secundário (`OutboxPublisher`) efetua o *polling* dessa tabela de forma assíncrona para despachar e sinalizar como publicado.

## 6. Auditoria Financeira e Reconciliação
Seguindo o padrão de contabilidade de partidas, cada operação que altera o saldo gera um registro imutável no `wallet_ledger` (Append-Only). Operações financeiras de perdas (`LOSS`), que não movimentam fundos reais, geram o recibo final sem criar lançamento contábil "vazio".
*   **Endpoint de Reconciliação:** O sistema possui uma rota (`/reconciliation`) que varre o ledger de uma carteira desde a sua origem matemática, somando créditos e subtraindo débitos por via do banco, retornando eventuais divergências contra a tabela `wallets` em tempo real.

## 7. Autenticação e Segurança
O serviço foi protegido utilizando a integração com provedores OAuth 2.0/OIDC (recomendado Keycloak). 
*   Um *Middleware* customizado foi injetado nas rotas do servidor web para interceptar requisições, exigir o *Bearer Token* e processar as assinaturas e *claims* do JWT (como extrair o `providerId` / `clientId`). Com isso, asseguramos o modelo restrito de permissões logo na camada de apresentação.

## 8. Limitações e Compromissos (Trade-offs)
*   **Reversões Assíncronas:** A lógica atual rejeita reversões de transações ainda não processadas. Em um cenário futuro e de escala máxima, uma reversão não identificada poderia ser salva como `PENDING_REFERENCE` no domínio para sofrer retentativas automáticas (*backoff* exponencial).
*   **Testes de Carga:** Os cenários de paralelismo máximo focaram na integridade e na blindagem de dados pelo PostgreSQL em ambiente local conteinerizado. Um teste de carga mais agressivo (via k6/Gatling) poderia evidenciar a necessidade de otimizar o polling da tabela Outbox (usando `SKIP LOCKED`, por exemplo).