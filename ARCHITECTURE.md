# Arquitetura e decisões do projeto

Este documento descreve o estado atual da implementação do desafio "Processamento Distribuído de Apostas em Go" e as decisões técnicas por trás dele. Ele reflete as abordagens de alta concorrência, consistência financeira e tolerância a falhas implementadas.

## 1. Visão Geral
O serviço segue os princípios de **Clean Architecture** (Ports & Adapters), garantindo a testabilidade e o isolamento das regras de negócio:
*   `internal/domain` — Entidades e Value Objects (Money, Wallet, WagerTransaction, LedgerEntry, Eventos). Totalmente livre de frameworks (Fx, HTTP, SQL).
*   `internal/application` — Casos de uso (UseCases) que orquestram o domínio através de interfaces implementadas pela infraestrutura.
*   `internal/infrastructure` — Adaptadores concretos: PostgreSQL (repositórios), SQS (consumers/publishers).
*   `internal/presentation` — Handlers HTTP (Echo) e Middlewares de segurança.
*   `cmd/api/main.go` — Composição de dependências gerenciada pelo **Uber Fx**.

O **Uber Fx** é utilizado via `fx.Lifecycle`, garantindo *Graceful Shutdown*: o servidor web e os *workers* SQS iniciam (`OnStart`) e encerram (`OnStop`) orquestradamente, não cortando conexões com o banco no meio de transações financeiras críticas.

## 2. Dinheiro (Value Object)
Para cumprir o requisito estrito de evitar ponto flutuante (`float32`/`float64`), o dinheiro foi modelado como um **Value Object imutável**.
*   **Representação:** Trafegado e armazenado internamente como `int64` (representando a menor unidade fracionária da moeda, ex: centavos).
*   **Segurança Matemática:** Aritmética isolada no domínio que valida inconsistências cambiais antes das operações. 
*   **Contrato HTTP:** O serviço faz o *parse* adequado das Strings de decimais exigidas pelo desafio (`{"amount":"25.00", "currency":"BRL"}`).

## 3. Agregado Wallet e Concorrência Otimizada
A coordenação de concorrência por carteira foi resolvida utilizando **Pessimistic Locking** (`SELECT ... FOR UPDATE`), orquestrado pelo padrão *Unit of Work*.
*   **Atomicidade Absoluta (`RunInTransaction`):** A leitura da carteira, a geração do recibo (`WagerTransaction`), o lançamento contábil no ledger (`WalletLedgerEntry`) e os eventos de outbox são todos executados dentro da **mesma transação SQL**.
*   Se ocorrer qualquer erro (ex: saldo insuficiente), o banco efetua o *Rollback* de tudo. Múltiplas requisições para a mesma carteira são enfileiradas pelo PostgreSQL, exterminando completamente anomalias de atualização perdida (*Lost Updates*).
*   Foi adotado o uso de `INSERT ... ON CONFLICT (id) DO UPDATE` para otimizar as atualizações no repósitório, salvando *round-trips* no banco.

## 4. WagerTransaction e Regras de Negócio
O sistema lida com diferentes *Kinds* orquestrados de forma segura:
*   **BET/WIN:** Debitam/Creditam o saldo, persistem no Ledger e emitem eventos.
*   **LOSS:** Validamos que o valor deve ser `0.00`. Ele não afeta o saldo, não emite linha no Ledger nem *WalletBalanceChanged*, mas emite o evento *WagerTransactionProcessed*.
*   **REFUND/ROLLBACK:** Operações de reversão que buscam a aposta originária via `FindWagerByExternalID` (usando `providerId` + `referenceExternalTransactionId`).

## 5. Inbox e Outbox (Garantia de Eventos e Mensageria)
O sistema sobrevive a envios duplicados, independente da via de entrada:
*   **Deduplicação HTTP:** Requisições via API trazem o `Idempotency-Key`. Se for interceptada uma chave duplicada, simulamos um replay (`idempotentReplay: true`) respondendo com o estado de sucesso.
*   **Transactional Outbox:** Eventos gerados pelo processamento financeiro (`WalletBalanceChanged`) são persistidos atamicamente na tabela `outbox`. Um *Worker* secundário (`OutboxPublisher`) efetua o polling seguro dessa tabela para despacho sem perda de eventos em crash da aplicação.
*   **Inbox Pattern (SQS Consumer):** O Worker do AWS SQS captura a mensagem da fila `wager-transactions.fifo` e extrai o `messageId`. Dentro da transação financeira, ele executa um `INSERT ... ON CONFLICT DO NOTHING` na tabela `inbox`. Se o SQS duplicar a entrega da mensagem, o banco rejeita a inserção silenciosamente sem abortar a transação do Postgres, retornando sucesso imediato sem causar efeitos colaterais financeiros.

## 6. Ledger Imutável e Reconciliação
Cada movimentação real de fundos grava um registro no `wallet_ledger` (Append-Only), contendo saldos anteriores e posteriores à transação.
*   **Reconciliação:** A rota `POST /wallets/:walletId/reconciliation` foi desenhada para recalcular iterativamente todo o histórico de transações (`SUM(CREDIT) - SUM(DEBIT)`) e atestar a veracidade do saldo cacheado na tabela `wallets`, emitindo um relatório analítico.

## 7. Autenticação e Autorização Segura (Keycloak / OIDC)
Segurança na camada de Middleware (`KeycloakAuthMiddleware`).
*   O sistema intercepta o JWT enviado via cabeçalho `Bearer`, decodifica as assinaturas e extrai nativamente *claims* de identidade (como o `providerId` / `clientId`).
*   Isso consolida as restrições arquiteturais para que parceiros manipulem apenas as carteiras cujas transações são autorizadas pelas políticas do broker.

## 8. Status por Área (Matriz de Requisitos)
| Requisito do Desafio | Componente/Solução | Status |
| :--- | :--- | :--- |
| **Money (Sem floats)** | Value Object, validações de formato e representação `int64`. | ✅ Concluído |
| **Pessimistic Locking** | `SELECT FOR UPDATE` isolado com injeção de transação no `context.Context`. | ✅ Concluído |
| **Idempotência (HTTP)** | Proteção por `Idempotency-Key` + Resposta com Replay. | ✅ Concluído |
| **Reversões (ROLLBACK/REFUND)** | Verificação e estorno validando o `ReferenceExternalTransactionID`. | ✅ Concluído |
| **Append-Only Ledger** | Lançamentos contábeis imutáveis (`wallet_ledger`). | ✅ Concluído |
| **Reconciliação** | Endpoint dedicado para auditoria analítica contra corrupção. | ✅ Concluído |
| **Mensageria (SQS Inbox)** | Deduplicação atômica usando `ON CONFLICT DO NOTHING` para `message_id`. | ✅ Concluído |
| **Transactional Outbox** | Escrita atômica + Worker assíncrono para publicação de eventos. | ✅ Concluído |
| **Autenticação (Keycloak)** | Middleware customizado com parse e extração de JWT Claims. | ✅ Concluído |

## 9. Limitações Conhecidas e Trabalhos Futuros
*   **Retentativas de Reversão (`PENDING_REFERENCE`):** Atualmente, se um `ROLLBACK` chegar *antes* da transação originária (out of order messages), ele é rejeitado. Uma melhoria seria modelá-lo como `PENDING_REFERENCE` e rodar um *Background Worker* com *backoff exponencial* que faz polling periódico aguardando a chegada da referência pendente até estourar um *TTL*.
*   **Desacoplamento de Fila SQS:** O `SQSConsumer` simula localmente a escuta, porém em ambiente Cloud real seria vital ajustar os perfis de concorrência (*MaxNumberOfMessages* e *WaitTimeSeconds* do AWS SDK v2).
